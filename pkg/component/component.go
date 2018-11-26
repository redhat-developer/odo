package component

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/log"
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

// Info holds all important information about one component
type Info struct {
	Name string
	Type string
}

// Description holds all information about component
type Description struct {
	ComponentName      string                `json:"componentName,omitempty"`
	ComponentImageType string                `json:"type,omitempty"`
	Path               string                `json:"source,omitempty"`
	URLs               []urlpkg.URL          `json:"url,omitempty"`
	Env                []corev1.EnvVar       `json:"environment,omitempty"`
	Storage            []storage.StorageInfo `json:"storage,omitempty"`
}

// GetComponentDir returns source repo name
// Parameters:
//		path: git url or source path or binary path
//		paramType: One of CreateType as in GIT/LOCAL/BINARY
// Returns: directory name
func GetComponentDir(path string, paramType occlient.CreateType) (string, error) {
	retVal := ""
	switch paramType {
	case occlient.GIT:
		retVal = strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".git")
	case occlient.LOCAL:
		retVal = filepath.Base(path)
	case occlient.BINARY:
		filename := filepath.Base(path)
		var extension = filepath.Ext(filename)
		retVal = filename[0 : len(filename)-len(extension)]
	default:
		currDir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrapf(err, "unable to generate a random name as getting current directory failed")
		}
		retVal = filepath.Base(currDir)
	}
	retVal = strings.TrimSpace(util.GetDNS1123Name(strings.ToLower(retVal)))
	return retVal, nil
}

