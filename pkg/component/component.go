package component

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"github.com/redhat-developer/odo/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// componentSourceURLAnnotation is an source url from which component was build
// it can be also file://
const componentSourceURLAnnotation = "app.kubernetes.io/url"
const componentSourceTypeAnnotation = "app.kubernetes.io/component-source-type"
const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1

// ComponentInfo holds all important information about one component
type ComponentInfo struct {
	Name string
	Type string
}

// GetDefaultComponentName generates a unique component name
// Parameters: desired default component name(w/o prefix) and slice of existing component names
// Returns: Unique component name and error if any
func GetDefaultComponentName(componentPath string, componentPathType util.ComponentCreateType, componentType string, existingComponentList []ComponentInfo) (string, error) {
	var prefix string

	// Get component names from component list
	var existingComponentNames []string
	for _, componentInfo := range existingComponentList {
		existingComponentNames = append(existingComponentNames, componentInfo.Name)
	}

	// Fetch config
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random component name")
	}

	// If there's no prefix in config file, or its value is config.ConfigPrefixDir use safe default - the current directory along with component type
	if cfg.OdoSettings.Prefix == nil || *cfg.OdoSettings.Prefix == config.ConfigPrefixDir {
		prefix, err = util.GetComponentDir(componentPath, componentPathType)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random component name")
		}
		prefix = util.TruncateString(prefix, componentRandomNamePartsMaxLen)
	} else {
		// Set the required prefix into componentName
		prefix = *cfg.OdoSettings.Prefix
	}

	// Generate unique name for the component using prefix and unique random suffix
	componentName, err := util.GetRandomName(
		fmt.Sprintf("%s-%s", prefix, componentType),
		componentNameMaxLen,
		existingComponentNames,
		componentNameMaxRetries,
	)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random component name")
	}

	return componentName, nil
}

// validateSourceType check if given sourceType is supported
func validateSourceType(sourceType string) bool {
	validSourceTypes := []string{
		"git",
		"local",
		"binary",
	}
	for _, valid := range validSourceTypes {
		if valid == sourceType {
			return true
		}
	}
	return false
}

// CreateFromGit inputPorts is the array containing the string port values
// inputPorts is the array containing the string port values
// envVars is the array containing the environment variables
func CreateFromGit(client *occlient.Client, name string, componentImageType string, url string, applicationName string, inputPorts []string, envVars []string) error {

	labels := componentlabels.GetLabels(name, applicationName, true)

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(componentImageType)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// save source path as annotation
	annotations := map[string]string{componentSourceURLAnnotation: url}
	annotations[componentSourceTypeAnnotation] = "git"

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(name, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        namespacedOpenShiftObject,
		Labels:      labels,
		Annotations: annotations,
	}

	err = client.NewAppS2I(commonObjectMeta, componentImageType, url, inputPorts, envVars)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", namespacedOpenShiftObject)
	}
	return nil
}

// GetComponentPorts provides slice of ports used by the component in the form port_no/protocol
func GetComponentPorts(client *occlient.Client, componentName string, applicationName string) (ports []string, err error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	dc, err := client.GetOneDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch deployment configs for the selector %v", componentSelector)
	}

	for _, container := range dc.Spec.Template.Spec.Containers {
		for _, port := range container.Ports {
			ports = append(ports, fmt.Sprintf("%v/%v", port.ContainerPort, port.Protocol))
		}
	}

	return ports, nil
}

// CreateFromPath create new component with source or binary from the given local path
// sourceType indicates the source type of the component and can be either local or binary
// envVars is the array containing the environment variables
func CreateFromPath(client *occlient.Client, name string, componentImageType string, path string, applicationName string, sourceType string, inputPorts []string, envVars []string) error {
	labels := componentlabels.GetLabels(name, applicationName, true)

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(componentImageType)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// save source path as annotation
	sourceURL := util.GenFileUrl(path, runtime.GOOS)
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL}
	annotations[componentSourceTypeAnnotation] = sourceType

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(name, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        namespacedOpenShiftObject,
		Labels:      labels,
		Annotations: annotations,
	}

	// Bootstrap the deployment with SupervisorD
	err = client.BootstrapSupervisoredS2I(commonObjectMeta, componentImageType, inputPorts, envVars)
	if err != nil {
		return err
	}

	return nil
}

