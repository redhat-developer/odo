package kubedev

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/v2/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/v2/pkg/util"
	"golang.org/x/sync/errgroup"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/dev/kubedev/storage"
	"github.com/redhat-developer/odo/pkg/dev/kubedev/utils"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/port"
	"github.com/redhat-developer/odo/pkg/service"
	storagepkg "github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/watch"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
// The componentStatus will be modified to reflect the status of the component when the function returns
func (o *DevClient) Push(ctx context.Context, parameters common.PushParameters, componentStatus *watch.ComponentStatus) (err error) {

	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
	)

	// preliminary checks
	err = dfutil.ValidateK8sResourceName("component name", componentName)
	if err != nil {
		return err
	}

	err = dfutil.ValidateK8sResourceName("component namespace", o.kubernetesClient.GetCurrentNamespace())
	if err != nil {
		return err
	}

	if componentStatus.State == watch.StateSyncOutdated {
		// Clear the cache of image components already applied, hence forcing image components to be reapplied.
		componentStatus.ImageComponentsAutoApplied = make(map[string]devfilev1.ImageComponent)
	}

	klog.V(4).Infof("component state: %q\n", componentStatus.State)
	err = o.buildPushAutoImageComponents(ctx, o.filesystem, parameters.Devfile, componentStatus)
	if err != nil {
		return err
	}

	deployment, deploymentExists, err := o.getComponentDeployment(ctx)
	if err != nil {
		return err
	}

	if componentStatus.State != watch.StateWaitDeployment && componentStatus.State != watch.StateReady {
		log.SpinnerNoSpin("Waiting for Kubernetes resources")
	}

	// Set the mode to Dev since we are using "odo dev" here
	runtime := component.GetComponentRuntimeFromDevfileMetadata(parameters.Devfile.Data.GetMetadata())
	labels := odolabels.GetLabels(componentName, appName, runtime, odolabels.ComponentDevMode, false)

	var updated bool
	deployment, updated, err = o.createOrUpdateComponent(ctx, parameters, deploymentExists, libdevfile.DevfileCommands{
		BuildCmd: parameters.DevfileBuildCmd,
		RunCmd:   parameters.DevfileRunCmd,
		DebugCmd: parameters.DevfileDebugCmd,
	}, deployment)
	if err != nil {
		return fmt.Errorf("unable to create or update component: %w", err)
	}
	ownerReference := generator.GetOwnerReference(deployment)

	// Delete remote resources that are not present in the Devfile
	selector := odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode, false)

	objectsToRemove, serviceBindingSecretsToRemove, err := o.getRemoteResourcesNotPresentInDevfile(ctx, parameters, selector)
	if err != nil {
		return fmt.Errorf("unable to determine resources to delete: %w", err)
	}

	err = o.deleteRemoteResources(objectsToRemove)
	if err != nil {
		return fmt.Errorf("unable to delete remote resources: %w", err)
	}

	// this is mainly useful when the Service Binding Operator is not installed;
	// and the service binding secrets must be deleted manually since they are created by odo
	if len(serviceBindingSecretsToRemove) != 0 {
		err = o.deleteServiceBindingSecrets(serviceBindingSecretsToRemove, deployment)
		if err != nil {
			return fmt.Errorf("unable to delete service binding secrets: %w", err)
		}
	}

	// Create all the K8s components defined in the devfile
	_, err = o.pushDevfileKubernetesComponents(ctx, parameters, labels, odolabels.ComponentDevMode, ownerReference)
	if err != nil {
		return err
	}

	err = o.updatePVCsOwnerReferences(ctx, ownerReference)
	if err != nil {
		return err
	}

	if updated {
		klog.V(4).Infof("Deployment has been updated to generation %d. Waiting new event...\n", deployment.GetGeneration())
		componentStatus.State = watch.StateWaitDeployment
		return nil
	}

	numberReplicas := deployment.Status.ReadyReplicas
	if numberReplicas != 1 {
		klog.V(4).Infof("Deployment has %d ready replicas. Waiting new event...\n", numberReplicas)
		componentStatus.State = watch.StateWaitDeployment
		return nil
	}

	injected, err := o.bindingClient.CheckServiceBindingsInjectionDone(componentName, appName)
	if err != nil {
		return err
	}

	if !injected {
		klog.V(4).Infof("Waiting for all service bindings to be injected...\n")
		return errors.New("some servicebindings are not injected")
	}

	// Check if endpoints changed in Devfile
	portsToForward, err := libdevfile.GetDevfileContainerEndpointMapping(parameters.Devfile, parameters.Debug)
	if err != nil {
		return err
	}
	portsChanged := !reflect.DeepEqual(portsToForward, o.portForwardClient.GetForwardedPorts())

	if componentStatus.State == watch.StateReady && !portsChanged {
		// If the deployment is already in Ready State, no need to continue
		return nil
	}

	// Now the Deployment has a Ready replica, we can get the Pod to work inside it
	pod, err := o.kubernetesClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", componentName, err)
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, syncFolder, err := common.GetFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", pod.GetName(), err)
	}

	s := log.Spinner("Syncing files into the container")
	defer s.End(false)

	// Get commands
	pushDevfileCommands, err := o.getPushDevfileCommands(parameters)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	podChanged := componentStatus.State == watch.StateWaitDeployment

	// Get a sync adapter. Check if project files have changed and sync accordingly
	compInfo := sync.ComponentInfo{
		ComponentName: componentName,
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}

	cmdKind := devfilev1.RunCommandGroupKind
	cmdName := parameters.DevfileRunCmd
	if parameters.Debug {
		cmdKind = devfilev1.DebugCommandGroupKind
		cmdName = parameters.DevfileDebugCmd
	}

	syncParams := sync.SyncParameters{
		Path:                     path,
		WatchFiles:               parameters.WatchFiles,
		WatchDeletedFiles:        parameters.WatchDeletedFiles,
		IgnoredFiles:             parameters.IgnoredFiles,
		DevfileScanIndexForWatch: parameters.DevfileScanIndexForWatch,

		CompInfo:  compInfo,
		ForcePush: !deploymentExists || podChanged,
		Files:     common.GetSyncFilesFromAttributes(pushDevfileCommands[cmdKind]),
	}

	execRequired, err := o.syncClient.SyncFiles(ctx, syncParams)
	if err != nil {
		componentStatus.State = watch.StateReady
		return fmt.Errorf("failed to sync to component with name %s: %w", componentName, err)
	}
	s.End(true)

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if !componentStatus.PostStartEventsDone && libdevfile.HasPostStartEvents(parameters.Devfile) {
		err = libdevfile.ExecPostStartEvents(ctx, parameters.Devfile, component.NewExecHandler(o.kubernetesClient, o.execClient, appName, componentName, pod.Name, "Executing post-start command in container", parameters.Show))
		if err != nil {
			return err
		}
	}
	componentStatus.PostStartEventsDone = true

	cmd, err := libdevfile.ValidateAndGetCommand(parameters.Devfile, cmdName, cmdKind)
	if err != nil {
		return err
	}

	commandType, err := parsercommon.GetCommandType(cmd)
	if err != nil {
		return err
	}
	var running bool
	var isComposite bool
	cmdHandler := runHandler{
		fs:            o.filesystem,
		execClient:    o.execClient,
		kubeClient:    o.kubernetesClient,
		appName:       appName,
		componentName: componentName,
		devfile:       parameters.Devfile,
		path:          path,
		podName:       pod.GetName(),
		ctx:           ctx,
	}

	if commandType == devfilev1.ExecCommandType {
		running, err = cmdHandler.IsRemoteProcessForCommandRunning(ctx, cmd, pod.Name)
		if err != nil {
			return err
		}
	} else if commandType == devfilev1.CompositeCommandType {
		// this handler will run each command in this composite command individually,
		// and will determine whether each command is running or not.
		isComposite = true
	} else {
		return fmt.Errorf("unsupported type %q for Devfile command %s, only exec and composite are handled",
			commandType, cmd.Id)
	}

	cmdHandler.componentExists = running || isComposite

	klog.V(4).Infof("running=%v, execRequired=%v",
		running, execRequired)

	if isComposite || !running || execRequired {
		// Invoke the build command once (before calling libdevfile.ExecuteCommandByNameAndKind), as, if cmd is a composite command,
		// the handler we pass will be called for each command in that composite command.
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(o.kubernetesClient, o.execClient, appName, componentName, pod.Name,
				"Building your application in container", parameters.Show)
			return libdevfile.Build(ctx, parameters.Devfile, parameters.DevfileBuildCmd, execHandler)
		}
		if running {
			if cmd.Exec == nil || !util.SafeGetBool(cmd.Exec.HotReloadCapable) {
				if err = doExecuteBuildCommand(); err != nil {
					return err
				}
			}
		} else {
			if err = doExecuteBuildCommand(); err != nil {
				return err
			}
		}
		err = libdevfile.ExecuteCommandByNameAndKind(ctx, parameters.Devfile, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
	}

	if podChanged || portsChanged {
		o.portForwardClient.StopPortForwarding(ctx, componentName)
	}

	// Check that the application is actually listening on the ports declared in the Devfile, so we are sure that port-forwarding will work
	appReadySpinner := log.Spinner("Waiting for the application to be ready")
	err = o.checkAppPorts(ctx, pod.Name, portsToForward)
	appReadySpinner.End(err == nil)
	if err != nil {
		log.Warningf("Port forwarding might not work correctly: %v", err)
		log.Warning("Running `odo logs --follow` might help in identifying the problem.")
		fmt.Fprintln(log.GetStdout())
	}

	err = o.portForwardClient.StartPortForwarding(ctx, parameters.Devfile, componentName, parameters.Debug, parameters.RandomPorts, log.GetStdout(), parameters.ErrOut, parameters.CustomForwardedPorts)
	if err != nil {
		return common.NewErrPortForward(err)
	}
	componentStatus.EndpointsForwarded = o.portForwardClient.GetForwardedPorts()

	componentStatus.State = watch.StateReady
	return nil
}

