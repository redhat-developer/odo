package component

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/envinfo"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"k8s.io/klog"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/utils"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/sync"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	DeployComponentSuffix = "-deploy"
	BuildTimeout          = 5 * time.Minute
)

// New instantiantes a component adapter
func New(adapterContext common.AdapterContext, client kclient.Client) Adapter {
	adapter := Adapter{Client: client}
	adapter.GenericAdapter = common.NewGenericAdapter(&client, adapterContext)
	adapter.GenericAdapter.InitWith(adapter)
	return adapter
}

// getPod lazily records and retrieves the pod associated with the component associated with this adapter. If refresh parameter
// is true, then the pod is refreshed from the cluster regardless of its current local state
func (a Adapter) getPod(refresh bool) (*corev1.Pod, error) {
	if refresh || a.pod == nil {
		podSelector := fmt.Sprintf("component=%s", a.ComponentName)
		watchOptions := metav1.ListOptions{
			LabelSelector: podSelector,
		}
		// Wait for Pod to be in running state otherwise we can't sync data to it.
		pod, err := a.Client.WaitAndGetPod(watchOptions, corev1.PodRunning, "Waiting for component to start", true)
		if err != nil {
			return nil, errors.Wrapf(err, "error while waiting for pod %s", podSelector)
		}
		a.pod = pod
	}
	return a.pod, nil
}

func (a Adapter) ComponentInfo(command versionsCommon.DevfileCommand) (common.ComponentInfo, error) {
	pod, err := a.getPod(false)
	if err != nil {
		return common.ComponentInfo{}, err
	}
	return common.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: command.Exec.Component,
	}, nil
}

func (a Adapter) SupervisorComponentInfo(command versionsCommon.DevfileCommand) (common.ComponentInfo, error) {
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
	Client kclient.Client
	*common.GenericAdapter

	devfileInitCmd   string
	devfileBuildCmd  string
	devfileRunCmd    string
	devfileDebugCmd  string
	devfileDebugPort int
	pod              *corev1.Pod
}

const dockerfilePath string = "Dockerfile"

func (a Adapter) runBuildConfig(client *occlient.Client, parameters common.BuildParameters) (err error) {
	buildName := a.ComponentName

	commonObjectMeta := metav1.ObjectMeta{
		Name: buildName,
	}

	buildOutput := "DockerImage"

	if parameters.Tag == "" {
		parameters.Tag = fmt.Sprintf("%s:latest", buildName)
		buildOutput = "ImageStreamTag"
	}

	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt, syscall.SIGTERM)
	go a.terminateBuild(controlC, client, commonObjectMeta)

	_, err = client.CreateDockerBuildConfigWithBinaryInput(commonObjectMeta, dockerfilePath, parameters.Tag, []corev1.EnvVar{}, buildOutput)
	if err != nil {
		return err
	}

	defer func() {
		// This will delete both the BuildConfig and any builds using that BuildConfig
		derr := client.DeleteBuildConfig(commonObjectMeta)
		if err == nil {
			err = derr
		}
	}()

	syncAdapter := sync.New(a.AdapterContext, &a.Client)
	reader, err := syncAdapter.SyncFilesBuild(parameters, dockerfilePath)
	if err != nil {
		return err
	}

	bc, err := client.RunBuildConfigWithBinaryInput(buildName, reader)
	if err != nil {
		return err
	}
	log.Successf("Started build %s using BuildConfig", bc.Name)

	reader, writer := io.Pipe()

	var cmdOutput string
	// This Go routine will automatically pipe the output from WaitForBuildToFinish to
	// our logger.
	// We pass the controlC os.Signal in order to output the logs within the terminateBuild
	// function if the process is interrupted by the user performing a ^C. If we didn't pass it
	// The Scanner would consume the log, and only output it if there was an err within this
	// func.
	go func(controlC chan os.Signal) {
		select {
		case <-controlC:
			return
		default:
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()

				if log.IsDebug() {
					_, err := fmt.Fprintln(os.Stdout, line)
					if err != nil {
						log.Errorf("Unable to print to stdout: %v", err)
					}
				}

				cmdOutput += fmt.Sprintln(line)
			}
		}
	}(controlC)

	s := log.Spinner("Waiting for build to complete")
	if err := client.WaitForBuildToFinish(bc.Name, writer, BuildTimeout); err != nil {
		s.End(false)
		return errors.Wrapf(err, "unable to build image using BuildConfig %s, error: %s", buildName, cmdOutput)
	}

	s.End(true)
	// Stop listening for a ^C so it doesnt perform terminateBuild during any later stages
	signal.Stop(controlC)
	log.Successf("Successfully built container image: %s", parameters.Tag)
	return
}

