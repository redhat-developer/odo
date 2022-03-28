package component

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"k8s.io/utils/pointer"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/redhat-developer/odo/pkg/component"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
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

const supervisorDStatusWaitTimeInterval = 1

// New instantiates a component adapter
func New(adapterContext common.AdapterContext, client kclient.ClientInterface, prefClient preference.Client) Adapter {

	adapter := Adapter{Client: client, prefClient: prefClient}
	adapter.GenericAdapter = common.NewGenericAdapter(&adapter, adapterContext)
	adapter.GenericAdapter.InitWith(&adapter)
	return adapter
}

// getPod lazily records and retrieves the pod associated with the component associated with this adapter. If refresh parameter
// is true, then the pod is refreshed from the cluster regardless of its current local state
func (a *Adapter) getPod(refresh bool) (*corev1.Pod, error) {
	if refresh || a.pod == nil {
		podSelector := fmt.Sprintf("component=%s", a.ComponentName)

		// Wait for Pod to be in running state otherwise we can't sync data to it.
		pod, err := a.Client.WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, time.Duration(a.prefClient.GetPushTimeout())*time.Second)
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

func (a *Adapter) SupervisorComponentInfo(command devfilev1.Command) (common.ComponentInfo, error) {
	pod, err := a.getPod(false)
	if err != nil {
		return common.ComponentInfo{}, err
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == command.Exec.Component && !reflect.DeepEqual(container.Command, []string{common.SupervisordBinaryPath}) {
			return common.ComponentInfo{
				ContainerName: command.Exec.Component,
				PodName:       pod.Name,
			}, nil
		}
	}
	return common.ComponentInfo{}, nil
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client     kclient.ClientInterface
	prefClient preference.Client

	*common.GenericAdapter

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
	selectorLabels := componentlabels.GetLabels(a.ComponentName, a.AppName, false)
	selectorLabels[componentlabels.OdoModeLabel] = componentlabels.ComponentDevName
	a.deployment, err = a.Client.GetOneDeploymentFromSelector(util.ConvertLabelsToSelector(selectorLabels))

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
		_, err = a.Client.GetOnePodFromSelector(fmt.Sprintf("component=%s", a.ComponentName))
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

	pushDevfileCommands, err := common.ValidateAndGetPushDevfileCommands(a.Devfile.Data, a.devfileBuildCmd, a.devfileRunCmd)
	if err != nil {
		return fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	// Set the mode to Dev since we are using "odo dev" here
	labels := componentlabels.GetLabels(a.ComponentName, a.AppName, true)
	labels[componentlabels.OdoModeLabel] = componentlabels.ComponentDevName

	// Set the annotations for the component type
	annotations := make(map[string]string)
	annotations[componentlabels.OdoProjectTypeAnnotation] = component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata())

	previousMode := parameters.EnvSpecificInfo.GetRunMode()
	currentMode := envinfo.Run

	if parameters.Debug {
		pushDevfileDebugCommands, e := common.ValidateAndGetDebugDevfileCommands(a.Devfile.Data, a.devfileDebugCmd)
		if e != nil {
			return fmt.Errorf("debug command is not valid: %w", err)
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
	err = service.ValidateResourcesExist(a.Client, k8sComponents, a.Context)
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	err = service.PushKubernetesResources(a.Client, k8sComponents, labels, annotations, a.Context)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	isMainStorageEphemeral := a.prefClient.GetEphemeralSourceVolume()
	err = a.createOrUpdateComponent(componentExists, parameters.EnvSpecificInfo, isMainStorageEphemeral)
	if err != nil {
		return fmt.Errorf("unable to create or update component: %w", err)
	}

	a.deployment, err = a.Client.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return fmt.Errorf("error while waiting for deployment rollout: %w", err)
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err := a.getPod(true)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, err)
	}

	// list the latest state of the PVCs
	pvcs, err := a.Client.ListPVCs(fmt.Sprintf("%v=%v", "component", a.ComponentName))
	if err != nil {
		return err
	}

	ownerReference := generator.GetOwnerReference(a.deployment)
	// update the owner reference of the PVCs with the deployment
	for i := range pvcs {
		if pvcs[i].OwnerReferences != nil || pvcs[i].DeletionTimestamp != nil {
			continue
		}
		err = a.Client.UpdateStorageOwnerReference(&pvcs[i], ownerReference)
		if err != nil {
			return err
		}
	}

	// Update all services with owner references
	err = service.UpdateServicesWithOwnerReferences(a.Client, k8sComponents, ownerReference, a.Context)
	if err != nil {
		return err
	}

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	needRestart, err := service.PushLinks(a.Client, k8sComponents, labels, a.deployment, a.Context)
	if err != nil {
		return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}

	if needRestart {
		err = a.Client.WaitForPodDeletion(pod.Name)
		if err != nil {
			return err
		}
	}

	a.deployment, err = a.Client.WaitForDeploymentRollout(a.deployment.Name)
	if err != nil {
		return fmt.Errorf("Failed to update config to component deployed: %w", err)
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
	syncAdapter := sync.New(a.AdapterContext, &a)
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
		return fmt.Errorf("Failed to sync to component with name %s: %w", a.ComponentName, err)
	}
	s.End(true)

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if !componentExists && libdevfile.HasPostStartEvents(a.Devfile) {
		err = libdevfile.ExecPostStartEvents(a.Devfile, a.ComponentName, component.NewExecHandler(a.Client, a.pod.Name, parameters.Show))
		if err != nil {
			return err
		}
	}

	runCommand := pushDevfileCommands[devfilev1.RunCommandGroupKind]
	if parameters.Debug {
		runCommand = pushDevfileCommands[devfilev1.DebugCommandGroupKind]
	}
	running, err := a.GetSupervisordCommandStatus(runCommand)
	if err != nil {
		return err
	}

	if !running || execRequired || parameters.RunModeChanged {
		err = a.ExecDevfile(pushDevfileCommands, componentExists, parameters)
		if err != nil {
			return err
		}

		// wait for a second
		wait := time.After(supervisorDStatusWaitTimeInterval * time.Second)
		<-wait

		err := a.CheckSupervisordCommandStatus(runCommand)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSupervisordCommandStatus returns true if the command is running
// based on `supervisord ctl` output and returns an error if
// the command is not known by supervisord
func (a Adapter) GetSupervisordCommandStatus(command devfilev1.Command) (bool, error) {
	statusInContainer := getSupervisordStatusInContainer(a.pod.Name, command.Exec.Component, a)

	supervisordProgramName := "devrun"

	// if the command is a debug one, we check against `debugrun`
	if command.Exec.Group.Kind == devfilev1.DebugCommandGroupKind {
		supervisordProgramName = "debugrun"
	}

	for _, status := range statusInContainer {
		if strings.EqualFold(status.program, supervisordProgramName) {
			return strings.EqualFold(status.status, "running"), nil
		}
	}
	return false, fmt.Errorf("the supervisord program %s not found", supervisordProgramName)
}

// CheckSupervisordCommandStatus checks if the command is running based on supervisord status output.
// if the command is not in a running state, we fetch the last 20 lines of the component's log and display it
func (a Adapter) CheckSupervisordCommandStatus(command devfilev1.Command) error {

	running, err := a.GetSupervisordCommandStatus(command)
	if err != nil {
		return err
	}

	if !running {
		numberOfLines := 20
		log.Warningf("devfile command %q exited with error status within %d sec", command.Id, supervisorDStatusWaitTimeInterval)
		log.Infof("Last %d lines of the component's log:", numberOfLines)

		rd, err := component.Log(a.Client, a.ComponentName, a.AppName, false, command)
		if err != nil {
			return err
		}

		err = util.DisplayLog(false, rd, os.Stderr, a.ComponentName, numberOfLines)
		if err != nil {
			return err
		}

		log.Info("To get the full log output, please run 'odo log'")
	}
	return nil
}

func (a *Adapter) createOrUpdateComponent(componentExists bool, ei envinfo.EnvSpecificInfo, isMainStorageEphemeral bool) (err error) {
	ei.SetDevfileObj(a.Devfile)
	componentName := a.ComponentName

	storageClient := storagepkg.NewClient(storagepkg.ClientOptions{
		Client:              a.Client,
		LocalConfigProvider: &ei,
	})

	// handle the ephemeral storage
	err = storage.HandleEphemeralStorage(a.Client, storageClient, componentName, isMainStorageEphemeral)
	if err != nil {
		return err
	}

	// From devfile info, create PVCs and return ephemeral storages
	ephemerals, err := storagepkg.Push(storageClient, &ei)
	if err != nil {
		return err
	}

	// Set the labels
	labels := componentlabels.GetLabels(componentName, a.AppName, true)
	labels[componentlabels.OdoModeLabel] = componentlabels.ComponentDevName
	labels["component"] = componentName

	annotations := make(map[string]string)
	annotations[componentlabels.OdoProjectTypeAnnotation] = component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata())
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

	containers, err = utils.UpdateContainersWithSupervisord(a.Devfile, containers, a.devfileRunCmd, a.devfileDebugCmd, a.devfileDebugPort)
	if err != nil {
		return err
	}

	initContainers, err := generator.GetInitContainers(a.Devfile)
	if err != nil {
		return err
	}

	initContainers = append(initContainers, kclient.GetBootstrapSupervisordInitContainer())

	// list all the pvcs for the component
	pvcs, err := a.Client.ListPVCs(fmt.Sprintf("%v=%v", "component", componentName))
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
	deployParams.InitContainers[0].ImagePullPolicy = corev1.PullIfNotPresent
	deployParams.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
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
	serviceObjectMeta := generator.GetObjectMeta(serviceName, a.Client.GetCurrentNamespace(), labels, serviceAnnotations)
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
		if a.Client.IsSSASupported() {
			klog.V(4).Info("Applying deployment")
			a.deployment, err = a.Client.ApplyDeployment(*deployment)
		} else {
			klog.V(4).Info("Updating deployment")
			a.deployment, err = a.Client.UpdateDeployment(*deployment)
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
		if a.Client.IsSSASupported() {
			a.deployment, err = a.Client.ApplyDeployment(*deployment)
		} else {
			a.deployment, err = a.Client.CreateDeployment(*deployment)
		}

		if err != nil {
			return err
		}

		klog.V(2).Infof("Successfully created component %v", componentName)
		ownerReference := generator.GetOwnerReference(a.deployment)
		svc.OwnerReferences = append(svc.OwnerReferences, ownerReference)
		if len(svc.Spec.Ports) > 0 {
			_, err = a.Client.CreateService(*svc)
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}

	return nil
}

func (a *Adapter) createOrUpdateServiceForComponent(svc *corev1.Service, componentName string) error {
	oldSvc, err := a.Client.GetOneService(a.ComponentName, a.AppName)
	ownerReference := generator.GetOwnerReference(a.deployment)
	svc.OwnerReferences = append(svc.OwnerReferences, ownerReference)
	if err != nil {
		// no old service was found, create a new one
		if len(svc.Spec.Ports) > 0 {
			_, err = a.Client.CreateService(*svc)
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
		_, err = a.Client.UpdateService(*svc)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully update Service for component %s", componentName)
		return nil
	}
	// delete the old existing service if the component currently doesn't expose any ports
	return a.Client.DeleteService(oldSvc.Name)
}

// generateDeploymentObjectMeta generates a ObjectMeta object for the given deployment's name, labels and annotations
// if no deployment exists, it creates a new deployment name
func (a Adapter) generateDeploymentObjectMeta(labels map[string]string, annotations map[string]string) (metav1.ObjectMeta, error) {
	if a.deployment != nil {
		return generator.GetObjectMeta(a.deployment.Name, a.Client.GetCurrentNamespace(), labels, annotations), nil
	} else {
		deploymentName, err := util.NamespaceKubernetesObject(a.ComponentName, a.AppName)
		if err != nil {
			return metav1.ObjectMeta{}, err
		}
		return generator.GetObjectMeta(deploymentName, a.Client.GetCurrentNamespace(), labels, annotations), nil
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

func (a Adapter) ExecCMDInContainer(componentInfo common.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return a.Client.ExecCMDInContainer(componentInfo.ContainerName, componentInfo.PodName, cmd, stdout, stderr, stdin, tty)
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (a Adapter) ExtractProjectToComponent(componentInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {
	return a.Client.ExtractProjectToComponent(componentInfo.ContainerName, componentInfo.PodName, targetPath, stdin)
}