func (o *DevClient) buildPushAutoImageComponents(ctx context.Context, fs filesystem.Filesystem, devfileObj parser.DevfileObj, compStatus *watch.ComponentStatus) error {
	components, err := libdevfile.GetImageComponentsToPushAutomatically(devfileObj)
	if err != nil {
		return err
	}

	for _, c := range components {
		if c.Image == nil {
			return fmt.Errorf("component %q should be an Image Component", c.Name)
		}
		alreadyApplied, ok := compStatus.ImageComponentsAutoApplied[c.Name]
		if ok && reflect.DeepEqual(*c.Image, alreadyApplied) {
			klog.V(1).Infof("Skipping image component %q; already applied and not changed", c.Name)
			continue
		}
		err = image.BuildPushSpecificImage(ctx, fs, c, true)
		if err != nil {
			return err
		}
		compStatus.ImageComponentsAutoApplied[c.Name] = *c.Image
	}

	// Remove keys that might no longer be valid
	devfileHasCompFn := func(n string) bool {
		for _, c := range components {
			if c.Name == n {
				return true
			}
		}
		return false
	}
	for n := range compStatus.ImageComponentsAutoApplied {
		if !devfileHasCompFn(n) {
			delete(compStatus.ImageComponentsAutoApplied, n)
		}
	}

	return nil
}