// Delete whole component
func Delete(client *occlient.Client, name string, applicationName string, projectName string) error {

	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to create new configuration to delete %s", name)
	}

	labels := componentlabels.GetLabels(name, applicationName, false)

	err = client.Delete(labels)
	if err != nil {
		return errors.Wrapf(err, "error deleting component %s", name)
	}

	// Get a list of all components
	components, err := List(client, applicationName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve list of components")
	}

	// First check is that we want to only update the active component if the component which is getting deleted is the
	// active component
	// Second check is that we want to do an update only if it is happening for the active application otherwise we need
	// not to care for the update of the active component
	activeComponent := cfg.GetActiveComponent(applicationName, projectName)
	activeApplication := cfg.GetActiveApplication(projectName)
	if activeComponent == name && activeApplication == applicationName {

		// If there's more than one component, set it to the first one..
		if len(components) > 0 {
			err = cfg.SetActiveComponent(components[0].Name, applicationName, projectName)

			if err != nil {
				return errors.Wrapf(err, "unable to set current component to '%s'", name)
			}
		} else {
			// Unset to blank
			err = cfg.UnsetActiveComponent(projectName)
			if err != nil {
				return errors.Wrapf(err, "error unsetting current component while deleting %s", name)
			}
		}
	}

	return nil
}

// SetCurrent sets the given component to active in odo config file
func SetCurrent(client *occlient.Client, name string, applicationName string, projectName string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	err = cfg.SetActiveComponent(name, applicationName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", name)
	}

	return nil
}

// GetCurrent component in active application
// returns "" if there is no active component
func GetCurrent(client *occlient.Client, applicationName string, projectName string) (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentComponent := cfg.GetActiveComponent(applicationName, projectName)

	return currentComponent, nil
}

// PushLocal push local code to the cluster and trigger build there.
// files is list of changed files captured during `odo watch` as well as binary file path
// During copying binary components, path represent base directory path to binary and files contains path of binary
// During copying local source components, path represent base directory path whereas files is empty
// During `odo watch`, path represent base directory path whereas files contains list of changed Files
func PushLocal(client *occlient.Client, componentName string, applicationName string, path string, out io.Writer, files []string) error {
	const targetPath = "/opt/app-root/src"

	// Find DeploymentConfig for component
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := client.GetOneDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return errors.Wrap(err, "unable to get deployment for component")
	}
	// Find Pod for component
	podSelector := fmt.Sprintf("deploymentconfig=%s", dc.Name)

	// Wait for Pod to be in running state otherwise we can't sync data to it.
	pod, err := client.WaitAndGetPod(podSelector)
	if err != nil {
		return errors.Wrapf(err, "error while waiting for pod  %s", podSelector)
	}

	// Copy the files to the pod
	glog.V(4).Infof("Copying to pod %s", pod.Name)
	err = client.CopyFile(path, pod.Name, targetPath, files)
	if err != nil {
		return errors.Wrap(err, "unable push files to pod")
	}
	fmt.Fprintf(out, "Please wait, building component....\n")

	// use pipes to write output from ExecCMDInContainer in yellow  to 'out' io.Writer
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		yellowFprintln := color.New(color.FgYellow).FprintlnFunc()
		scanner := bufio.NewScanner(pipeReader)
		for scanner.Scan() {
			line := scanner.Text()
			// color.Output is temporarily used as there is a error when passing in color.Output from cmd/create.go and casting to io.writer in windows
			// TODO: Fix this in the future, more upstream in the code at cmd/create.go rather than within this function.
			yellowFprintln(color.Output, line)
		}
	}()

	err = client.ExecCMDInContainer(pod.Name,
		// We will use the assemble-and-restart script located within the supervisord container we've created
		[]string{"/var/lib/supervisord/bin/assemble-and-restart"},
		pipeWriter, pipeWriter, nil, false)
	if err != nil {
		return errors.Wrap(err, "unable to execute assemble script")
	}

	return nil
}