// terminateBuild is triggered if the user performs a ^C action within the terminal during the build phase
// of the deploy.
// It cleans up the resources created for the build, as the defer function would not be reached.
// The subsequent deploy would fail if these resources are not cleaned up.
func (a Adapter) terminateBuild(c chan os.Signal, client *occlient.Client, commonObjectMeta metav1.ObjectMeta) {
	<-c

	log.Info("\nBuild process interrupted, terminating build, this might take a few seconds")
	err := client.DeleteBuildConfig(commonObjectMeta)
	if err != nil {
		log.Info("\n", err.Error())
	}
	os.Exit(0)
}

// Build image for devfile project
func (a Adapter) Build(parameters common.BuildParameters) (err error) {
	// TODO: set namespace from user flag
	client, err := occlient.New()
	if err != nil {
		return err
	}

	isBuildConfigSupported, err := client.IsBuildConfigSupported()
	if err != nil {
		return err
	}

	if isBuildConfigSupported {
		return a.runBuildConfig(client, parameters)
	}

	return errors.New("unable to build image, only Openshift BuildConfig build is supported")
}

// Perform the substitutions in the manifest file(s)
func substitueYamlVariables(baseYaml []byte, yamlSubstitutions map[string]string) ([]byte, error) {
	// create new template from parsing file
	tmpl, err := template.New("deploy").Parse(string(baseYaml))
	if err != nil {
		return []byte{}, errors.Wrap(err, "error creating template")
	}

	// define a buffer to store the results
	var buf bytes.Buffer

	// apply template to yaml file
	_ = tmpl.Execute(&buf, yamlSubstitutions)

	return buf.Bytes(), nil
}

