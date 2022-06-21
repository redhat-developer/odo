package component

import (
	"fmt"
	"io"

	"k8s.io/utils/pointer"

	"github.com/devfile/library/pkg/devfile/generator"
	devfileCommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/service"
	storagepkg "github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// New instantiates a component adapter
func New(adapterContext common.AdapterContext, kubeClient kclient.ClientInterface, prefClient preference.Client) Adapter {
	return Adapter{
		kubeClient:     kubeClient,
		prefClient:     prefClient,
		AdapterContext: adapterContext,
		logger:         machineoutput.NewMachineEventLoggingClient(),
	}
}

// getPod lazily records and retrieves the pod associated with the component associated with this adapter. If refresh parameter
// is true, then the pod is refreshed from the cluster regardless of its current local state
func (a *Adapter) getPod(refresh bool) (*corev1.Pod, error) {
	if refresh || a.pod == nil {
		podSelector := fmt.Sprintf("component=%s", a.ComponentName)

		// Wait for Pod to be in running state otherwise we can't sync data to it.
		pod, err := a.kubeClient.WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, a.prefClient.GetPushTimeout())
		if err != nil {
			return nil, fmt.Errorf("error while waiting for pod %s: %w", podSelector, err)
		}
		a.pod = pod
	}
	return a.pod, nil
}