// Build component from BuildConfig.
// If 'streamLogs' is true than it streams build logs on stdout, set 'wait' to true if you want to return error if build fails.
// If 'wait' is true than it waits for build to successfully complete.
// If 'wait' is false than this function won't return error even if build failed.
func Build(client *occlient.Client, componentName string, applicationName string, streamLogs bool, wait bool, stdout io.Writer) error {

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	buildName, err := client.StartBuild(namespacedOpenShiftObject)
	if err != nil {
		return errors.Wrapf(err, "unable to rebuild %s", componentName)
	}
	if streamLogs {
		if err := client.FollowBuildLog(buildName, stdout); err != nil {
			return errors.Wrapf(err, "unable to follow logs for %s", buildName)
		}
	}
	if wait {
		if err := client.WaitForBuildToFinish(buildName); err != nil {
			return errors.Wrapf(err, "unable to wait for build %s", buildName)
		}
	}

	return nil
}

// GetComponentType returns type of component in given application and project
func GetComponentType(client *occlient.Client, componentName string, applicationName string, projectName string) (string, error) {

	// filter according to component and application name
	selector := fmt.Sprintf("%s=%s,%s=%s", componentlabels.ComponentLabel, componentName, applabels.ApplicationLabel, applicationName)
	componentImageTypes, err := client.GetLabelValues(projectName, componentlabels.ComponentTypeLabel, selector)
	if err != nil {
		return "", errors.Wrap(err, "unable to get type of %s component")
	}
	if len(componentImageTypes) < 1 {
		// no type returned
		return "", errors.Wrap(err, "unable to find type of %s component")

	}
	// check if all types are the same
	// it should be as we are secting only exactly one component, and it doesn't make sense
	// to have one component labeled with different component type labels
	for _, componentImageType := range componentImageTypes {
		if componentImageTypes[0] != componentImageType {
			return "", errors.Wrap(err, "data mismatch: %s component has objects with different types")
		}

	}
	return componentImageTypes[0], nil
}

// List lists components in active application
func List(client *occlient.Client, applicationName string, projectName string) ([]ComponentInfo, error) {

	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)
	componentNames, err := client.GetLabelValues(projectName, componentlabels.ComponentLabel, applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	components := []ComponentInfo{}

	for _, name := range componentNames {
		componentImageType, err := GetComponentType(client, name, applicationName, projectName)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to list components")
		}
		components = append(components, ComponentInfo{Name: name, Type: componentImageType})
	}

	return components, nil
}

// GetComponentSource what source type given component uses
// The first returned string is component source type ("git" or "local" or "binary")
// The second returned string is a source (url to git repository or local path or path to binary)
// we retrieve the source type by looking up the DeploymentConfig that's deployed
func GetComponentSource(client *occlient.Client, componentName string, applicationName string, projectName string) (string, string, error) {

	// Namespace the application
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to create namespaced name")
	}

	deploymentConfig, err := client.GetDeploymentConfigFromName(namespacedOpenShiftObject, projectName)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get source path for component %s", namespacedOpenShiftObject)
	}

	sourcePath := deploymentConfig.ObjectMeta.Annotations[componentSourceURLAnnotation]
	sourceType := deploymentConfig.ObjectMeta.Annotations[componentSourceTypeAnnotation]

	if !validateSourceType(sourceType) {
		return "", "", fmt.Errorf("unsupported component source type %s", sourceType)
	}

	glog.V(4).Infof("Source for component %s is %s (%s)", componentName, sourcePath, sourceType)
	return sourceType, sourcePath, nil
}