// Build image for devfile project
func (a Adapter) Deploy(parameters common.DeployParameters) (err error) {
	// TODO: Can we use a occlient created somewhere else rather than create another
	client, err := occlient.New()
	if err != nil {
		return err
	}

	namespace := a.Client.Namespace
	applicationName := a.ComponentName + DeployComponentSuffix
	deploymentManifest := &unstructured.Unstructured{}

	var imageStream *imagev1.ImageStream
	if parameters.Tag == "" {
		imageStream, err = client.GetImageStream(namespace, a.ComponentName, "latest")
		if err != nil {
			return err
		}

		imageStreamImage, err := client.GetImageStreamImage(imageStream, "latest")
		if err != nil {
			return err
		}
		parameters.Tag = imageStreamImage.Image.DockerImageReference
	}

	// Specify the substitution keys and values
	yamlSubstitutions := map[string]string{
		"CONTAINER_IMAGE": parameters.Tag,
		"COMPONENT_NAME":  applicationName,
		"PORT":            strconv.Itoa(parameters.DeploymentPort),
	}

	// Build a yaml decoder with the unstructured Scheme
	yamlDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// This will override if manifest.yaml is present
	writtenToManifest := false
	manifestFile, err := os.Create(filepath.Join(a.Context, ".odo", "manifest.yaml"))
	if err != nil {
		err = manifestFile.Close()
		return errors.Wrap(err, "Unable to create the local manifest file")
	}

	defer func() {
		merr := manifestFile.Close()
		if err == nil {
			err = merr
		}
	}()

	manifests := bytes.Split(parameters.ManifestSource, []byte("---"))
	for _, manifest := range manifests {
		if len(manifest) > 0 {
			// Substitute the values in the manifest file
			deployYaml, err := substitueYamlVariables(manifest, yamlSubstitutions)
			if err != nil {
				return errors.Wrap(err, "unable to substitute variables in manifest")
			}

			_, gvk, err := yamlDecoder.Decode([]byte(deployYaml), nil, deploymentManifest)
			if err != nil {
				return errors.New("Failed to decode the manifest yaml")
			}

			kind := utils.PluraliseKind(gvk.Kind)
			gvr := schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: kind}
			klog.V(3).Infof("Manifest type: %s", gvr.String())

			labels := map[string]string{
				"component": applicationName,
			}

			manifestLabels := deploymentManifest.GetLabels()
			if manifestLabels != nil {
				for key, value := range labels {
					manifestLabels[key] = value
				}
				deploymentManifest.SetLabels(manifestLabels)
			} else {
				deploymentManifest.SetLabels(labels)
			}

			// Check to see whether deployed resource already exists. If not, create else update
			instanceFound := false
			item, err := a.Client.DynamicClient.Resource(gvr).Namespace(namespace).Get(deploymentManifest.GetName(), metav1.GetOptions{})
			if item != nil && err == nil {
				instanceFound = true
				deploymentManifest.SetResourceVersion(item.GetResourceVersion())
				deploymentManifest.SetAnnotations(item.GetAnnotations())
				// If deployment is a `Service` of type `ClusterIP` then the service in the manifest will probably not
				// have a ClusterIP defined, as this is determined when the manifest is applied. When updating the Service
				// the manifest cannot have an empty `ClusterIP` defintion, so we need to copy this from the existing definition.
				if item.GetKind() == "Service" {
					currentServiceSpec := item.UnstructuredContent()["spec"].(map[string]interface{})
					if currentServiceSpec["clusterIP"] != nil && currentServiceSpec["clusterIP"] != "" {
						newService := deploymentManifest.UnstructuredContent()
						newService["spec"].(map[string]interface{})["clusterIP"] = currentServiceSpec["clusterIP"]
						deploymentManifest.SetUnstructuredContent(newService)
					}
				}
			}

			actionType := "Creating"
			if instanceFound {
				actionType = "Updating" // Update deployment
			}
			s := log.Spinnerf("%s resource of kind %s", strings.Title(actionType), gvk.Kind)
			var result *unstructured.Unstructured
			if !instanceFound {
				result, err = a.Client.DynamicClient.Resource(gvr).Namespace(namespace).Create(deploymentManifest, metav1.CreateOptions{})
			} else {
				result, err = a.Client.DynamicClient.Resource(gvr).Namespace(namespace).Update(deploymentManifest, metav1.UpdateOptions{})
			}
			if err != nil {
				s.End(false)
				return errors.Wrapf(err, "Failed when %s manifest %s", actionType, gvk.Kind)
			}
			s.End(true)

			if imageStream != nil {
				ownerReference := metav1.OwnerReference{
					APIVersion: result.GetAPIVersion(),
					Kind:       result.GetKind(),
					Name:       result.GetName(),
					UID:        result.GetUID(),
				}

				imageStream.ObjectMeta.OwnerReferences = append(imageStream.ObjectMeta.OwnerReferences, ownerReference)
			}

			// Write the returned manifest to the local manifest file
			if writtenToManifest {
				_, err = manifestFile.WriteString("---\n")
				if err != nil {
					return errors.Wrap(err, "Unable to write to local manifest file")
				}
			}
			err = yamlDecoder.Encode(result, manifestFile)
			if err != nil {
				return errors.Wrap(err, "Unable to write to local manifest file")
			}
			writtenToManifest = true
		}
	}

	if imageStream != nil {
		err = client.UpdateImageStream(imageStream)
		if err != nil {
			return err
		}
	}
	s := log.Spinner("Determining the application URL")

	// Need to wait for a second to give the server time to create the artifacts
	// TODO: Replace wait with a wait for object to be created correctly
	time.Sleep(2 * time.Second)

	labelSelector := fmt.Sprintf("%v=%v", "component", applicationName)
	fullURL, err := client.GetApplicationURL(applicationName, labelSelector)
	if err != nil {
		s.End(false)
		log.Warningf("Unable to determine the application URL for component %s: %s", a.ComponentName, err)
	} else {
		s.End(true)
		log.Successf("Successfully deployed component: %s", fullURL)
	}

	return nil
}