// getComponentDeployment returns the deployment associated with the component, if deployed
// and indicate if the deployment has been found
func (o *DevClient) getComponentDeployment(ctx context.Context) (*appsv1.Deployment, bool, error) {
	var (
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)
	)

	// Get the Dev deployment:
	// Since `odo deploy` can theoretically deploy a deployment as well with the same instance name
	// we make sure that we are retrieving the deployment with the Dev mode, NOT Deploy.
	selectorLabels := odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode, true)
	deployment, err := o.kubernetesClient.GetOneDeploymentFromSelector(selectorLabels)

	if err != nil {
		if _, ok := err.(*kclient.DeploymentNotFoundError); !ok {
			return nil, false, fmt.Errorf("unable to determine if component %s exists: %w", componentName, err)
		}
	}
	componentExists := deployment != nil
	return deployment, componentExists, nil
}

// createOrUpdateComponent creates the deployment or updates it if it already exists
// with the expected spec.
// Returns the new deployment and if the generation of the deployment has been updated
func (o *DevClient) createOrUpdateComponent(
	ctx context.Context,
	parameters common.PushParameters,
	componentExists bool,
	commands libdevfile.DevfileCommands,
	deployment *appsv1.Deployment,
) (*appsv1.Deployment, bool, error) {

	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
	)

	runtime := component.GetComponentRuntimeFromDevfileMetadata(parameters.Devfile.Data.GetMetadata())

	// Set the labels
	labels := odolabels.GetLabels(componentName, appName, runtime, odolabels.ComponentDevMode, true)

	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(parameters.Devfile.Data.GetMetadata()))
	odolabels.AddCommonAnnotations(annotations)
	klog.V(4).Infof("We are deploying these annotations: %s", annotations)

	deploymentObjectMeta, err := o.generateDeploymentObjectMeta(ctx, deployment, labels, annotations)
	if err != nil {
		return nil, false, err
	}

	policy, err := o.kubernetesClient.GetCurrentNamespacePolicy()
	if err != nil {
		return nil, false, err
	}
	podTemplateSpec, err := generator.GetPodTemplateSpec(parameters.Devfile, generator.PodTemplateParams{
		ObjectMeta:                 deploymentObjectMeta,
		PodSecurityAdmissionPolicy: policy,
	})
	if err != nil {
		return nil, false, err
	}
	containers := podTemplateSpec.Spec.Containers
	if len(containers) == 0 {
		return nil, false, fmt.Errorf("no valid components found in the devfile")
	}

	initContainers := podTemplateSpec.Spec.InitContainers

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(parameters.Devfile, containers, commands.BuildCmd, commands.RunCmd, commands.DebugCmd)
	if err != nil {
		return nil, false, err
	}

	// Returns the volumes to add to the PodTemplate and adds volumeMounts to the containers and initContainers
	volumes, err := o.buildVolumes(ctx, parameters, containers, initContainers)
	if err != nil {
		return nil, false, err
	}
	podTemplateSpec.Spec.Volumes = volumes

	selectorLabels := map[string]string{
		"component": componentName,
	}

	deployParams := generator.DeploymentParams{
		TypeMeta:          generator.GetTypeMeta(kclient.DeploymentKind, kclient.DeploymentAPIVersion),
		ObjectMeta:        deploymentObjectMeta,
		PodTemplateSpec:   podTemplateSpec,
		PodSelectorLabels: selectorLabels,
		Replicas:          pointer.Int32(1),
	}

	// Save generation to check if deployment is updated later
	var originalGeneration int64 = 0
	if deployment != nil {
		originalGeneration = deployment.GetGeneration()
	}

	deployment, err = generator.GetDeployment(parameters.Devfile, deployParams)
	if err != nil {
		return nil, false, err
	}
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	if vcsUri := util.GetGitOriginPath(path); vcsUri != "" {
		deployment.Annotations["app.openshift.io/vcs-uri"] = vcsUri
	}

	// add the annotations to the service for linking
	serviceAnnotations := make(map[string]string)
	serviceAnnotations["service.binding/backend_ip"] = "path={.spec.clusterIP}"
	serviceAnnotations["service.binding/backend_port"] = "path={.spec.ports},elementType=sliceOfMaps,sourceKey=name,sourceValue=port"

	serviceName, err := util.NamespaceKubernetesObjectWithTrim(componentName, appName, 63)
	if err != nil {
		return nil, false, err
	}
	serviceObjectMeta := generator.GetObjectMeta(serviceName, o.kubernetesClient.GetCurrentNamespace(), labels, serviceAnnotations)
	serviceParams := generator.ServiceParams{
		ObjectMeta:     serviceObjectMeta,
		SelectorLabels: selectorLabels,
	}
	svc, err := generator.GetService(parameters.Devfile, serviceParams, parsercommon.DevfileOptions{})

	if err != nil {
		return nil, false, err
	}
	klog.V(2).Infof("Creating deployment %v", deployment.Spec.Template.GetName())
	klog.V(2).Infof("The component name is %v", componentName)
	if componentExists {
		// If the component already exists, get the resource version of the deploy before updating
		klog.V(2).Info("The component already exists, attempting to update it")
		if o.kubernetesClient.IsSSASupported() {
			klog.V(4).Info("Applying deployment")
			deployment, err = o.kubernetesClient.ApplyDeployment(*deployment)
		} else {
			klog.V(4).Info("Updating deployment")
			deployment, err = o.kubernetesClient.UpdateDeployment(*deployment)
		}
		if err != nil {
			return nil, false, err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		ownerReference := generator.GetOwnerReference(deployment)
		err = o.createOrUpdateServiceForComponent(ctx, svc, ownerReference)
		if err != nil {
			return nil, false, err
		}
	} else {
		if o.kubernetesClient.IsSSASupported() {
			deployment, err = o.kubernetesClient.ApplyDeployment(*deployment)
		} else {
			deployment, err = o.kubernetesClient.CreateDeployment(*deployment)
		}

		if err != nil {
			return nil, false, err
		}

		klog.V(2).Infof("Successfully created component %v", componentName)
		if len(svc.Spec.Ports) > 0 {
			ownerReference := generator.GetOwnerReference(deployment)
			originOwnerRefs := svc.OwnerReferences
			err = o.kubernetesClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerRefs, ownerRef)
				_, err = o.kubernetesClient.CreateService(*svc)
				return err
			})
			if err != nil {
				return nil, false, err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}
	newGeneration := deployment.GetGeneration()

	return deployment, newGeneration != originalGeneration, nil
}