func (a *Adapter) ComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	pod, err := a.getPod(false)
	if err != nil {
		return common.ComponentInfo{}, err
	}
	return common.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: command.Exec.Component,
	}, nil
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	kubeClient kclient.ClientInterface
	prefClient preference.Client

	common.AdapterContext
	logger machineoutput.MachineEventLoggingClient

	devfileBuildCmd  string
	devfileRunCmd    string
	devfileDebugCmd  string
	devfileDebugPort int
	pod              *corev1.Pod
	deployment       *appsv1.Deployment
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
func (a Adapter) Push(parameters common.PushParameters) (err error) {

	// Get the Dev deployment:
	// Since `odo deploy` can theoretically deploy a deployment as well with the same instance name
	// we make sure that we are retrieving the deployment with the Dev mode, NOT Deploy.
	selectorLabels := odolabels.GetSelector(a.ComponentName, a.AppName, odolabels.ComponentDevMode)
	a.deployment, err = a.kubeClient.GetOneDeploymentFromSelector(selectorLabels)

	if err != nil {
		if _, ok := err.(*kclient.DeploymentNotFoundError); !ok {
			return fmt.Errorf("unable to determine if component %s exists: %w", a.ComponentName, err)
		}
	}
	componentExists := a.deployment != nil

	a.devfileBuildCmd = parameters.DevfileBuildCmd
	a.devfileRunCmd = parameters.DevfileRunCmd
	a.devfileDebugCmd = parameters.DevfileDebugCmd
	a.devfileDebugPort = parameters.DebugPort

	podChanged := false
	var podName string

	// If the component already exists, retrieve the pod's name before it's potentially updated
	if componentExists {
		// First see if the component does have a pod. it could have been scaled down to zero
		_, err = a.kubeClient.GetOnePodFromSelector(fmt.Sprintf("component=%s", a.ComponentName))
		// If an error occurs, we don't call a.getPod (a blocking function that waits till it finds a pod in "Running" state.)
		// We would rely on a call to a.createOrUpdateComponent to reset the pod count for the component to one.
		if err == nil {
			pod, podErr := a.getPod(true)
			if podErr != nil {
				return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, podErr)
			}
			podName = pod.GetName()
		}
	}

	s := log.Spinner("Waiting for Kubernetes resources")
	defer s.End(false)

	err = dfutil.ValidateK8sResourceName("component name", a.ComponentName)
	if err != nil {
		return err
	}

	err = dfutil.ValidateK8sResourceName("component namespace", parameters.EnvSpecificInfo.GetNamespace())
	if err != nil {
		return err
	}

	pushDevfileCommands, err := libdevfile.ValidateAndGetPushCommands(a.Devfile, a.devfileBuildCmd, a.devfileRunCmd)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	// Set the mode to Dev since we are using "odo dev" here
	labels := odolabels.GetLabels(a.ComponentName, a.AppName, odolabels.ComponentDevMode)

	// Set the annotations for the component type
	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata()))

	previousMode := parameters.EnvSpecificInfo.GetRunMode()
	currentMode := envinfo.Run

	if parameters.Debug {
		pushDevfileDebugCommands, e := libdevfile.ValidateAndGetCommand(a.Devfile, a.devfileDebugCmd, devfilev1.DebugCommandGroupKind)
		if e != nil {
			return fmt.Errorf("debug command is not valid: %w", e)
		}
		pushDevfileCommands[devfilev1.DebugCommandGroupKind] = pushDevfileDebugCommands
		currentMode = envinfo.Debug
	}

	if currentMode != previousMode {
		parameters.RunModeChanged = true
	}

	// fetch the "kubernetes inlined components" to create them on cluster
	// from odo standpoint, these components contain yaml manifest of an odo service or an odo link
	k8sComponents, err := devfile.GetKubernetesComponentsToPush(a.Devfile)
	if err != nil {
		return fmt.Errorf("error while trying to fetch service(s) from devfile: %w", err)
	}

	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	err = service.ValidateResourcesExist(a.kubeClient, a.Devfile, k8sComponents, a.Context)
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	err = service.PushKubernetesResources(a.kubeClient, a.Devfile, k8sComponents, labels, annotations, a.Context)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	isMainStorageEphemeral := a.prefClient.GetEphemeralSourceVolume()
	err = a.createOrUpdateComponent(componentExists, parameters.EnvSpecificInfo, isMainStorageEphemeral)
	if err != nil {
		return fmt.Errorf("unable to create or update component: %w", err)
	}

	a.deployment, err = a.kubeClient.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return fmt.Errorf("error while waiting for deployment rollout: %w", err)
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err := a.getPod(true)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	// list the latest state of the PVCs
	pvcs, err := a.kubeClient.ListPVCs(fmt.Sprintf("%v=%v", "component", a.ComponentName))
	if err != nil {
		return err
	}

	ownerReference := generator.GetOwnerReference(a.deployment)
	// update the owner reference of the PVCs with the deployment
	for i := range pvcs {
		if pvcs[i].OwnerReferences != nil || pvcs[i].DeletionTimestamp != nil {
			continue
		}
		err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
			return a.kubeClient.UpdateStorageOwnerReference(&pvcs[i], ownerRef)
		})
		if err != nil {
			return err
		}
	}

	// Update all services with owner references
	err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
		return service.UpdateServicesWithOwnerReferences(a.kubeClient, a.Devfile, k8sComponents, ownerRef, a.Context)
	})
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	needRestart, err := service.PushLinks(a.kubeClient, a.Devfile, k8sComponents, labels, a.deployment, a.Context)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	if needRestart {
		err = a.kubeClient.WaitForPodDeletion(pod.Name)
		if err != nil {
			return err
		}
	}

	a.deployment, err = a.kubeClient.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return fmt.Errorf("failed to update config to component deployed: %w", err)
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err = a.getPod(true)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	parameters.EnvSpecificInfo.SetDevfileObj(a.Devfile)

	// Compare the name of the pod with the one before the rollout. If they differ, it means there's a new pod and a force push is required
	if componentExists && podName != pod.GetName() {
		podChanged = true
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, syncFolder, err := getFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", podName, err)
	}
	s.End(true)

	s = log.Spinner("Syncing files into the container")
	defer s.End(false)
	// Get a sync adapter. Check if project files have changed and sync accordingly
	syncAdapter := sync.New(a.AdapterContext, &a, a.kubeClient)
	compInfo := common.ComponentInfo{
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}
	syncParams := common.SyncParameters{
		PushParams:      parameters,
		CompInfo:        compInfo,
		ComponentExists: componentExists,
		PodChanged:      podChanged,
		Files:           common.GetSyncFilesFromAttributes(pushDevfileCommands),
	}

	execRequired, err := syncAdapter.SyncFiles(syncParams)
	if err != nil {
		return fmt.Errorf("failed to sync to component with name %s: %w", a.ComponentName, err)
	}
	s.End(true)

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if !componentExists && libdevfile.HasPostStartEvents(a.Devfile) {
		err = libdevfile.ExecPostStartEvents(a.Devfile,
			component.NewExecHandler(a.kubeClient, a.AppName, a.ComponentName, a.pod.Name, "", parameters.Show))
		if err != nil {
			return err
		}
	}

	cmdKind := devfilev1.RunCommandGroupKind
	if parameters.Debug {
		cmdKind = devfilev1.DebugCommandGroupKind
	}

	cmd, err := libdevfile.GetDefaultCommand(a.Devfile, cmdKind)
	if err != nil {
		return err
	}

	cmdHandler := adapterHandler{
		Adapter:         a,
		parameters:      parameters,
		componentExists: componentExists,
	}

	commandType, err := devfileCommon.GetCommandType(cmd)
	if err != nil {
		return err
	}
	var running bool
	var isComposite bool
	if commandType == devfilev1.ExecCommandType {
		running, err = cmdHandler.isRemoteProcessForCommandRunning(cmd)
		if err != nil {
			return err
		}
	} else if commandType == devfilev1.CompositeCommandType {
		//this handler will run each command in this composite command individually,
		//and will determine whether each command is running or not.
		isComposite = true
	} else {
		return fmt.Errorf("unsupported type %q for Devfile command %s, only exec and composite are handled",
			commandType, cmd.Id)
	}

	klog.V(4).Infof("running=%v, execRequired=%v, parameters.RunModeChanged=%v",
		running, execRequired, parameters.RunModeChanged)

	if isComposite || !running || execRequired || parameters.RunModeChanged {
		// Invoke the build command once (before calling libdevfile.ExecuteCommandByKind), as, if cmd is a composite command,
		// the handler we pass will be called for each command in that composite command.
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(a.kubeClient, a.AppName, a.ComponentName, a.pod.Name,
				"Building your application in container on cluster", parameters.Show)
			return libdevfile.Build(a.Devfile, execHandler)
		}
		if componentExists {
			if parameters.RunModeChanged || cmd.Exec == nil || !util.SafeGetBool(cmd.Exec.HotReloadCapable) {
				if err = doExecuteBuildCommand(); err != nil {
					return err
				}
			}
		} else {
			if err = doExecuteBuildCommand(); err != nil {
				return err
			}
		}
		err = libdevfile.ExecuteCommandByKind(a.Devfile, cmdKind, &cmdHandler, false)
	}

	return err
}

