package component

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/envinfo"
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
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"

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

	AdapterContext
	logger machineoutput.MachineEventLoggingClient
}

// AdapterContext is a construct that is common to all adapters
type AdapterContext struct {
	ComponentName string            // ComponentName is the odo component name, it is NOT related to any devfile components
	Context       string            // Context is the given directory containing the source code and configs
	AppName       string            // the application name associated to a component
	Devfile       parser.DevfileObj // Devfile is the object returned by the Devfile parser
}

var _ sync.SyncClient = (*Adapter)(nil)
var _ ComponentAdapter = (*Adapter)(nil)

// NewKubernetesAdapter returns a Devfile adapter for the targeted platform
func NewKubernetesAdapter(
	kubernetesClient kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	context AdapterContext,
	namespace string,
) Adapter {

	if namespace != "" {
		kubernetesClient.SetNamespace(namespace)
	}

	return Adapter{
		kubeClient:        kubernetesClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		AdapterContext:    context,
		logger:            machineoutput.NewMachineEventLoggingClient(),
	}
}

// getPod lazily records and retrieves the pod associated with the component associated with this adapter. If refresh parameter
// is true, then the pod is refreshed from the cluster regardless of its current local state
func (a *Adapter) getPod(pod *corev1.Pod, refresh bool) (*corev1.Pod, error) {
	result := pod
	if refresh || result == nil {
		podSelector := fmt.Sprintf("component=%s", a.ComponentName)

		// Wait for Pod to be in running state otherwise we can't sync data to it.
		var err error
		result, err = a.kubeClient.WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, a.prefClient.GetPushTimeout())
		if err != nil {
			return nil, fmt.Errorf("error while waiting for pod %s: %w", podSelector, err)
		}
	}
	return result, nil
}