// getRemoteResourcesNotPresentInDevfile compares the list of Devfile K8s component and remote K8s resources
// and returns a list of the remote resources not present in the Devfile and in case the SBO is not installed, a list of service binding secrets that must be deleted;
// it ignores the core components (such as deployments, svc, pods; all resources with `component:<something>` label)
func (o DevClient) getRemoteResourcesNotPresentInDevfile(ctx context.Context, parameters common.PushParameters, selector string) (objectsToRemove, serviceBindingSecretsToRemove []unstructured.Unstructured, err error) {
	var (
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
	)

	currentNamespace := o.kubernetesClient.GetCurrentNamespace()
	allRemoteK8sResources, err := o.kubernetesClient.GetAllResourcesFromSelector(selector, currentNamespace)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to fetch remote resources: %w", err)
	}

	var remoteK8sResources []unstructured.Unstructured
	// Filter core components
	for _, remoteK := range allRemoteK8sResources {
		if !odolabels.IsCoreComponent(remoteK.GetLabels()) {
			// ignore the resources that are already set for deletion
			// ignore the resources that do not have projecttype annotation set; they will be the resources that are not created by odo
			// for e.g. PodMetrics is a resource that is created if Monitoring is enabled on OCP;
			// this resource has the same label as it's deployment, it has no owner reference; but it does not have the annotation either
			if remoteK.GetDeletionTimestamp() != nil && !odolabels.IsProjectTypeSetInAnnotations(remoteK.GetAnnotations()) {
				continue
			}
			remoteK8sResources = append(remoteK8sResources, remoteK)
		}
	}

	var devfileK8sResources []devfilev1.Component
	devfileK8sResources, err = libdevfile.GetK8sAndOcComponentsToPush(parameters.Devfile, true)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to obtain resources from the Devfile: %w", err)
	}

	// convert all devfileK8sResources to unstructured data
	var devfileK8sResourcesUnstructured []unstructured.Unstructured
	for _, devfileK := range devfileK8sResources {
		var devfileKUnstructuredList []unstructured.Unstructured
		devfileKUnstructuredList, err = libdevfile.GetK8sComponentAsUnstructuredList(parameters.Devfile, devfileK.Name, path, devfilefs.DefaultFs{})
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read the resource: %w", err)
		}
		devfileK8sResourcesUnstructured = append(devfileK8sResourcesUnstructured, devfileKUnstructuredList...)
	}

	isSBOSupported, err := o.kubernetesClient.IsServiceBindingSupported()
	if err != nil {
		return nil, nil, fmt.Errorf("error in determining support for the Service Binding Operator: %w", err)
	}

	// check if the remote resource is also present in the Devfile
	for _, remoteK := range remoteK8sResources {
		matchFound := false
		isServiceBindingSecret := false
		for _, devfileK := range devfileK8sResourcesUnstructured {
			// only check against GroupKind because version might not always match
			if remoteResourceIsPresentInDevfile := devfileK.GroupVersionKind().GroupKind() == remoteK.GroupVersionKind().GroupKind() &&
				devfileK.GetName() == remoteK.GetName(); remoteResourceIsPresentInDevfile {
				matchFound = true
				break
			}

			// if the resource is a secret and the SBO is not installed, then check if it's related to a local ServiceBinding by checking the labels
			if !isSBOSupported && remoteK.GroupVersionKind() == kclient.SecretGVK {
				if remoteSecretHasLocalServiceBindingOwner := service.IsLinkSecret(remoteK.GetLabels()) &&
					remoteK.GetLabels()[service.LinkLabel] == devfileK.GetName(); remoteSecretHasLocalServiceBindingOwner {
					matchFound = true
					isServiceBindingSecret = true
					break
				}
			}
		}

		if !matchFound {
			if isServiceBindingSecret {
				serviceBindingSecretsToRemove = append(serviceBindingSecretsToRemove, remoteK)
			} else {
				objectsToRemove = append(objectsToRemove, remoteK)
			}
		}
	}
	return objectsToRemove, serviceBindingSecretsToRemove, nil
}