func (a *Adapter) createOrUpdateComponent(componentExists bool, ei envinfo.EnvSpecificInfo, isMainStorageEphemeral bool) (err error) {
	ei.SetDevfileObj(a.Devfile)
	componentName := a.ComponentName

	storageClient := storagepkg.NewClient(storagepkg.ClientOptions{
		Client:              a.kubeClient,
		LocalConfigProvider: &ei,
	})

	// handle the ephemeral storage
	err = storage.HandleEphemeralStorage(a.kubeClient, storageClient, componentName, isMainStorageEphemeral)
	if err != nil {
		return err
	}

	// From devfile info, create PVCs and return ephemeral storages
	ephemerals, err := storagepkg.Push(storageClient, &ei)
	if err != nil {
		return err
	}

	// Set the labels
	labels := odolabels.GetLabels(componentName, a.AppName, odolabels.ComponentDevMode)
	labels["component"] = componentName

	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata()))
	klog.V(4).Infof("We are deploying these annotations: %s", annotations)

	containers, err := generator.GetContainers(a.Devfile, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return fmt.Errorf("no valid components found in the devfile")
	}

	// Add the project volume before generating init containers
	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	containers, err = utils.UpdateContainerEnvVars(a.Devfile, containers, a.devfileDebugCmd, a.devfileDebugPort)
	if err != nil {
		return err
	}

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(a.Devfile, containers, a.devfileBuildCmd, a.devfileRunCmd, a.devfileDebugCmd)
	if err != nil {
		return err
	}

	initContainers, err := generator.GetInitContainers(a.Devfile)
	if err != nil {
		return err
	}

	// list all the pvcs for the component
	pvcs, err := a.kubeClient.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
	if err != nil {
		return err
	}

	odoSourcePVCName, volumeNameToVolInfo, err := storage.GetVolumeInfos(pvcs)
	if err != nil {
		return err
	}

	var allVolumes []corev1.Volume

	// Get PVC volumes and Volume Mounts
	pvcVolumes, err := storage.GetPersistentVolumesAndVolumeMounts(a.Devfile, containers, initContainers, volumeNameToVolInfo, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	allVolumes = append(allVolumes, pvcVolumes...)

	ephemeralVolumes, err := storage.GetEphemeralVolumesAndVolumeMounts(a.Devfile, containers, initContainers, ephemerals, parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	allVolumes = append(allVolumes, ephemeralVolumes...)

	odoMandatoryVolumes := utils.GetOdoContainerVolumes(odoSourcePVCName)
	allVolumes = append(allVolumes, odoMandatoryVolumes...)

	selectorLabels := map[string]string{
		"component": componentName,
	}

	deploymentObjectMeta, err := a.generateDeploymentObjectMeta(labels, annotations)
	if err != nil {
		return err
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

	deployment, err := generator.GetDeployment(a.Devfile, deployParams)
	if err != nil {
		return err
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

	serviceName, err := util.NamespaceKubernetesObjectWithTrim(componentName, a.AppName)
	if err != nil {
		return err
	}
	serviceObjectMeta := generator.GetObjectMeta(serviceName, a.kubeClient.GetCurrentNamespace(), labels, serviceAnnotations)
	serviceParams := generator.ServiceParams{
		ObjectMeta:     serviceObjectMeta,
		SelectorLabels: selectorLabels,
	}
	svc, err := generator.GetService(a.Devfile, serviceParams, parsercommon.DevfileOptions{})

	if err != nil {
		return err
	}
	klog.V(2).Infof("Creating deployment %v", deployment.Spec.Template.GetName())
	klog.V(2).Infof("The component name is %v", componentName)
	if componentExists {
		// If the component already exists, get the resource version of the deploy before updating
		klog.V(2).Info("The component already exists, attempting to update it")
		if a.kubeClient.IsSSASupported() {
			klog.V(4).Info("Applying deployment")
			a.deployment, err = a.kubeClient.ApplyDeployment(*deployment)
		} else {
			klog.V(4).Info("Updating deployment")
			a.deployment, err = a.kubeClient.UpdateDeployment(*deployment)
		}
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		err = a.createOrUpdateServiceForComponent(svc, componentName)
		if err != nil {
			return err
		}
	} else {
		if a.kubeClient.IsSSASupported() {
			a.deployment, err = a.kubeClient.ApplyDeployment(*deployment)
		} else {
			a.deployment, err = a.kubeClient.CreateDeployment(*deployment)
		}

		if err != nil {
			return err
		}

		klog.V(2).Infof("Successfully created component %v", componentName)
		if len(svc.Spec.Ports) > 0 {
			ownerReference := generator.GetOwnerReference(a.deployment)
			originOwnerRefs := svc.OwnerReferences
			err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerRefs, ownerRef)
				_, err = a.kubeClient.CreateService(*svc)
				return err
			})
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}

	return nil
}

func (a *Adapter) createOrUpdateServiceForComponent(svc *corev1.Service, componentName string) error {
	oldSvc, err := a.kubeClient.GetOneService(a.ComponentName, a.AppName)
	originOwnerReferences := svc.OwnerReferences
	ownerReference := generator.GetOwnerReference(a.deployment)
	if err != nil {
		// no old service was found, create a new one
		if len(svc.Spec.Ports) > 0 {
			err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
				svc.OwnerReferences = append(originOwnerReferences, ownerRef)
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
func (a Adapter) generateDeploymentObjectMeta(labels map[string]string, annotations map[string]string) (metav1.ObjectMeta, error) {
	if a.deployment != nil {
		return generator.GetObjectMeta(a.deployment.Name, a.kubeClient.GetCurrentNamespace(), labels, annotations), nil
	} else {
		deploymentName, err := util.NamespaceKubernetesObject(a.ComponentName, a.AppName)
		if err != nil {
			return metav1.ObjectMeta{}, err
		}
		return generator.GetObjectMeta(deploymentName, a.kubeClient.GetCurrentNamespace(), labels, annotations), nil
	}
}

// getFirstContainerWithSourceVolume returns the first container that set mountSources: true as well
// as the path to the source volume inside the container.
// Because the source volume is shared across all components that need it, we only need to sync once,
// so we only need to find one container. If no container was found, that means there's no
// container to sync to, so return an error
func getFirstContainerWithSourceVolume(containers []corev1.Container) (string, string, error) {
	for _, c := range containers {
		for _, env := range c.Env {
			if env.Name == generator.EnvProjectsSrc {
				return c.Name, env.Value, nil
			}
		}
	}

	return "", "", fmt.Errorf("in order to sync files, odo requires at least one component in a devfile to set 'mountSources: true'")
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (a Adapter) ExtractProjectToComponent(componentInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {
	return a.kubeClient.ExtractProjectToComponent(componentInfo.ContainerName, componentInfo.PodName, targetPath, stdin)
}