// Update updates the requested component
// componentName is the name of the component to be updated
// applicationName is the name of the application of the component
// newSourceType indicates the type of the new source i.e git/local/binary
// newSource indicates path of the source directory or binary or the git URL
// stdout is the io writer for streaming build logs on stdout
func Update(client *occlient.Client, componentName string, applicationName string, newSourceType string, newSource string, stdout io.Writer) error {

	// STEP 1. Create the common Object Meta for updating.

	// Retrieve the current project name
	projectName := client.GetCurrentProjectName()

	// Retrieve the old source type
	oldSourceType, _, err := GetComponentSource(client, componentName, applicationName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to get source of %s component", componentName)
	}

	// Namespace the application
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Create annotations
	annotations := map[string]string{componentSourceURLAnnotation: newSource}
	annotations[componentSourceTypeAnnotation] = newSourceType

	// Component Type
	componentImageType, err := GetComponentType(client, componentName, applicationName, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get component image type for updating")
	}

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(componentImageType)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// Retrieve labels
	// Save component type as label
	labels := componentlabels.GetLabels(componentName, applicationName, true)
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// ObjectMetadata are the same for all generated objects
	// Create common metadata that will be updated throughout all objects.
	commonObjectMeta := metav1.ObjectMeta{
		Name:        namespacedOpenShiftObject,
		Labels:      labels,
		Annotations: annotations,
	}

	envVars, err := client.GetEnvVarsFromDC(namespacedOpenShiftObject, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to get env vars of %s component", componentName)
	}

	// STEP 2. Determine what the new source is going to be

	glog.V(4).Infof("Updating component %s, from %s to %s (%s).", componentName, oldSourceType, newSource, newSourceType)

	if (oldSourceType == "local" || oldSourceType == "binary") && newSourceType == "git" {
		// Steps to update component from local or binary to git
		// 1. Create a BuildConfig
		// 2. Update DeploymentConfig with the new image
		// 3. Clean up
		// 4. Build the application

		// CreateBuildConfig here!
		glog.V(4).Infof("Creating BuildConfig %s using imageName: %s for updating", namespacedOpenShiftObject, imageName)
		bc, err := client.CreateBuildConfig(commonObjectMeta, imageName, newSource, envVars)
		if err != nil {
			return errors.Wrapf(err, "unable to update BuildConfig  for %s component", componentName)
		}

		// Update / replace the current DeploymentConfig with a Git one (not SupervisorD!)
		glog.V(4).Infof("Updating the DeploymentConfig %s image to %s", namespacedOpenShiftObject, bc.Spec.Output.To.Name)
		err = client.UpdateDCToGit(commonObjectMeta, bc.Spec.Output.To.Name)
		if err != nil {
			return errors.Wrapf(err, "unable to update DeploymentConfig image for %s component", componentName)
		}

		// Cleanup after the supervisor
		err = client.CleanupAfterSupervisor(namespacedOpenShiftObject, projectName, annotations)
		if err != nil {
			return errors.Wrapf(err, "unable to update DeploymentConfig  for %s component", componentName)
		}

		// Finally, we build!
		err = Build(client, componentName, applicationName, true, true, stdout)
		if err != nil {
			return errors.Wrapf(err, "unable to build the component %v", componentName)
		}

	} else if oldSourceType == "git" && (newSourceType == "binary" || newSourceType == "local") {
		// Steps to update component from git to local or binary

		// Update the sourceURL since it is not a local/binary file.
		sourceURL := util.GenFileUrl(newSource, runtime.GOOS)
		annotations[componentSourceURLAnnotation] = sourceURL

		// Need to delete the old BuildConfig
		err = client.DeleteBuildConfig(commonObjectMeta)
		if err != nil {
			return errors.Wrapf(err, "unable to delete BuildConfig for %s component", componentName)
		}

		// Update the DeploymentConfig
		err = client.UpdateDCToSupervisor(commonObjectMeta, componentImageType)
		if err != nil {
			return errors.Wrapf(err, "unable to update DeploymentConfig for %s component", componentName)
		}

	} else {
		// save source path as annotation
		// this part is for updates where the source does not change or change from local to binary and vice versa

		if newSourceType == "git" {

			// Update the BuildConfig
			err = client.UpdateBuildConfig(namespacedOpenShiftObject, projectName, newSource, annotations)
			if err != nil {
				return errors.Wrapf(err, "unable to update the build config %v", componentName)
			}

			// Update DeploymentConfig annotations as well
			err = client.UpdateDCAnnotations(namespacedOpenShiftObject, annotations)
			if err != nil {
				return errors.Wrapf(err, "unable to update the deployment config %v", componentName)
			}

			// Build it
			err = Build(client, componentName, applicationName, true, true, stdout)

		} else if newSourceType == "local" || newSourceType == "binary" {

			// Update the sourceURL
			sourceURL := util.GenFileUrl(newSource, runtime.GOOS)
			annotations[componentSourceURLAnnotation] = sourceURL
			err = client.UpdateDCAnnotations(namespacedOpenShiftObject, annotations)
		}

		if err != nil {
			return errors.Wrap(err, "unable to update the component")
		}
	}
	return nil
}