// deleteRemoteResources takes a list of remote resources to be deleted
func (o DevClient) deleteRemoteResources(objectsToRemove []unstructured.Unstructured) error {
	if len(objectsToRemove) == 0 {
		return nil
	}

	var resources []string
	for _, u := range objectsToRemove {
		resources = append(resources, fmt.Sprintf("%s/%s", u.GetKind(), u.GetName()))
	}

	// Delete the resources present on the cluster but not in the Devfile
	klog.V(3).Infof("Deleting %d resource(s) not present in the Devfile: %s", len(resources), strings.Join(resources, ", "))
	g := new(errgroup.Group)
	for _, objectToRemove := range objectsToRemove {
		// Avoid re-use of the same `objectToRemove` value in each goroutine closure.
		// See https://golang.org/doc/faq#closures_and_goroutines for more details.
		objectToRemove := objectToRemove
		g.Go(func() error {
			gvr, err := o.kubernetesClient.GetGVRFromGVK(objectToRemove.GroupVersionKind())
			if err != nil {
				return fmt.Errorf("unable to get information about resource: %s/%s: %w", objectToRemove.GetKind(), objectToRemove.GetName(), err)
			}

			err = o.kubernetesClient.DeleteDynamicResource(objectToRemove.GetName(), gvr, true)
			if err != nil {
				if !(kerrors.IsNotFound(err) || kerrors.IsMethodNotSupported(err)) {
					return fmt.Errorf("unable to delete resource: %s/%s: %w", objectToRemove.GetKind(), objectToRemove.GetName(), err)
				}
				klog.V(3).Infof("Failed to delete resource: %s/%s; resource not found or method not supported", objectToRemove.GetKind(), objectToRemove.GetName())
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

// deleteServiceBindingSecrets takes a list of Service Binding secrets(unstructured) that should be deleted;
// this is helpful when Service Binding Operator is not installed on the cluster
func (o DevClient) deleteServiceBindingSecrets(serviceBindingSecretsToRemove []unstructured.Unstructured, deployment *appsv1.Deployment) error {
	for _, secretToRemove := range serviceBindingSecretsToRemove {
		spinner := log.Spinnerf("Deleting Kubernetes resource: %s/%s", secretToRemove.GetKind(), secretToRemove.GetName())
		defer spinner.End(false)

		err := service.UnbindWithLibrary(o.kubernetesClient, secretToRemove, deployment)
		if err != nil {
			return fmt.Errorf("failed to unbind secret %q from the application", secretToRemove.GetName())
		}

		// since the library currently doesn't delete the secret after unbinding
		// delete the secret manually
		err = o.kubernetesClient.DeleteSecret(secretToRemove.GetName(), o.kubernetesClient.GetCurrentNamespace())
		if err != nil {
			if !kerrors.IsNotFound(err) {
				return fmt.Errorf("unable to delete Kubernetes resource: %s/%s: %s", secretToRemove.GetKind(), secretToRemove.GetName(), err.Error())
			}
			klog.V(4).Infof("Failed to delete Kubernetes resource: %s/%s; resource not found", secretToRemove.GetKind(), secretToRemove.GetName())
		}
		spinner.End(true)
	}
	return nil
}

// pushDevfileKubernetesComponents gets the Kubernetes components from the Devfile and push them to the cluster
// adding the specified labels and ownerreference to them
func (o *DevClient) pushDevfileKubernetesComponents(
	ctx context.Context,
	parameters common.PushParameters,
	labels map[string]string,
	mode string,
	reference metav1.OwnerReference,
) ([]devfilev1.Component, error) {
	var (
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
	)

	// fetch the "kubernetes inlined components" to create them on cluster
	// from odo standpoint, these components contain yaml manifest of ServiceBinding
	k8sComponents, err := libdevfile.GetK8sAndOcComponentsToPush(parameters.Devfile, false)
	if err != nil {
		return nil, fmt.Errorf("error while trying to fetch service(s) from devfile: %w", err)
	}

	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	err = component.ValidateResourcesExist(o.kubernetesClient, parameters.Devfile, k8sComponents, path)
	if err != nil {
		return nil, err
	}

	// Set the annotations for the component type
	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(parameters.Devfile.Data.GetMetadata()))

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	err = service.PushKubernetesResources(o.kubernetesClient, parameters.Devfile, k8sComponents, labels, annotations, path, mode, reference)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes resources associated with the component: %w", err)
	}
	return k8sComponents, nil
}

func (o *DevClient) updatePVCsOwnerReferences(ctx context.Context, ownerReference metav1.OwnerReference) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
	)

	// list the latest state of the PVCs
	pvcs, err := o.kubernetesClient.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
	if err != nil {
		return err
	}

	// update the owner reference of the PVCs with the deployment
	for i := range pvcs {
		if pvcs[i].OwnerReferences != nil || pvcs[i].DeletionTimestamp != nil {
			continue
		}
		err = o.kubernetesClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
			return o.kubernetesClient.UpdateStorageOwnerReference(&pvcs[i], ownerRef)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *DevClient) getPushDevfileCommands(parameters common.PushParameters) (map[devfilev1.CommandGroupKind]devfilev1.Command, error) {
	pushDevfileCommands, err := libdevfile.ValidateAndGetPushCommands(parameters.Devfile, parameters.DevfileBuildCmd, parameters.DevfileRunCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	if parameters.Debug {
		pushDevfileDebugCommands, e := libdevfile.ValidateAndGetCommand(parameters.Devfile, parameters.DevfileDebugCmd, devfilev1.DebugCommandGroupKind)
		if e != nil {
			return nil, fmt.Errorf("debug command is not valid: %w", e)
		}
		pushDevfileCommands[devfilev1.DebugCommandGroupKind] = pushDevfileDebugCommands
	}

	return pushDevfileCommands, nil
}

func (o *DevClient) checkAppPorts(ctx context.Context, podName string, portsToFwd map[string][]devfilev1.Endpoint) error {
	containerPortsMapping := make(map[string][]int)
	for c, ports := range portsToFwd {
		for _, p := range ports {
			containerPortsMapping[c] = append(containerPortsMapping[c], p.TargetPort)
		}
	}
	return port.CheckAppPortsListening(ctx, o.execClient, podName, containerPortsMapping, 1*time.Minute)
}

// generateDeploymentObjectMeta generates a ObjectMeta object for the given deployment's name, labels and annotations
// if no deployment exists, it creates a new deployment name
func (o DevClient) generateDeploymentObjectMeta(ctx context.Context, deployment *appsv1.Deployment, labels map[string]string, annotations map[string]string) (metav1.ObjectMeta, error) {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)
	if deployment != nil {
		return generator.GetObjectMeta(deployment.Name, o.kubernetesClient.GetCurrentNamespace(), labels, annotations), nil
	} else {
		deploymentName, err := util.NamespaceKubernetesObject(componentName, appName)
		if err != nil {
			return metav1.ObjectMeta{}, err
		}
		return generator.GetObjectMeta(deploymentName, o.kubernetesClient.GetCurrentNamespace(), labels, annotations), nil
	}
}

// buildVolumes:
// - (side effect on cluster) creates the PVC for the project sources if Epehemeral preference is false
// - (side effect on cluster) creates the PVCs for non-ephemeral volumes defined in the Devfile
// - (side effect on input parameters) adds volumeMounts to containers and initContainers for the PVCs and Ephemeral volumes
// - (side effect on input parameters) adds volumeMounts for automounted volumes
// => Returns the list of Volumes to add to the PodTemplate
func (o *DevClient) buildVolumes(ctx context.Context, parameters common.PushParameters, containers, initContainers []corev1.Container) ([]corev1.Volume, error) {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)

	runtime := component.GetComponentRuntimeFromDevfileMetadata(parameters.Devfile.Data.GetMetadata())

	storageClient := storagepkg.NewClient(componentName, appName, storagepkg.ClientOptions{
		Client:  o.kubernetesClient,
		Runtime: runtime,
	})

	// Create the PVC for the project sources, if not ephemeral
	err := storage.HandleOdoSourceStorage(o.kubernetesClient, storageClient, componentName, o.prefClient.GetEphemeralSourceVolume())
	if err != nil {
		return nil, err
	}

	// Create PVCs for non-ephemeral Volumes defined in the Devfile
	// and returns the Ephemeral volumes defined in the Devfile
	ephemerals, err := storagepkg.Push(storageClient, parameters.Devfile)
	if err != nil {
		return nil, err
	}

	// get all the PVCs from the cluster belonging to the component
	// These PVCs have been created earlier with `storage.HandleOdoSourceStorage` and `storagepkg.Push`
	pvcs, err := o.kubernetesClient.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
	if err != nil {
		return nil, err
	}

	var allVolumes []corev1.Volume

	// Get the name of the PVC for project sources + a map of (storageName => VolumeInfo)
	// odoSourcePVCName will be empty when Ephemeral preference is true
	odoSourcePVCName, volumeNameToVolInfo, err := storage.GetVolumeInfos(pvcs)
	if err != nil {
		return nil, err
	}

	// Add the volumes for the projects source and the Odo-specific directory
	odoMandatoryVolumes := utils.GetOdoContainerVolumes(odoSourcePVCName)
	allVolumes = append(allVolumes, odoMandatoryVolumes...)

	// Add the volumeMounts for the project sources volume and the Odo-specific volume into the containers
	utils.AddOdoProjectVolume(containers)
	utils.AddOdoMandatoryVolume(containers)

	// Get PVC volumes and Volume Mounts
	pvcVolumes, err := storage.GetPersistentVolumesAndVolumeMounts(parameters.Devfile, containers, initContainers, volumeNameToVolInfo, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	allVolumes = append(allVolumes, pvcVolumes...)

	ephemeralVolumes, err := storage.GetEphemeralVolumesAndVolumeMounts(parameters.Devfile, containers, initContainers, ephemerals, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	allVolumes = append(allVolumes, ephemeralVolumes...)

	automountVolumes, err := storage.GetAutomountVolumes(o.configAutomountClient, containers, initContainers)
	if err != nil {
		return nil, err
	}
	allVolumes = append(allVolumes, automountVolumes...)

	return allVolumes, nil
}

func (o *DevClient) createOrUpdateServiceForComponent(ctx context.Context, svc *corev1.Service, ownerReference metav1.OwnerReference) error {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)
	oldSvc, err := o.kubernetesClient.GetOneService(componentName, appName, true)
	originOwnerReferences := svc.OwnerReferences
	if err != nil {
		// no old service was found, create a new one
		if len(svc.Spec.Ports) > 0 {
			err = o.kubernetesClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerReferences, ownerReference)
				_, err = o.kubernetesClient.CreateService(*svc)
				return err
			})
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}
		return nil
	}
	if len(svc.Spec.Ports) > 0 {
		svc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		svc.ResourceVersion = oldSvc.GetResourceVersion()
		err = o.kubernetesClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
			svc.OwnerReferences = append(originOwnerReferences, ownerRef)
			_, err = o.kubernetesClient.UpdateService(*svc)
			return err
		})
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully update Service for component %s", componentName)
		return nil
	}
	// delete the old existing service if the component currently doesn't expose any ports
	return o.kubernetesClient.DeleteService(oldSvc.Name)
}