func (a Adapter) DeployDelete(manifest []byte) (err error) {
	deploymentManifest := &unstructured.Unstructured{}
	// Build a yaml decoder with the unstructured Scheme
	yamlDecoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	manifests := bytes.Split(manifest, []byte("---"))
	for _, splitManifest := range manifests {
		if len(manifest) > 0 {
			_, gvk, err := yamlDecoder.Decode([]byte(splitManifest), nil, deploymentManifest)
			if err != nil {
				return err
			}
			klog.V(3).Infof("Deploy manifest:\n\n%s", deploymentManifest)
			kind := utils.PluraliseKind(gvk.Kind)
			gvr := schema.GroupVersionResource{Group: gvk.Group, Version: gvk.Version, Resource: kind}
			klog.V(3).Infof("Manifest type: %s", gvr.String())

			_, err = a.Client.DynamicClient.Resource(gvr).Namespace(a.Client.Namespace).Get(deploymentManifest.GetName(), metav1.GetOptions{})
			if err != nil {
				errorMessage := "Could not delete component " + deploymentManifest.GetName() + " as component was not found"
				return errors.New(errorMessage)
			}

			err = a.Client.DynamicClient.Resource(gvr).Namespace(a.Client.Namespace).Delete(deploymentManifest.GetName(), &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
// Once the component has started, it will sync the source code to it.
func (a Adapter) Push(parameters common.PushParameters) (err error) {
	componentExists, err := utils.ComponentExists(a.Client, a.ComponentName)
	if err != nil {
		return errors.Wrapf(err, "unable to determine if component %s exists", a.ComponentName)
	}

	a.devfileInitCmd = parameters.DevfileInitCmd
	a.devfileBuildCmd = parameters.DevfileBuildCmd
	a.devfileRunCmd = parameters.DevfileRunCmd
	a.devfileDebugCmd = parameters.DevfileDebugCmd
	a.devfileDebugPort = parameters.DebugPort

	podChanged := false
	var podName string

	// If the component already exists, retrieve the pod's name before it's potentially updated
	if componentExists {
		pod, err := a.getPod(true)
		if err != nil {
			return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
		}
		podName = pod.GetName()
	}

	// Validate the devfile build and run commands
	log.Info("\nValidation")
	s := log.Spinner("Validating the devfile")
	pushDevfileCommands, err := common.ValidateAndGetPushDevfileCommands(a.Devfile.Data, a.devfileInitCmd, a.devfileBuildCmd, a.devfileRunCmd)
	if err != nil {
		s.End(false)
		return errors.Wrap(err, "failed to validate devfile build and run commands")
	}
	s.End(true)

	log.Infof("\nCreating Kubernetes resources for component %s", a.ComponentName)

	previousMode := parameters.EnvSpecificInfo.GetRunMode()
	currentMode := envinfo.Run

	if parameters.Debug {
		pushDevfileDebugCommands, err := common.ValidateAndGetDebugDevfileCommands(a.Devfile.Data, a.devfileDebugCmd)
		if err != nil {
			return fmt.Errorf("debug command is not valid")
		}
		pushDevfileCommands[versionsCommon.DebugCommandGroupType] = pushDevfileDebugCommands
		currentMode = envinfo.Debug
	}

	if currentMode != previousMode {
		parameters.RunModeChanged = true
	}
	containerComponents := common.GetDevfileContainerComponents(a.Devfile.Data)
	portExposureMap := utils.GetPortExposure(containerComponents)

	err = a.createOrUpdateComponent(componentExists, parameters.EnvSpecificInfo, portExposureMap)
	if err != nil {
		return errors.Wrap(err, "unable to create or update component")
	}

	_, err = a.Client.WaitForDeploymentRollout(a.ComponentName)
	if err != nil {
		return errors.Wrap(err, "error while waiting for deployment rollout")
	}

	// Wait for Pod to be in running state otherwise we can't sync data or exec commands to it.
	pod, err := a.getPod(true)
	if err != nil {
		return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
	}

	err = component.ApplyConfig(nil, &a.Client, config.LocalConfigInfo{}, parameters.EnvSpecificInfo, color.Output, componentExists, containerComponents, false)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed.")
	}

	// Compare the name of the pod with the one before the rollout. If they differ, it means there's a new pod and a force push is required
	if componentExists && podName != pod.GetName() {
		podChanged = true
	}

	// Find at least one pod with the source volume mounted, error out if none can be found
	containerName, sourceMount, err := getFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return errors.Wrapf(err, "error while retrieving container from pod %s with a mounted project volume", podName)
	}

	log.Infof("\nSyncing to component %s", a.ComponentName)
	// Get a sync adapter. Check if project files have changed and sync accordingly
	syncAdapter := sync.New(a.AdapterContext, &a.Client)
	compInfo := common.ComponentInfo{
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SourceMount:   sourceMount,
	}
	syncParams := common.SyncParameters{
		PushParams:      parameters,
		CompInfo:        compInfo,
		ComponentExists: componentExists,
		PodChanged:      podChanged,
	}
	execRequired, err := syncAdapter.SyncFiles(syncParams)
	if err != nil {
		return errors.Wrapf(err, "Failed to sync to component with name %s", a.ComponentName)
	}

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	postStartEvents := a.Devfile.Data.GetEvents().PostStart
	if !componentExists && len(postStartEvents) > 0 {
		err = a.ExecDevfileEvent(postStartEvents, common.PostStart, parameters.Show)
		if err != nil {
			return err

		}
	}

	if execRequired || parameters.RunModeChanged {
		log.Infof("\nExecuting devfile commands for component %s", a.ComponentName)
		err = a.ExecDevfile(pushDevfileCommands, componentExists, parameters)
		if err != nil {
			return err
		}
	}

	return nil
}

// Test runs the devfile test command
func (a Adapter) Test(testCmd string, show bool) (err error) {
	pod, err := a.Client.GetPodUsingComponentName(a.ComponentName)
	if err != nil {
		return fmt.Errorf("error occurred while getting the pod: %w", err)
	}
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod for component %s is not running", a.ComponentName)
	}

	log.Infof("\nExecuting devfile test command for component %s", a.ComponentName)

	testCommand, err := common.ValidateAndGetTestDevfileCommands(a.Devfile.Data, testCmd)
	if err != nil {
		return errors.Wrap(err, "failed to validate devfile test command")
	}
	err = a.ExecuteDevfileCommand(testCommand, show)
	if err != nil {
		return errors.Wrapf(err, "failed to execute devfile commands for component %s", a.ComponentName)
	}
	return nil
}

// DoesComponentExist returns true if a component with the specified name exists, false otherwise
func (a Adapter) DoesComponentExist(cmpName string) (bool, error) {
	return utils.ComponentExists(a.Client, cmpName)
}

func (a Adapter) createOrUpdateComponent(componentExists bool, ei envinfo.EnvSpecificInfo, portExposureMap map[int32]versionsCommon.ExposureType) (err error) {
	componentName := a.ComponentName

	componentType := strings.TrimSuffix(a.AdapterContext.Devfile.Data.GetMetadata().Name, "-")

	labels := componentlabels.GetLabels(componentName, a.AppName, true)
	labels["component"] = componentName
	labels[componentlabels.ComponentTypeLabel] = componentType

	containers, err := utils.GetContainers(a.Devfile)
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	containers, err = utils.UpdateContainersWithSupervisord(a.Devfile, containers, a.devfileRunCmd, a.devfileDebugCmd, a.devfileDebugPort)
	if err != nil {
		return err
	}

	// set EnvFrom to the container that's supposed to have link to the Operator backed service
	containers, err = utils.UpdateContainerWithEnvFrom(containers, a.Devfile, a.devfileRunCmd, ei)
	if err != nil {
		return err
	}

	objectMeta := kclient.CreateObjectMeta(componentName, a.Client.Namespace, labels, nil)
	podTemplateSpec := kclient.GeneratePodTemplateSpec(objectMeta, containers)

	kclient.AddBootstrapSupervisordInitContainer(podTemplateSpec)

	// if there are preStart events, add them as init containers to the podTemplateSpec
	preStartEvents := a.Devfile.Data.GetEvents().PreStart
	if len(preStartEvents) > 0 {
		var eventCommands []string
		commandsMap := a.Devfile.Data.GetCommands()
		containersMap := utils.GetContainersMap(containers)

		for _, event := range preStartEvents {
			eventSubCommands := common.GetCommandsFromEvent(commandsMap, strings.ToLower(event))
			eventCommands = append(eventCommands, eventSubCommands...)
		}

		klog.V(4).Infof("PreStart event commands are: %v", strings.Join(eventCommands, ","))
		utils.AddPreStartEventInitContainer(podTemplateSpec, commandsMap, eventCommands, containersMap)
		if len(eventCommands) > 0 {
			log.Successf("PreStart commands have been added to the component: %s", strings.Join(eventCommands, ","))
		}
	}

	containerNameToVolumes := common.GetVolumes(a.Devfile)

	var uniqueStorages []common.Storage
	volumeNameToPVCName := make(map[string]string)
	processedVolumes := make(map[string]bool)

	// Get a list of all the unique volume names and generate their PVC names
	// we do not use the volume components which are unique here because
	// not all volume components maybe referenced by a container component.
	// We only want to create PVCs which are going to be used by a container
	for _, volumes := range containerNameToVolumes {
		for _, vol := range volumes {
			if _, ok := processedVolumes[vol.Name]; !ok {
				processedVolumes[vol.Name] = true

				// Generate the PVC Names
				klog.V(2).Infof("Generating PVC name for %v", vol.Name)
				generatedPVCName, err := storage.GeneratePVCNameFromDevfileVol(vol.Name, componentName)
				if err != nil {
					return err
				}

				// Check if we have an existing PVC with the labels, overwrite the generated name with the existing name if present
				existingPVCName, err := storage.GetExistingPVC(&a.Client, vol.Name, componentName)
				if err != nil {
					return err
				}
				if len(existingPVCName) > 0 {
					klog.V(2).Infof("Found an existing PVC for %v, PVC %v will be re-used", vol.Name, existingPVCName)
					generatedPVCName = existingPVCName
				}

				pvc := common.Storage{
					Name:   generatedPVCName,
					Volume: vol,
				}
				uniqueStorages = append(uniqueStorages, pvc)
				volumeNameToPVCName[vol.Name] = generatedPVCName
			}
		}
	}

	err = storage.DeleteOldPVCs(&a.Client, componentName, processedVolumes)
	if err != nil {
		return err
	}

	// Add PVC and Volume Mounts to the podTemplateSpec
	err = kclient.AddPVCAndVolumeMount(podTemplateSpec, volumeNameToPVCName, containerNameToVolumes)
	if err != nil {
		return err
	}

	deploymentSpec := kclient.GenerateDeploymentSpec(*podTemplateSpec, map[string]string{
		"component": componentName,
	})

	var containerPorts []corev1.ContainerPort

	for _, c := range deploymentSpec.Template.Spec.Containers {
		// No need to check
		if reflect.DeepEqual(a.Devfile.Ctx.GetApiVersion(), "1.0.0") {
			containerPorts = append(containerPorts, c.Ports...)
		} else {
			for _, port := range c.Ports {
				portExist := false
				for _, entry := range containerPorts {
					if entry.ContainerPort == port.ContainerPort {
						portExist = true
						break
					}
				}
				// if Exposure == none, should not create a service for that port
				if !portExist && portExposureMap[port.ContainerPort] != versionsCommon.None {
					containerPorts = append(containerPorts, port)
				}
			}
		}
	}

	serviceSpec := kclient.GenerateServiceSpec(objectMeta.Name, containerPorts)
	klog.V(2).Infof("Creating deployment %v", deploymentSpec.Template.GetName())
	klog.V(2).Infof("The component name is %v", componentName)

	if componentExists {
		// If the component already exists, get the resource version of the deploy before updating
		klog.V(2).Info("The component already exists, attempting to update it")
		deployment, err := a.Client.UpdateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully updated component %v", componentName)
		oldSvc, err := a.Client.KubeClient.CoreV1().Services(a.Client.Namespace).Get(componentName, metav1.GetOptions{})
		objectMetaTemp := objectMeta
		ownerReference := kclient.GenerateOwnerReference(deployment)
		objectMetaTemp.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)
		if err != nil {
			// no old service was found, create a new one
			if len(serviceSpec.Ports) > 0 {
				_, err = a.Client.CreateService(objectMetaTemp, *serviceSpec)
				if err != nil {
					return err
				}
				klog.V(2).Infof("Successfully created Service for component %s", componentName)
			}
		} else {
			if len(serviceSpec.Ports) > 0 {
				serviceSpec.ClusterIP = oldSvc.Spec.ClusterIP
				objectMetaTemp.ResourceVersion = oldSvc.GetResourceVersion()
				_, err = a.Client.UpdateService(objectMetaTemp, *serviceSpec)
				if err != nil {
					return err
				}
				klog.V(2).Infof("Successfully update Service for component %s", componentName)
			} else {
				err = a.Client.KubeClient.CoreV1().Services(a.Client.Namespace).Delete(componentName, &metav1.DeleteOptions{})
				if err != nil {
					return err
				}
			}
		}
	} else {
		deployment, err := a.Client.CreateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		klog.V(2).Infof("Successfully created component %v", componentName)
		ownerReference := kclient.GenerateOwnerReference(deployment)
		objectMetaTemp := objectMeta
		objectMetaTemp.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)
		if len(serviceSpec.Ports) > 0 {
			_, err = a.Client.CreateService(objectMetaTemp, *serviceSpec)
			if err != nil {
				return err
			}
			klog.V(2).Infof("Successfully created Service for component %s", componentName)
		}

	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(uniqueStorages)
	if err != nil {
		return err
	}

	return nil
}