// Exists checks whether a component with the given name exists in the current application or not
// componentName is the component name to perform check for
// The first returned parameter is a bool indicating if a component with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client *occlient.Client, componentName, applicationName, projectName string) (bool, error) {

	componentList, err := List(client, applicationName, projectName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get the component list")
	}
	for _, component := range componentList {
		if component.Name == componentName {
			return true, nil
		}
	}
	return false, nil
}

// LinkInfo contains the information about the link being created between
// components
type LinkInfo struct {
	SourceComponent string
	TargetComponent string
	Envs            []string
}

// Link injects connection information of the target component into the source
// component as environment variables
func Link(client *occlient.Client, sourceComponent, targetComponent, applicationName string) (*LinkInfo, error) {
	// Get Service associated with the target component
	serviceLabels := componentlabels.GetLabels(targetComponent, applicationName, false)
	serviceSelector := util.ConvertLabelsToSelector(serviceLabels)
	targetService, err := client.GetOneServiceFromSelector(serviceSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get service associated with the component %v", targetComponent)
	}

	// Generate environment variables to inject
	linkHostKey := fmt.Sprintf("COMPONENT_%v_HOST", strings.ToUpper(targetComponent))
	linkHostValue := targetService.Name

	linkPortKey := fmt.Sprintf("COMPONENT_%v_PORT", strings.ToUpper(targetComponent))
	linkPort, err := getPortFromService(targetService)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get port from Service %v", targetService.Name)
	}
	linkPortValue := fmt.Sprint(linkPort)

	// Inject environment variable to source component
	dcLabels := componentlabels.GetLabels(sourceComponent, applicationName, false)
	dcSelector := util.ConvertLabelsToSelector(dcLabels)
	sourceDC, err := client.GetOneDeploymentConfigFromSelector(dcSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get Deployment Config for component %v", sourceComponent)
	}

	if err := client.AddEnvironmentVariablesToDeploymentConfig(
		[]corev1.EnvVar{
			{
				Name:  linkHostKey,
				Value: linkHostValue,
			},
			{
				Name:  linkPortKey,
				Value: linkPortValue,
			},
		}, sourceDC); err != nil {
		return nil, errors.Wrapf(err, "unable to add environment variables to Deployment Config %v", sourceDC.Name)
	}

	return &LinkInfo{
		SourceComponent: sourceComponent,
		TargetComponent: targetComponent,
		Envs: []string{
			fmt.Sprintf("%v=%v", linkHostKey, linkHostValue),
			fmt.Sprintf("%v=%v", linkPortKey, linkPortValue),
		},
	}, nil
}

// getPortFromService returns the first port listed in the service
func getPortFromService(service *corev1.Service) (int32, error) {
	numServicePorts := len(service.Spec.Ports)
	if numServicePorts != 1 {
		return 0, fmt.Errorf("expected exactly one port in the service %v, but got %v", service.Name, numServicePorts)
	}

	return service.Spec.Ports[0].Port, nil

}

// GetComponentDesc provides description such as source, url & storage about given component
func GetComponentDesc(client *occlient.Client, currentComponent string, currentApplication string, currentProject string) (componentImageType string, path string, componentURL string, appStore []storage.StorageInfo, err error) {
	// Component Type
	componentImageType, err = GetComponentType(client, currentComponent, currentApplication, currentProject)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get source path")
	}
	// Source
	_, path, err = GetComponentSource(client, currentComponent, currentApplication, currentProject)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get source path")
	}
	// URL
	urlList, err := urlpkg.List(client, currentComponent, currentApplication)
	if len(urlList) != 0 {
		componentURL = urlList[0].URL
	}
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get url list")
	}
	//Storage
	appStore, err = storage.List(client, currentComponent, currentApplication)
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "unable to get storage list")
	}

	return componentImageType, path, componentURL, appStore, nil
}

// GetLogs follow the DeploymentConfig logs if follow is set to true
func GetLogs(client *occlient.Client, componentName string, applicationName string, follow bool, stdout io.Writer) error {

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Retrieve the logs
	err = client.DisplayDeploymentConfigLog(namespacedOpenShiftObject, follow, stdout)
	if err != nil {
		return err
	}

	return nil
}
