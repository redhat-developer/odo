package component

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	devfilefs "github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"golang.org/x/sync/errgroup"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/binding"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/service"
	storagepkg "github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/watch"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/generator"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/v2/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	kubeClient        kclient.ClientInterface
	prefClient        preference.Client
	portForwardClient portForward.Client
	bindingClient     binding.Client
	syncClient        sync.Client
	execClient        exec.Client

	AdapterContext
	logger machineoutput.MachineEventLoggingClient
}

// AdapterContext is a construct that is common to all adapters
type AdapterContext struct {
	ComponentName string                // ComponentName is the odo component name, it is NOT related to any devfile components
	Context       string                // Context is the given directory containing the source code and configs
	AppName       string                // the application name associated to a component
	Devfile       parser.DevfileObj     // Devfile is the object returned by the Devfile parser
	FS            filesystem.Filesystem // FS is the object used for building image component if present
}

var _ ComponentAdapter = (*Adapter)(nil)

// NewKubernetesAdapter returns a Devfile adapter for the targeted platform
func NewKubernetesAdapter(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	bindingClient binding.Client,
	syncClient sync.Client,
	execClient exec.Client,
	context AdapterContext,
) Adapter {
	return Adapter{
		kubeClient:        kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		bindingClient:     bindingClient,
		syncClient:        syncClient,
		execClient:        execClient,
		AdapterContext:    context,
		logger:            machineoutput.NewMachineEventLoggingClient(),
	}
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
// The componentStatus will be modified to reflect the status of the component when the function returns
func (a Adapter) Push(ctx context.Context, parameters adapters.PushParameters, componentStatus *watch.ComponentStatus) (err error) {

	// preliminary checks
	err = dfutil.ValidateK8sResourceName("component name", a.ComponentName)
	if err != nil {
		return err
	}

	err = dfutil.ValidateK8sResourceName("component namespace", a.kubeClient.GetCurrentNamespace())
	if err != nil {
		return err
	}

	deployment, deploymentExists, err := a.getComponentDeployment()
	if err != nil {
		return err
	}

	if componentStatus.State != watch.StateWaitDeployment && componentStatus.State != watch.StateReady {
		log.SpinnerNoSpin("Waiting for Kubernetes resources")
	}

	// Set the mode to Dev since we are using "odo dev" here
	runtime := component.GetComponentRuntimeFromDevfileMetadata(a.Devfile.Data.GetMetadata())
	labels := odolabels.GetLabels(a.ComponentName, a.AppName, runtime, odolabels.ComponentDevMode, false)

	var updated bool
	deployment, updated, err = a.createOrUpdateComponent(deploymentExists, libdevfile.DevfileCommands{
		BuildCmd: parameters.DevfileBuildCmd,
		RunCmd:   parameters.DevfileRunCmd,
		DebugCmd: parameters.DevfileDebugCmd,
	}, deployment)
	if err != nil {
		return fmt.Errorf("unable to create or update component: %w", err)
	}
	ownerReference := generator.GetOwnerReference(deployment)

	// Delete remote resources that are not present in the Devfile
	selector := odolabels.GetSelector(a.ComponentName, a.AppName, odolabels.ComponentDevMode, false)

	objectsToRemove, serviceBindingSecretsToRemove, err := a.getRemoteResourcesNotPresentInDevfile(selector)
	if err != nil {
		return fmt.Errorf("unable to determine resources to delete: %w", err)
	}

	err = a.deleteRemoteResources(objectsToRemove)
	if err != nil {
		return fmt.Errorf("unable to delete remote resources: %w", err)
	}

	// this is mainly useful when the Service Binding Operator is not installed;
	// and the service binding secrets must be deleted manually since they are created by odo
	if len(serviceBindingSecretsToRemove) != 0 {
		err = a.deleteServiceBindingSecrets(serviceBindingSecretsToRemove, deployment)
		if err != nil {
			return fmt.Errorf("unable to delete service binding secrets: %w", err)
		}
	}

	// Create all the K8s components defined in the devfile
	_, err = a.pushDevfileKubernetesComponents(labels, odolabels.ComponentDevMode, ownerReference)
	if err != nil {
		return err
	}

	err = a.updatePVCsOwnerReferences(ownerReference)
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

	injected, err := a.bindingClient.CheckServiceBindingsInjectionDone(a.ComponentName, a.AppName)
	if err != nil {
		return err
	}

	if !injected {
		klog.V(4).Infof("Waiting for all service bindings to be injected...\n")
		return errors.New("some servicebindings are not injected")
	}

	// Check if endpoints changed in Devfile
	portsToForward, err := libdevfile.GetDevfileContainerEndpointMapping(a.Devfile, parameters.Debug)
	if err != nil {
		return err
	}
	portsChanged := !reflect.DeepEqual(portsToForward, a.portForwardClient.GetForwardedPorts())

	if componentStatus.State == watch.StateReady && !portsChanged {
		// If the deployment is already in Ready State, no need to continue
		return nil
	}

	// Now the Deployment has a Ready replica, we can get the Pod to work inside it
	pod, err := a.kubeClient.GetPodUsingComponentName(a.ComponentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, syncFolder, err := common.GetFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", pod.GetName(), err)
	}

	s := log.Spinner("Syncing files into the container")
	defer s.End(false)

	// Get commands
	pushDevfileCommands, err := a.getPushDevfileCommands(parameters)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	podChanged := componentStatus.State == watch.StateWaitDeployment

	// Get a sync adapter. Check if project files have changed and sync accordingly
	compInfo := sync.ComponentInfo{
		ComponentName: a.ComponentName,
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
		Path:                     parameters.Path,
		WatchFiles:               parameters.WatchFiles,
		WatchDeletedFiles:        parameters.WatchDeletedFiles,
		IgnoredFiles:             parameters.IgnoredFiles,
		DevfileScanIndexForWatch: parameters.DevfileScanIndexForWatch,

		CompInfo:  compInfo,
		ForcePush: !deploymentExists || podChanged,
		Files:     adapters.GetSyncFilesFromAttributes(pushDevfileCommands[cmdKind]),
	}

	execRequired, err := a.syncClient.SyncFiles(syncParams)
	if err != nil {
		componentStatus.State = watch.StateReady
		return fmt.Errorf("failed to sync to component with name %s: %w", a.ComponentName, err)
	}
	s.End(true)

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if !componentStatus.PostStartEventsDone && libdevfile.HasPostStartEvents(a.Devfile) {
		err = libdevfile.ExecPostStartEvents(a.Devfile,
			component.NewExecHandler(a.kubeClient, a.execClient, a.AppName, a.ComponentName, pod.Name, "Executing post-start command in container", parameters.Show))
		if err != nil {
			return err
		}
	}
	componentStatus.PostStartEventsDone = true

	cmd, err := libdevfile.ValidateAndGetCommand(a.Devfile, cmdName, cmdKind)
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
		fs:            a.FS,
		execClient:    a.execClient,
		kubeClient:    a.kubeClient,
		appName:       a.AppName,
		componentName: a.ComponentName,
		devfile:       a.Devfile,
		path:          parameters.Path,
		podName:       pod.GetName(),
		ctx:           ctx,
	}

	if commandType == devfilev1.ExecCommandType {
		running, err = cmdHandler.IsRemoteProcessForCommandRunning(cmd, pod.Name)
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
			execHandler := component.NewExecHandler(a.kubeClient, a.execClient, a.AppName, a.ComponentName, pod.Name,
				"Building your application in container", parameters.Show)
			return libdevfile.Build(a.Devfile, parameters.DevfileBuildCmd, execHandler)
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
		err = libdevfile.ExecuteCommandByNameAndKind(a.Devfile, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
	}

	if podChanged || portsChanged {
		a.portForwardClient.StopPortForwarding(a.ComponentName)
	}

	err = a.portForwardClient.StartPortForwarding(a.Devfile, a.ComponentName, parameters.Debug, parameters.RandomPorts, log.GetStdout(), parameters.ErrOut, nil)
	if err != nil {
		return adapters.NewErrPortForward(err)
	}
	componentStatus.EndpointsForwarded = a.portForwardClient.GetForwardedPorts()

	componentStatus.State = watch.StateReady
	return nil
}

// createOrUpdateComponent creates the deployment or updates it if it already exists
// with the expected spec.
// Returns the new deployment and if the generation of the deployment has been updated
func (a *Adapter) createOrUpdateComponent(
	componentExists bool,
	commands libdevfile.DevfileCommands,
	deployment *appsv1.Deployment,
) (*appsv1.Deployment, bool, error) {

	isMainStorageEphemeral := a.prefClient.GetEphemeralSourceVolume()

	componentName := a.ComponentName

	runtime := component.GetComponentRuntimeFromDevfileMetadata(a.Devfile.Data.GetMetadata())

	storageClient := storagepkg.NewClient(componentName, a.AppName, storagepkg.ClientOptions{
		Client:  a.kubeClient,
		Runtime: runtime,
	})

	// handle the ephemeral storage
	err := storage.HandleEphemeralStorage(a.kubeClient, storageClient, componentName, isMainStorageEphemeral)
	if err != nil {
		return nil, false, err
	}

	// From devfile info, create PVCs and return ephemeral storages
	ephemerals, err := storagepkg.Push(storageClient, a.Devfile)
	if err != nil {
		return nil, false, err
	}

	// Set the labels
	labels := odolabels.GetLabels(componentName, a.AppName, runtime, odolabels.ComponentDevMode, true)

	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata()))
	odolabels.AddCommonAnnotations(annotations)
	klog.V(4).Infof("We are deploying these annotations: %s", annotations)

	containers, err := generator.GetContainers(a.Devfile, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, false, err
	}
	if len(containers) == 0 {
		return nil, false, fmt.Errorf("no valid components found in the devfile")
	}

	// Add the project volume before generating init containers
	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(a.Devfile, containers, commands.BuildCmd, commands.RunCmd, commands.DebugCmd)
	if err != nil {
		return nil, false, err
	}

	initContainers, err := generator.GetInitContainers(a.Devfile)
	if err != nil {
		return nil, false, err
	}

	// list all the pvcs for the component
	pvcs, err := a.kubeClient.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
	if err != nil {
		return nil, false, err
	}

	odoSourcePVCName, volumeNameToVolInfo, err := storage.GetVolumeInfos(pvcs)
	if err != nil {
		return nil, false, err
	}

	var allVolumes []corev1.Volume

	// Get PVC volumes and Volume Mounts
	pvcVolumes, err := storage.GetPersistentVolumesAndVolumeMounts(a.Devfile, containers, initContainers, volumeNameToVolInfo, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, false, err
	}
	allVolumes = append(allVolumes, pvcVolumes...)

	ephemeralVolumes, err := storage.GetEphemeralVolumesAndVolumeMounts(a.Devfile, containers, initContainers, ephemerals, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, false, err
	}
	allVolumes = append(allVolumes, ephemeralVolumes...)

	odoMandatoryVolumes := utils.GetOdoContainerVolumes(odoSourcePVCName)
	allVolumes = append(allVolumes, odoMandatoryVolumes...)

	selectorLabels := map[string]string{
		"component": componentName,
	}

	deploymentObjectMeta, err := a.generateDeploymentObjectMeta(deployment, labels, annotations)
	if err != nil {
		return nil, false, err
	}

	deployParams := generator.DeploymentParams{
		TypeMeta:          generator.GetTypeMeta(kclient.DeploymentKind, kclient.DeploymentAPIVersion),
		ObjectMeta:        deploymentObjectMeta,
		InitContainers:    initContainers,
		Containers:        containers,
		Volumes:           allVolumes,
		PodSelectorLabels: selectorLabels,
		Replicas:          pointer.Int32Ptr(1),
	}

	// Save generation to check if deployment is updated later
	var originalGeneration int64 = 0
	if deployment != nil {
		originalGeneration = deployment.GetGeneration()
	}

	deployment, err = generator.GetDeployment(a.Devfile, deployParams)
	if err != nil {
		return nil, false, err
	}
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	if vcsUri := util.GetGitOriginPath(a.Context); vcsUri != "" {
		deployment.Annotations["app.openshift.io/vcs-uri"] = vcsUri
	}

	// add the annotations to the service for linking
	serviceAnnotations := make(map[string]string)
	serviceAnnotations["service.binding/backend_ip"] = "path={.spec.clusterIP}"
	serviceAnnotations["service.binding/backend_port"] = "path={.spec.ports},elementType=sliceOfMaps,sourceKey=name,sourceValue=port"

	serviceName, err := util.NamespaceKubernetesObjectWithTrim(componentName, a.AppName, 63)
	if err != nil {
		return nil, false, err
	}
	serviceObjectMeta := generator.GetObjectMeta(serviceName, a.kubeClient.GetCurrentNamespace(), labels, serviceAnnotations)
	serviceParams := generator.ServiceParams{
		ObjectMeta:     serviceObjectMeta,
		SelectorLabels: selectorLabels,
	}
	svc, err := generator.GetService(a.Devfile, serviceParams, parsercommon.DevfileOptions{})

	if err != nil {
		return nil, false, err
	}
	klog.V(2).Infof("Creating deployment %v", deployment.Spec.Template.GetName())
	klog.V(2).Infof("The component name is %v", componentName)
	if componentExists {
		// If the component already exists, get the resource version of the deploy before updating
		klog.V(2).Info("The component already exists, attempting to update it")
		if a.kubeClient.IsSSASupported() {
			klog.V(4).Info("Applying deployment")
			deployment, err = a.kubeClient.ApplyDeployment(*deployment)
		} else {
			klog.V(4).Info("Updating deployment")
			deployment, err = a.kubeClient.UpdateDeployment(*deployment)
		}
		if err != nil {
			return nil, false, err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		ownerReference := generator.GetOwnerReference(deployment)
		err = a.createOrUpdateServiceForComponent(svc, componentName, ownerReference)
		if err != nil {
			return nil, false, err
		}
	} else {
		if a.kubeClient.IsSSASupported() {
			deployment, err = a.kubeClient.ApplyDeployment(*deployment)
		} else {
			deployment, err = a.kubeClient.CreateDeployment(*deployment)
		}

		if err != nil {
			return nil, false, err
		}

		klog.V(2).Infof("Successfully created component %v", componentName)
		if len(svc.Spec.Ports) > 0 {
			ownerReference := generator.GetOwnerReference(deployment)
			originOwnerRefs := svc.OwnerReferences
			err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerRefs, ownerRef)
				_, err = a.kubeClient.CreateService(*svc)
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

func (a *Adapter) createOrUpdateServiceForComponent(svc *corev1.Service, componentName string, ownerReference metav1.OwnerReference) error {
	oldSvc, err := a.kubeClient.GetOneService(a.ComponentName, a.AppName, true)
	originOwnerReferences := svc.OwnerReferences
	if err != nil {
		// no old service was found, create a new one
		if len(svc.Spec.Ports) > 0 {
			err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerReferences, ownerReference)
				_, err = a.kubeClient.CreateService(*svc)
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
		err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
			svc.OwnerReferences = append(originOwnerReferences, ownerRef)
			_, err = a.kubeClient.UpdateService(*svc)
			return err
		})
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully update Service for component %s", componentName)
		return nil
	}
	// delete the old existing service if the component currently doesn't expose any ports
	return a.kubeClient.DeleteService(oldSvc.Name)
}