// GetDefaultComponentName generates a unique component name
// Parameters: desired default component name(w/o prefix) and slice of existing component names
// Returns: Unique component name and error if any
func GetDefaultComponentName(componentPath string, componentPathType occlient.CreateType, componentType string, existingComponentList []Info) (string, error) {
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

	// If there's no prefix in config file, or its value is empty string use safe default - the current directory along with component type
	if cfg.OdoSettings.NamePrefix == nil || *cfg.OdoSettings.NamePrefix == "" {
		prefix, err = GetComponentDir(componentPath, componentPathType)
		if err != nil {
			return "", errors.Wrap(err, "unable to generate random component name")
		}
		prefix = util.TruncateString(prefix, componentRandomNamePartsMaxLen)
	} else {
		// Set the required prefix into componentName
		prefix = *cfg.OdoSettings.NamePrefix
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

	return util.GetDNS1123Name(componentName), nil
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
func CreateFromGit(client *occlient.Client, params occlient.CreateArgs) error {

	labels := componentlabels.GetLabels(params.Name, params.ApplicationName, true)

	// Loading spinner
	s := log.Spinnerf("Creating component %s", params.Name)
	defer s.End(false)

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(params.ImageName)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// save source path as annotation
	annotations := map[string]string{componentSourceURLAnnotation: params.SourcePath}
	annotations[componentSourceTypeAnnotation] = "git"

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(params.Name, params.ApplicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Create CommonObjectMeta to be passed in
	commonObjectMeta := metav1.ObjectMeta{
		Name:        namespacedOpenShiftObject,
		Labels:      labels,
		Annotations: annotations,
	}

	err = client.NewAppS2I(params, commonObjectMeta)
	if err != nil {
		return errors.Wrapf(err, "unable to create git component %s", namespacedOpenShiftObject)
	}

	s.End(true)
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
func CreateFromPath(client *occlient.Client, params occlient.CreateArgs) error {
	labels := componentlabels.GetLabels(params.Name, params.ApplicationName, true)

	// Loading spinner
	s := log.Spinnerf("Creating component %s", params.Name)
	defer s.End(false)

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(params.ImageName)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// save source path as annotation
	sourceURL := util.GenFileURL(params.SourcePath, runtime.GOOS)
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL}
	annotations[componentSourceTypeAnnotation] = string(params.SourceType)

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(params.Name, params.ApplicationName)
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
	err = client.BootstrapSupervisoredS2I(params, commonObjectMeta)
	if err != nil {
		return err
	}

	if params.Watch {
		selector := fmt.Sprintf("deployment=%s-%s-1", labels["app.kubernetes.io/component-name"], labels["app"])
		_, err = client.WaitAndGetPod(selector)
		return err
	}

	s.End(true)
	return nil
}

// Delete whole component
func Delete(client *occlient.Client, componentName string, applicationName string) error {

	// Loading spinner
	s := log.Spinnerf("Deleting component %s", componentName)
	defer s.End(false)

	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to create new configuration to delete %s", componentName)
	}

	labels := componentlabels.GetLabels(componentName, applicationName, false)
	err = client.Delete(labels)
	if err != nil {
		return errors.Wrapf(err, "error deleting component %s", componentName)
	}

	// Get a list of all components
	components, err := List(client, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve list of components")
	}

	// First check is that we want to only update the active component if the component which is getting deleted is the
	// active component
	// Second check is that we want to do an update only if it is happening for the active application otherwise we need
	// not to care for the update of the active component
	activeComponent := cfg.GetActiveComponent(applicationName, client.Namespace)
	activeApplication := cfg.GetActiveApplication(client.Namespace)
	if activeComponent == componentName && activeApplication == applicationName {
		// We will *only* set a new component if either len(components) is zero, or the
		// current component matches the one being deleted.
		if current := cfg.GetActiveComponent(applicationName, client.Namespace); current == componentName || len(components) == 0 {

			// If there's more than one component, set it to the first one..
			if len(components) > 0 {
				err = cfg.SetActiveComponent(components[0].Name, applicationName, client.Namespace)

				if err != nil {
					return errors.Wrapf(err, "unable to set current component to '%s'", componentName)
				}
			} else {
				// Unset to blank
				err = cfg.UnsetActiveComponent(client.Namespace)
				if err != nil {
					return errors.Wrapf(err, "error unsetting current component while deleting %s", componentName)
				}
			}
		}

	}

	s.End(true)
	return nil
}

// SetCurrent sets the given componentName as active component
func SetCurrent(componentName string, applicationName string, projectName string) error {
	cfg, err := config.New()
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", componentName)
	}

	err = cfg.SetActiveComponent(componentName, applicationName, projectName)
	if err != nil {
		return errors.Wrapf(err, "unable to set current component %s", componentName)
	}

	return nil
}

// GetCurrent component in active application
// returns "" if there is no active component
func GetCurrent(applicationName string, projectName string) (string, error) {
	cfg, err := config.New()
	if err != nil {
		return "", errors.Wrap(err, "unable to get config")
	}
	currentComponent := cfg.GetActiveComponent(applicationName, projectName)

	return currentComponent, nil
}

// getEnvFromPodEnvs loops through the passed slice of pod#EnvVars and gets the value corresponding to the key passed, returns empty stirng if not available
func getEnvFromPodEnvs(envName string, podEnvs []corev1.EnvVar) string {
	for _, podEnv := range podEnvs {
		if podEnv.Name == envName {
			return podEnv.Value
		}
	}
	return ""
}

// PushLocal push local code to the cluster and trigger build there.
// files is list of changed files captured during `odo watch` as well as binary file path
// During copying binary components, path represent base directory path to binary and files contains path of binary
// During copying local source components, path represent base directory path whereas files is empty
// During `odo watch`, path represent base directory path whereas files contains list of changed Files
func PushLocal(client *occlient.Client, componentName string, applicationName string, path string, out io.Writer, files []string) error {
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

	// Get S2I Source/Binary Path from Pod Env variables created at the time of component create
	s2iSrcPath := getEnvFromPodEnvs(occlient.EnvS2ISrcOrBinPath, pod.Spec.Containers[0].Env)
	if s2iSrcPath == "" {
		s2iSrcPath = occlient.DefaultS2ISrcOrBinPath
	}
	targetPath := fmt.Sprintf("%s/src", s2iSrcPath)

	// Copy the files to the pod

	s := log.Spinner("Copying files to pod")
	err = client.CopyFile(path, pod.Name, targetPath, files)
	if err != nil {
		s.End(false)
		return errors.Wrap(err, "unable push files to pod")
	}
	s.End(true)

	s = log.Spinner("Building component")
	defer s.End(false)

	// use pipes to write output from ExecCMDInContainer in yellow  to 'out' io.Writer
	pipeReader, pipeWriter := io.Pipe()
	var cmdOutput string
	go func() {
		yellowFprintln := color.New(color.FgYellow).FprintlnFunc()
		scanner := bufio.NewScanner(pipeReader)
		for scanner.Scan() {
			line := scanner.Text()
			// color.Output is temporarily used as there is a error when passing in color.Output from cmd/create.go and casting to io.writer in windows
			// TODO: Fix this in the future, more upstream in the code at cmd/create.go rather than within this function.
			// If we are in debug mode, we should show the output
			if log.IsDebug() {
				yellowFprintln(color.Output, line)
			}

			cmdOutput += fmt.Sprintln(line)
		}
	}()

	err = client.ExecCMDInContainer(pod.Name,
		// We will use the assemble-and-restart script located within the supervisord container we've created
		[]string{"/var/lib/supervisord/bin/assemble-and-restart"},
		pipeWriter, pipeWriter, nil, false)

	if err != nil {
		// If we fail, log the output
		log.Errorf("Unable to build files\n%v", cmdOutput)
		return errors.Wrap(err, "unable to execute assemble script")
	}

	s.End(true)

	return nil
}

// Build component from BuildConfig.
// If 'wait' is true than it waits for build to successfully complete.
// If 'wait' is false than this function won't return error even if build failed.
func Build(client *occlient.Client, componentName string, applicationName string, wait bool, stdout io.Writer) error {

	// Loading spinner
	s := log.Spinnerf("Triggering build from git")
	defer s.End(false)

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	buildName, err := client.StartBuild(namespacedOpenShiftObject)
	if err != nil {
		return errors.Wrapf(err, "unable to rebuild %s", componentName)
	}

	// Retrieve the Build Log and write to buffer if debug is disabled, else we we output to stdout / debug.

	var b bytes.Buffer
	if !log.IsDebug() {
		stdout = bufio.NewWriter(&b)
	}

	if err := client.FollowBuildLog(buildName, stdout); err != nil {
		return errors.Wrapf(err, "unable to follow logs for %s", buildName)
	}

	if wait {
		if err := client.WaitForBuildToFinish(buildName); err != nil {
			return errors.Wrapf(err, "unable to build %s, error: %s", buildName, b.String())
		}
	}

	s.End(true)
	return nil
}

// GetComponentType returns type of component in given application and project
func GetComponentType(client *occlient.Client, componentName string, applicationName string) (string, error) {

	// filter according to component and application name
	selector := fmt.Sprintf("%s=%s,%s=%s", componentlabels.ComponentLabel, componentName, applabels.ApplicationLabel, applicationName)
	componentImageTypes, err := client.GetLabelValues(componentlabels.ComponentTypeLabel, selector)
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
func List(client *occlient.Client, applicationName string) ([]Info, error) {

	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)

	// retrieve all the deployment configs that are associated with this application
	dcList, err := client.GetDeploymentConfigsFromSelector(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}

	var components []Info

	// extract the labels we care about from each component
	for _, elem := range dcList {
		components = append(components,
			Info{
				Name: elem.Labels[componentlabels.ComponentLabel],
				Type: elem.Labels[componentlabels.ComponentTypeLabel],
			},
		)
	}

	return components, nil
}

// GetComponentSource what source type given component uses
// The first returned string is component source type ("git" or "local" or "binary")
// The second returned string is a source (url to git repository or local path or path to binary)
// we retrieve the source type by looking up the DeploymentConfig that's deployed
func GetComponentSource(client *occlient.Client, componentName string, applicationName string) (string, string, error) {

	// Namespace the application
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to create namespaced name")
	}

	deploymentConfig, err := client.GetDeploymentConfigFromName(namespacedOpenShiftObject)
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
func Update(client *occlient.Client, componentName string, applicationName string, newSourceType string, newSource string, newSourceRef string, stdout io.Writer) error {

	// STEP 1. Create the common Object Meta for updating.

	// Retrieve the old source type
	oldSourceType, _, err := GetComponentSource(client, componentName, applicationName)
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
	componentImageType, err := GetComponentType(client, componentName, applicationName)
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

	envVars, err := client.GetEnvVarsFromDC(namespacedOpenShiftObject)
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
		bc, err := client.CreateBuildConfig(commonObjectMeta, imageName, newSource, newSourceRef, envVars)
		if err != nil {
			return errors.Wrapf(err, "unable to update BuildConfig  for %s component", componentName)
		}

		// Update / replace the current DeploymentConfig with a Git one (not SupervisorD!)
		glog.V(4).Infof("Updating the DeploymentConfig %s image to %s", namespacedOpenShiftObject, bc.Spec.Output.To.Name)
		err = client.UpdateDCToGit(commonObjectMeta, bc.Spec.Output.To.Name)
		if err != nil {
			return errors.Wrapf(err, "unable to update DeploymentConfig image for %s component", componentName)
		}

		// Finally, we build!
		err = Build(client, componentName, applicationName, true, stdout)
		if err != nil {
			return errors.Wrapf(err, "unable to build the component %v", componentName)
		}

	} else if oldSourceType == "git" && (newSourceType == "binary" || newSourceType == "local") {
		// Steps to update component from git to local or binary

		// Update the sourceURL since it is not a local/binary file.
		sourceURL := util.GenFileURL(newSource, runtime.GOOS)
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
			err = client.UpdateBuildConfig(namespacedOpenShiftObject, newSource, annotations)
			if err != nil {
				return errors.Wrapf(err, "unable to update the build config %v", componentName)
			}

			// Update DeploymentConfig annotations as well
			err = client.UpdateDCAnnotations(namespacedOpenShiftObject, annotations)
			if err != nil {
				return errors.Wrapf(err, "unable to update the deployment config %v", componentName)
			}

			// Build it
			err = Build(client, componentName, applicationName, true, stdout)

		} else if newSourceType == "local" || newSourceType == "binary" {

			// Update the sourceURL
			sourceURL := util.GenFileURL(newSource, runtime.GOOS)
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
func Exists(client *occlient.Client, componentName, applicationName string) (bool, error) {

	componentList, err := List(client, applicationName)
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

// GetComponentDesc provides description such as source, url & storage about given component
func GetComponentDesc(client *occlient.Client, componentName string, applicationName string) (componentDesc Description, err error) {
	// Component Type
	componentImageType, err := GetComponentType(client, componentName, applicationName)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get source path")
	}
	// Source
	_, path, err := GetComponentSource(client, componentName, applicationName)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get source path")
	}
	// URL
	urlList, err := urlpkg.List(client, componentName, applicationName)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get url list")
	}
	// Storage
	appStore, err := storage.List(client, componentName, applicationName)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get storage list")
	}
	// Environment Variables
	DC, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get DC list")
	}
	envVars, err := client.GetEnvVarsFromDC(DC)
	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get envVars list")
	}
	var filteredEnv []corev1.EnvVar
	for _, env := range envVars {
		if !strings.Contains(env.Name, "ODO") {
			filteredEnv = append(filteredEnv, env)
		}
	}

	if err != nil {
		return componentDesc, errors.Wrap(err, "unable to get envVars list")
	}
	componentDesc = Description{
		ComponentName:      componentName,
		ComponentImageType: componentImageType,
		Path:               path,
		Env:                filteredEnv,
		Storage:            appStore,
		URLs:               urlList,
	}
	return componentDesc, nil
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