// getFirstContainerWithSourceVolume returns the first container that set mountSources: true as well
// as the path to the source volume inside the container.
// Because the source volume is shared across all components that need it, we only need to sync once,
// so we only need to find one container. If no container was found, that means there's no
// container to sync to, so return an error
func getFirstContainerWithSourceVolume(containers []corev1.Container) (string, string, error) {
	for _, c := range containers {
		for _, vol := range c.VolumeMounts {
			if vol.Name == kclient.OdoSourceVolume {
				return c.Name, vol.MountPath, nil
			}
		}
	}

	return "", "", fmt.Errorf("In order to sync files, odo requires at least one component in a devfile to set 'mountSources: true'")
}

// Delete deletes the component
func (a Adapter) Delete(labels map[string]string, show bool) error {

	log.Infof("\nGathering information for component %s", a.ComponentName)
	podSpinner := log.Spinner("Checking status for component")
	defer podSpinner.End(false)

	pod, err := a.Client.GetPodUsingComponentName(a.ComponentName)
	if kerrors.IsForbidden(err) {
		klog.V(2).Infof("Resource for %s forbidden", a.ComponentName)
		// log the error if it failed to determine if the component exists due to insufficient RBACs
		podSpinner.End(false)
		log.Warningf("%v", err)
		return nil
	} else if e, ok := err.(*kclient.PodNotFoundError); ok {
		podSpinner.End(false)
		log.Warningf("%v", e)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "unable to determine if component %s exists", a.ComponentName)
	}

	podSpinner.End(true)

	// if there are preStop events, execute them before deleting the deployment
	preStopEvents := a.Devfile.Data.GetEvents().PreStop
	if len(preStopEvents) > 0 {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("unable to execute preStop events, pod for component %s is not running", a.ComponentName)
		}

		err = a.ExecDevfileEvent(preStopEvents, common.PreStop, show)
		if err != nil {
			return err
		}
	}

	log.Infof("\nDeleting component %s", a.ComponentName)
	spinner := log.Spinner("Deleting Kubernetes resources for component")
	defer spinner.End(false)

	err = a.Client.DeleteDeployment(labels)
	if err != nil {
		return err
	}

	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
}