func (a *Adapter) ComponentInfo(pod *corev1.Pod, command devfilev1.Command) (adapters.ComponentInfo, error) {
	pod, err := a.getPod(pod, false)
	if err != nil {
		return adapters.ComponentInfo{}, err
	}
	return adapters.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: command.Exec.Component,
	}, nil
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
func (a Adapter) Push(parameters adapters.PushParameters) (err error) {

	// preliminary checks
	err = dfutil.ValidateK8sResourceName("component name", a.ComponentName)
	if err != nil {
		return err
	}

	err = dfutil.ValidateK8sResourceName("component namespace", parameters.EnvSpecificInfo.GetNamespace())
	if err != nil {
		return err
	}

	deployment, componentExists, err := a.getComponentDeployment()
	if err != nil {
		return err
	}

	// If the component already exists, retrieve the pod's name before it's potentially updated
	podName := ""
	if componentExists {
		podName, err = a.getPodName()
	}

	s := log.Spinner("Waiting for Kubernetes resources")
	defer s.End(false)

	// Set the mode to Dev since we are using "odo dev" here
	labels := odolabels.GetLabels(a.ComponentName, a.AppName, odolabels.ComponentDevMode)

	k8sComponents, err := a.pushKubernetesComponents(labels)
	if err != nil {
		return err
	}

	deployment, err = a.createOrUpdateComponent(componentExists, parameters.EnvSpecificInfo, libdevfile.DevfileCommands{
		BuildCmd: parameters.DevfileBuildCmd,
		RunCmd:   parameters.DevfileRunCmd,
		DebugCmd: parameters.DevfileDebugCmd,
	}, parameters.DebugPort, deployment)
	if err != nil {
		return fmt.Errorf("unable to create or update component: %w", err)
	}

	deployment, err = a.kubeClient.WaitForDeploymentRollout(deployment.Name)
	if err != nil {
		return fmt.Errorf("error while waiting for deployment rollout: %w", err)
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err := a.getPod(nil, true)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	ownerReference := generator.GetOwnerReference(deployment)
	err = a.updatePVCsOwnerReferences(ownerReference)
	if err != nil {
		return err
	}

	// Update all services with owner references
	err = a.kubeClient.TryWithBlockOwnerDeletion(ownerReference, func(ownerRef metav1.OwnerReference) error {
		return service.UpdateServicesWithOwnerReferences(a.kubeClient, a.Devfile, k8sComponents, ownerRef, a.Context)
	})
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	needRestart, err := service.PushLinks(a.kubeClient, a.Devfile, k8sComponents, labels, deployment, a.Context)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	if needRestart {
		err = a.kubeClient.WaitForPodDeletion(pod.Name)
		if err != nil {
			return err
		}
	}

	_, err = a.kubeClient.WaitForDeploymentRollout(deployment.Name)
	if err != nil {
		return fmt.Errorf("failed to update config to component deployed: %w", err)
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err = a.getPod(pod, true)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	parameters.EnvSpecificInfo.SetDevfileObj(a.Devfile)

	// Compare the name of the pod with the one before the rollout. If they differ, it means there's a new pod and a force push is required
	podChanged := false
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

	// Get commands
	pushDevfileCommands, err := a.getPushDevfileCommands(parameters)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	// Get a sync adapter. Check if project files have changed and sync accordingly
	syncAdapter := sync.New(&a, a.kubeClient, a.ComponentName)
	compInfo := adapters.ComponentInfo{
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}
	syncParams := adapters.SyncParameters{
		PushParams:      parameters,
		CompInfo:        compInfo,
		ComponentExists: componentExists,
		PodChanged:      podChanged,
		Files:           getSyncFilesFromAttributes(pushDevfileCommands),
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
			component.NewExecHandler(a.kubeClient, a.AppName, a.ComponentName, pod.Name, "", parameters.Show))
		if err != nil {
			return err
		}
	}

	cmdKind := devfilev1.RunCommandGroupKind
	cmdName := parameters.DevfileRunCmd
	if parameters.Debug {
		cmdKind = devfilev1.DebugCommandGroupKind
		cmdName = parameters.DevfileDebugCmd
	}

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
	cmdHandler := adapterHandler{
		Adapter:         a,
		parameters:      parameters,
		componentExists: componentExists,
		podName:         pod.GetName(),
	}

	if commandType == devfilev1.ExecCommandType {
		running, err = cmdHandler.isRemoteProcessForCommandRunning(cmd, pod.Name)
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

	klog.V(4).Infof("running=%v, execRequired=%v",
		running, execRequired)

	if isComposite || !running || execRequired {
		// Invoke the build command once (before calling libdevfile.ExecuteCommandByNameAndKind), as, if cmd is a composite command,
		// the handler we pass will be called for each command in that composite command.
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(a.kubeClient, a.AppName, a.ComponentName, pod.Name,
				"Building your application in container on cluster", parameters.Show)
			return libdevfile.Build(a.Devfile, parameters.DevfileBuildCmd, execHandler)
		}
		if componentExists {
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

	if podChanged {
		a.portForwardClient.StopPortForwarding()
	}

	err = a.portForwardClient.StartPortForwarding(a.Devfile, a.ComponentName, parameters.RandomPorts, parameters.ErrOut)
	if err != nil {
		return fmt.Errorf("fail starting the port forwarding: %w", err)
	}

	return nil
}

func (a *Adapter) createOrUpdateComponent(
	componentExists bool,
	ei envinfo.EnvSpecificInfo,
	commands libdevfile.DevfileCommands,
	devfileDebugPort int,
	deployment *appsv1.Deployment,
) (*appsv1.Deployment, error) {

	isMainStorageEphemeral := a.prefClient.GetEphemeralSourceVolume()

	ei.SetDevfileObj(a.Devfile)
	componentName := a.ComponentName

	storageClient := storagepkg.NewClient(componentName, a.AppName, storagepkg.ClientOptions{
		Client:              a.kubeClient,
		LocalConfigProvider: &ei,
	})

	// handle the ephemeral storage
	err := storage.HandleEphemeralStorage(a.kubeClient, storageClient, componentName, isMainStorageEphemeral)
	if err != nil {
		return nil, err
	}

	// From devfile info, create PVCs and return ephemeral storages
	ephemerals, err := storagepkg.Push(storageClient, &ei)
	if err != nil {
		return nil, err
	}

	// Set the labels
	labels := odolabels.GetLabels(componentName, a.AppName, odolabels.ComponentDevMode)
	labels["component"] = componentName

	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata()))
	klog.V(4).Infof("We are deploying these annotations: %s", annotations)

	containers, err := generator.GetContainers(a.Devfile, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("no valid components found in the devfile")
	}

	// Add the project volume before generating init containers
	utils.AddOdoProjectVolume(&containers)
	utils.AddOdoMandatoryVolume(&containers)

	containers, err = utils.UpdateContainerEnvVars(a.Devfile, containers, commands.DebugCmd, devfileDebugPort)
	if err != nil {
		return nil, err
	}

	containers, err = utils.UpdateContainersEntrypointsIfNeeded(a.Devfile, containers, commands.BuildCmd, commands.RunCmd, commands.DebugCmd)
	if err != nil {
		return nil, err
	}

	initContainers, err := generator.GetInitContainers(a.Devfile)
	if err != nil {
		return nil, err
	}

	// list all the pvcs for the component
	pvcs, err := a.kubeClient.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
	if err != nil {
		return nil, err
	}

	odoSourcePVCName, volumeNameToVolInfo, err := storage.GetVolumeInfos(pvcs)
	if err != nil {
		return nil, err
	}

	var allVolumes []corev1.Volume

	// Get PVC volumes and Volume Mounts
	pvcVolumes, err := storage.GetPersistentVolumesAndVolumeMounts(a.Devfile, containers, initContainers, volumeNameToVolInfo, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	allVolumes = append(allVolumes, pvcVolumes...)

	ephemeralVolumes, err := storage.GetEphemeralVolumesAndVolumeMounts(a.Devfile, containers, initContainers, ephemerals, parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	allVolumes = append(allVolumes, ephemeralVolumes...)

	odoMandatoryVolumes := utils.GetOdoContainerVolumes(odoSourcePVCName)
	allVolumes = append(allVolumes, odoMandatoryVolumes...)

	selectorLabels := map[string]string{
		"component": componentName,
	}

	deploymentObjectMeta, err := a.generateDeploymentObjectMeta(deployment, labels, annotations)
	if err != nil {
		return nil, err
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

	deployment, err = generator.GetDeployment(a.Devfile, deployParams)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	serviceObjectMeta := generator.GetObjectMeta(serviceName, a.kubeClient.GetCurrentNamespace(), labels, serviceAnnotations)
	serviceParams := generator.ServiceParams{
		ObjectMeta:     serviceObjectMeta,
		SelectorLabels: selectorLabels,
	}
	svc, err := generator.GetService(a.Devfile, serviceParams, parsercommon.DevfileOptions{})

	if err != nil {
		return nil, err
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
			return nil, err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		ownerReference := generator.GetOwnerReference(deployment)
		err = a.createOrUpdateServiceForComponent(svc, componentName, ownerReference)
		if err != nil {
			return nil, err
		}
	} else {
		if a.kubeClient.IsSSASupported() {
			deployment, err = a.kubeClient.ApplyDeployment(*deployment)
		} else {
			deployment, err = a.kubeClient.CreateDeployment(*deployment)
		}

		if err != nil {
			return nil, err
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
				return nil, err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}

	return deployment, nil
}

func (a *Adapter) createOrUpdateServiceForComponent(svc *corev1.Service, componentName string, ownerReference metav1.OwnerReference) error {
	oldSvc, err := a.kubeClient.GetOneService(a.ComponentName, a.AppName)
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
func (a Adapter) ExtractProjectToComponent(componentInfo adapters.ComponentInfo, targetPath string, stdin io.Reader) error {
	return a.kubeClient.ExtractProjectToComponent(componentInfo.ContainerName, componentInfo.PodName, targetPath, stdin)
}

// PushCommandsMap stores the commands to be executed as per their types.
type PushCommandsMap map[devfilev1.CommandGroupKind]devfilev1.Command

// getSyncFilesFromAttributes gets the target files and folders along with their respective remote destination from the devfile
// it uses the "dev.odo.push.path" attribute in the run command
func getSyncFilesFromAttributes(commandsMap PushCommandsMap) map[string]string {
	syncMap := make(map[string]string)
	if value, ok := commandsMap[devfilev1.RunCommandGroupKind]; ok {
		for key, value := range value.Attributes.Strings(nil) {
			if strings.HasPrefix(key, "dev.odo.push.path:") {
				localValue := strings.ReplaceAll(key, "dev.odo.push.path:", "")
				syncMap[filepath.Clean(localValue)] = filepath.ToSlash(filepath.Clean(value))
			}
		}
	}
	return syncMap
}