// generateDeploymentObjectMeta generates a ObjectMeta object for the given deployment's name, labels and annotations
// if no deployment exists, it creates a new deployment name
func (a Adapter) generateDeploymentObjectMeta(deployment *appsv1.Deployment, labels map[string]string, annotations map[string]string) (metav1.ObjectMeta, error) {
	if deployment != nil {
		return generator.GetObjectMeta(deployment.Name, a.kubeClient.GetCurrentNamespace(), labels, annotations), nil
	} else {
		deploymentName, err := util.NamespaceKubernetesObject(a.ComponentName, a.AppName)
		if err != nil {
			return metav1.ObjectMeta{}, err
		}
		return generator.GetObjectMeta(deploymentName, a.kubeClient.GetCurrentNamespace(), labels, annotations), nil
	}
}

// getRemoteResourcesNotPresentInDevfile compares the list of Devfile K8s component and remote K8s resources
// and returns a list of the remote resources not present in the Devfile and in case the SBO is not installed, a list of service binding secrets that must be deleted;
// it ignores the core components (such as deployments, svc, pods; all resources with `component:<something>` label)
func (a Adapter) getRemoteResourcesNotPresentInDevfile(selector string) (objectsToRemove, serviceBindingSecretsToRemove []unstructured.Unstructured, err error) {
	currentNamespace := a.kubeClient.GetCurrentNamespace()
	allRemoteK8sResources, err := a.kubeClient.GetAllResourcesFromSelector(selector, currentNamespace)
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
	devfileK8sResources, err = devfile.GetK8sAndOcComponentsToPush(a.Devfile, true)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to obtain resources from the Devfile: %w", err)
	}

	// convert all devfileK8sResources to unstructured data
	var devfileK8sResourcesUnstructured []unstructured.Unstructured
	for _, devfileK := range devfileK8sResources {
		var devfileKUnstructuredList []unstructured.Unstructured
		devfileKUnstructuredList, err = libdevfile.GetK8sComponentAsUnstructuredList(a.Devfile, devfileK.Name, a.Context, devfilefs.DefaultFs{})
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read the resource: %w", err)
		}
		devfileK8sResourcesUnstructured = append(devfileK8sResourcesUnstructured, devfileKUnstructuredList...)
	}

	isSBOSupported, err := a.kubeClient.IsServiceBindingSupported()
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
func (a Adapter) deleteRemoteResources(objectsToRemove []unstructured.Unstructured) error {
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
			gvr, err := a.kubeClient.GetGVRFromGVK(objectToRemove.GroupVersionKind())
			if err != nil {
				return fmt.Errorf("unable to get information about resource: %s/%s: %w", objectToRemove.GetKind(), objectToRemove.GetName(), err)
			}

			err = a.kubeClient.DeleteDynamicResource(objectToRemove.GetName(), gvr, true)
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
func (a Adapter) deleteServiceBindingSecrets(serviceBindingSecretsToRemove []unstructured.Unstructured, deployment *appsv1.Deployment) error {
	for _, secretToRemove := range serviceBindingSecretsToRemove {
		spinner := log.Spinnerf("Deleting Kubernetes resource: %s/%s", secretToRemove.GetKind(), secretToRemove.GetName())
		defer spinner.End(false)

		err := service.UnbindWithLibrary(a.kubeClient, secretToRemove, deployment)
		if err != nil {
			return fmt.Errorf("failed to unbind secret %q from the application", secretToRemove.GetName())
		}

		// since the library currently doesn't delete the secret after unbinding
		// delete the secret manually
		err = a.kubeClient.DeleteSecret(secretToRemove.GetName(), a.kubeClient.GetCurrentNamespace())
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

// PushCommandsMap stores the commands to be executed as per their types.
type PushCommandsMap map[devfilev1.CommandGroupKind]devfilev1.Command