// Log returns log from component
func (a Adapter) Log(follow, debug bool) (io.ReadCloser, error) {

	pod, err := a.Client.GetPodUsingComponentName(a.ComponentName)
	if err != nil {
		return nil, errors.Errorf("the component %s doesn't exist on the cluster", a.ComponentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, errors.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	var command versionsCommon.DevfileCommand
	if debug {
		command, err = common.GetDebugCommand(a.Devfile.Data, "")
		if err != nil {
			return nil, err
		}
		if reflect.DeepEqual(versionsCommon.DevfileCommand{}, command) {
			return nil, errors.Errorf("no debug command found in devfile, please run \"odo log\" for run command logs")
		}

	} else {
		command, err = common.GetRunCommand(a.Devfile.Data, "")
		if err != nil {
			return nil, err
		}
	}

	containerName := command.Exec.Component

	return a.Client.GetPodLogs(pod.Name, containerName, follow)
}

// Exec executes a command in the component
func (a Adapter) Exec(command []string) error {
	exists, err := utils.ComponentExists(a.Client, a.ComponentName)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("the component %s doesn't exist on the cluster", a.ComponentName)
	}

	runCommand, err := common.GetRunCommand(a.Devfile.Data, "")
	if err != nil {
		return err
	}
	containerName := runCommand.Exec.Component

	// get the pod
	pod, err := a.Client.GetPodUsingComponentName(a.ComponentName)
	if err != nil {
		return errors.Wrapf(err, "unable to get pod for component %s", a.ComponentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to exec as the component is not running. Current status=%v", pod.Status.Phase)
	}

	componentInfo := common.ComponentInfo{
		PodName:       pod.Name,
		ContainerName: containerName,
	}

	return a.ExecuteCommand(componentInfo, command, true, nil, nil)
}
