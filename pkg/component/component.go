package component

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/catalog"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/storage"
	urlpkg "github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	pkgUtil "github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// componentSourceURLAnnotation is an source url from which component was build
// it can be also file://
const componentSourceURLAnnotation = "app.kubernetes.io/url"
const ComponentSourceTypeAnnotation = "app.kubernetes.io/component-source-type"
const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1

// GetComponentDir returns source repo name
// Parameters:
//		path: git url or source path or binary path
//		paramType: One of CreateType as in GIT/LOCAL/BINARY
// Returns: directory name
func GetComponentDir(path string, paramType config.SrcType) (string, error) {
	retVal := ""
	switch paramType {
	case config.GIT:
		retVal = strings.TrimSuffix(path[strings.LastIndex(path, "/")+1:], ".git")
	case config.LOCAL:
		retVal = filepath.Base(path)
	case config.BINARY:
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
func GetDefaultComponentName(componentPath string, componentPathType config.SrcType, componentType string, existingComponentList ComponentList) (string, error) {
	var prefix string

	// Get component names from component list
	var existingComponentNames []string
	for _, component := range existingComponentList.Items {
		existingComponentNames = append(existingComponentNames, component.Name)
	}

	// Fetch config
	cfg, err := preference.New()
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
	annotations[ComponentSourceTypeAnnotation] = "git"

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

// GetComponentLinkedSecretNames provides a slice containing the names of the secrets that are present in envFrom
func GetComponentLinkedSecretNames(client *occlient.Client, componentName string, applicationName string) (secretNames []string, err error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	dc, err := client.GetOneDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to fetch deployment configs for the selector %v", componentSelector)
	}

	for _, env := range dc.Spec.Template.Spec.Containers[0].EnvFrom {
		if env.SecretRef != nil {
			secretNames = append(secretNames, env.SecretRef.Name)
		}
	}

	return secretNames, nil
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
	sourceURL := util.GenFileURL(params.SourcePath)
	annotations := map[string]string{componentSourceURLAnnotation: sourceURL}
	annotations[ComponentSourceTypeAnnotation] = string(params.SourceType)

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
	s.End(true)

	if params.Wait {
		// if wait flag is present then extract the podselector
		// use the podselector for calling WaitAndGetPod
		selectorLabels, err := util.NamespaceOpenShiftObject(labels[componentlabels.ComponentLabel], labels["app"])
		if err != nil {
			return err
		}

		podSelector := fmt.Sprintf("deploymentconfig=%s", selectorLabels)
		_, err = client.WaitAndGetPod(podSelector, corev1.PodRunning, "Waiting for component to start")
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// Delete whole component
func Delete(client *occlient.Client, componentName string, applicationName string) error {

	// Loading spinner
	s := log.Spinnerf("Deleting component %s", componentName)
	defer s.End(false)

	labels := componentlabels.GetLabels(componentName, applicationName, false)
	err := client.Delete(labels)
	if err != nil {
		return errors.Wrapf(err, "error deleting component %s", componentName)
	}

	s.End(true)
	return nil
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

// getS2IPaths returns slice of s2i paths of odo interest
// Parameters:
//	podEnvs: Slice of env vars extracted from pod template
// Returns:
//	Slice of s2i paths extracted from passed parameters
func getS2IPaths(podEnvs []corev1.EnvVar) []string {
	retVal := []string{}
	// List of s2i Paths exported for use in container pod for working with source/binary
	s2iPathEnvs := []string{
		occlient.EnvS2IDeploymentDir,
		occlient.EnvS2ISrcOrBinPath,
		occlient.EnvS2IWorkingDir,
		occlient.EnvS2ISrcBackupDir,
	}
	// For each of the required env var
	for _, s2iPathEnv := range s2iPathEnvs {
		// try to fetch the value of required env from the ones set already in the component container like for the case of watch or multiple pushes
		envVal := getEnvFromPodEnvs(s2iPathEnv, podEnvs)
		isEnvValPresent := false
		if envVal != "" {
			for _, e := range retVal {
				if envVal == e {
					isEnvValPresent = true
					break
				}
			}
			if !isEnvValPresent {
				// If `src` not in path, append it
				if filepath.Base(envVal) != "src" {
					envVal = filepath.Join(envVal, "src")
				}
				retVal = append(retVal, envVal)
			}
		}
	}
	// Append binary backup path to s2i paths list
	retVal = append(retVal, occlient.DefaultS2IDeploymentBackupDir)
	return retVal
}

// CreateComponent creates component as per the passed component settings
//	Parameters:
//		client: occlient instance
//		componentConfig: the component configuration that holds all details of component
//		context: the component context indicating the location of component config and hence its source as well
//		stdout: io.Writer instance to write output to
//	Returns:
//		err: errors if any
func CreateComponent(client *occlient.Client, componentConfig config.LocalConfigInfo, context string, stdout io.Writer) (err error) {

	cmpName := componentConfig.GetName()
	cmpType := componentConfig.GetType()
	cmpSrcType := componentConfig.GetSourceType()
	cmpPorts := componentConfig.GetPorts()
	cmpSrcRef := componentConfig.GetRef()
	appName := componentConfig.GetApplication()

	log.Successf("Initializing '%s' component", cmpName)
	createArgs := occlient.CreateArgs{
		Name:            cmpName,
		ImageName:       cmpType,
		ApplicationName: appName,
	}
	createArgs.SourceType = cmpSrcType
	createArgs.SourcePath = componentConfig.GetSourceLocation()

	if len(cmpPorts) > 0 {
		createArgs.Ports = cmpPorts
	}

	createArgs.Resources, err = occlient.GetResourceRequirementsFromCmpSettings(componentConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create component")
	}

	switch cmpSrcType {
	case config.GIT:
		// Use Git
		if cmpSrcRef != "" {
			createArgs.SourceRef = cmpSrcRef
		}
		if err = CreateFromGit(
			client,
			createArgs,
		); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", createArgs)
		}
		// Git is the only one using BuildConfig since we need to retrieve the git
		if err = Build(client, createArgs.Name, createArgs.ApplicationName, true, stdout); err != nil {
			return errors.Wrapf(err, "failed to build component with args %+v", componentConfig)
		}
	case config.LOCAL:
		fileInfo, err := os.Stat(createArgs.SourcePath)
		if err != nil {
			return errors.Wrapf(err, "failed to get info of path %+v of component %+v", createArgs.SourcePath, createArgs)
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("component creation with args %+v as path needs to be a directory", createArgs)
		}
		// Create
		if err = CreateFromPath(client, createArgs); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", createArgs)
		}
	case config.BINARY:
		if err = CreateFromPath(client, createArgs); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", createArgs)
		}
	default:
		// If the user does not provide anything (local, git or binary), use the current absolute path and deploy it
		createArgs.SourceType = config.LOCAL
		dir, err := os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "cannot create component with current path as local source path since no component source details are passed")
		}
		createArgs.SourcePath = dir
		if err = CreateFromPath(client, createArgs); err != nil {
			return errors.Wrapf(err, "")
		}
	}
	return
}

// ValidateComponentCreateRequest validates request for component creation and returns errors if any
// Parameters:
//	componentSettings: Component settings
//	isCmpExistsCheck: boolean to indicate whether or not error out if component with same name already exists
// Returns:
//	errors if any
func ValidateComponentCreateRequest(client *occlient.Client, componentSettings config.ComponentSettings, isCmpExistsCheck bool) (err error) {
	// Check manddatory fields are set
	if componentSettings.Name == nil ||
		componentSettings.Application == nil ||
		componentSettings.Project == nil ||
		componentSettings.SourceType == nil ||
		componentSettings.SourceLocation == nil ||
		componentSettings.Type == nil {
		return fmt.Errorf("component name, application, project, source type, source location and component type are mandatory parameters for component creation")
	}

	_, componentType, _, componentVersion := util.ParseComponentImageName(*componentSettings.Type)
	// Check to see if the catalog type actually exists
	exists, err := catalog.Exists(client, componentType)
	if err != nil {
		return errors.Wrapf(err, "Failed to create component of type %s", componentType)
	}
	if !exists {
		log.Info("Run 'odo catalog list components' for a list of supported component types")
		return fmt.Errorf("Failed to find component of type %s", componentType)
	}

	// Check to see if that particular version exists
	versionExists, err := catalog.VersionExists(client, componentType, componentVersion)
	if err != nil {
		return errors.Wrapf(err, "Failed to create component of type %s of version %s", componentType, componentVersion)
	}
	if !versionExists {
		log.Info("Run 'odo catalog list components' to see a list of supported component type versions")
		return fmt.Errorf("Invalid component version %s:%s", componentType, componentVersion)
	}

	// Validate component name
	err = validation.ValidateName(*componentSettings.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to create component of name %s", *componentSettings.Name)
	}

	if isCmpExistsCheck {
		exists, err = Exists(client, *componentSettings.Name, *componentSettings.Application)
		if err != nil {
			return errors.Wrapf(err, "failed to check if component of name %s exists in application %s", *componentSettings.Name, *componentSettings.Application)
		}
		if exists {
			return fmt.Errorf("component with name %s already exists in application %s", *componentSettings.Name, *componentSettings.Application)
		}
	}

	if *componentSettings.SourceType == config.LOCAL {
		srcLocInfo, err := os.Stat(*(componentSettings.SourceLocation))
		if err != nil {
			return errors.Wrap(err, "failed to create component. Please view the settings used using the command `odo config view`")
		}
		if !srcLocInfo.IsDir() {
			return fmt.Errorf("source path for component created for local source needs to be a directory")
		}
	}

	return
}

// ApplyConfig applies the component config onto component dc
// Parameters:
//	client: occlient instance
//	appName: Name of application of which the component is a part
//	componentName: Name of the component which is being patched with config
//	componentConfig: Component configuration
// Returns:
//	err: Errors if any else nil
func ApplyConfig(client *occlient.Client, componentConfig config.LocalConfigInfo, context string, stdout io.Writer) (err error) {

	cmpName := componentConfig.GetName()
	cmpSrcLoc := componentConfig.GetSourceLocation()
	cmpSrcType := componentConfig.GetSourceType()

	if cmpSrcType == config.GIT {
		if err := Update(client, componentConfig, cmpSrcLoc, stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully", cmpName)
	} else if cmpSrcType == config.LOCAL {
		cmpSrcLoc, err := pkgUtil.GetAbsPath(cmpSrcLoc)
		if err != nil {
			return err
		}
		fileInfo, err := os.Stat(cmpSrcLoc)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("incorrect context passed. It should be a path to the compoennt source directory")
		}
		if err = Update(client, componentConfig, cmpSrcLoc, stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully", cmpName)
	} else if cmpSrcType == config.BINARY {
		cmpSrcLoc, err := pkgUtil.GetAbsPath(cmpSrcLoc)
		if err != nil {
			return err
		}
		if err = Update(client, componentConfig, cmpSrcLoc, stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully", cmpName)
	}
	return
}

// PushLocal push local code to the cluster and trigger build there.
// During copying binary components, path represent base directory path to binary and files contains path of binary
// During copying local source components, path represent base directory path whereas files is empty
// During `odo watch`, path represent base directory path whereas files contains list of changed Files
// Parameters:
//	componentName is name of the component to update sources to
//	applicationName is the name of the application of which the component is a part
//	path is base path of the component source/binary
// 	files is list of changed files captured during `odo watch` as well as binary file path
// 	delFiles is the list of files identified as deleted
// 	isForcePush indicates if the sources to be updated are due to a push in which case its a full source directory push or only push of identified sources
// 	globExps are the glob expressions which are to be ignored during the push
// Returns
//	Error if any
func PushLocal(client *occlient.Client, componentName string, applicationName string, path string, out io.Writer, files []string, delFiles []string, isForcePush bool, globExps []string) error {
	glog.V(4).Infof("PushLocal: componentName: %s, applicationName: %s, path: %s, files: %s, delFiles: %s, isForcePush: %+v", componentName, applicationName, path, files, delFiles, isForcePush)
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
	pod, err := client.WaitAndGetPod(podSelector, corev1.PodRunning, "Waiting for component to start")
	if err != nil {
		return errors.Wrapf(err, "error while waiting for pod  %s", podSelector)
	}

	// Get S2I Source/Binary Path from Pod Env variables created at the time of component create
	s2iSrcPath := getEnvFromPodEnvs(occlient.EnvS2ISrcOrBinPath, pod.Spec.Containers[0].Env)
	if s2iSrcPath == "" {
		s2iSrcPath = occlient.DefaultS2ISrcOrBinPath
	}
	targetPath := fmt.Sprintf("%s/src", s2iSrcPath)

	// If there are files identified as deleted, propagate them to the component pod
	if len(delFiles) > 0 {
		glog.V(4).Infof("propogating deletion of files %s to pod", strings.Join(delFiles, " "))
		/*
			Delete files observed by watch to have been deleted from each of s2i directories like:
				deployment dir: In interpreted runtimes like python, source is copied over to deployment dir so delete needs to happen here as well
				destination dir: This is the directory where s2i expects source to be copied for it be built and deployed
				working dir: Directory where, sources are copied over from deployment dir from where the s2i builds and deploys source.
							 Deletes need to happen here as well otherwise, even if the latest source is copied over, the stale source files remain
				source backup dir: Directory used for backing up source across multiple iterations of push and watch in component container
								   In case of python, s2i image moves sources from destination dir to workingdir which means sources are deleted from destination dir
								   So, during the subsequent watch pushing new diff to component pod, the source as a whole doesn't exist at destination dir and hence needs
								   to be backed up.
		*/
		err := client.PropagateDeletes(pod.Name, delFiles, getS2IPaths(pod.Spec.Containers[0].Env))
		if err != nil {
			return errors.Wrapf(err, "unable to propagate file deletions %+v", delFiles)
		}
	}

	// Copy the files to the pod
	s := log.Spinner("Copying files to component")

	if !isForcePush {
		if len(files) == 0 && len(delFiles) == 0 {
			return fmt.Errorf("pass files modifications/deletions to sync to component pod or force push")
		}
	}

	if isForcePush || len(files) > 0 {
		glog.V(4).Infof("Copying files %s to pod", strings.Join(files, " "))
		err = client.CopyFile(path, pod.Name, targetPath, files, globExps)
		if err != nil {
			s.End(false)
			return errors.Wrap(err, "unable push files to pod")
		}
	}
	s.End(true)

	s = log.Spinner("Building component")

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
		s.End(false)
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
	s.End(true)

	// Retrieve the Build Log and write to buffer if debug is disabled, else we we output to stdout / debug.

	var b bytes.Buffer
	if !log.IsDebug() {
		stdout = bufio.NewWriter(&b)
	}

	if wait {

		s := log.Spinnerf("Waiting for build to finish")
		defer s.End(false)
		if err := client.FollowBuildLog(buildName, stdout); err != nil {
			return errors.Wrapf(err, "unable to follow logs for %s", buildName)
		}

		if err := client.WaitForBuildToFinish(buildName); err != nil {
			return errors.Wrapf(err, "unable to build %s, error: %s", buildName, b.String())
		}
		s.End(true)
	}

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
func List(client *occlient.Client, applicationName string) (ComponentList, error) {

	var applicationSelector string
	if applicationName != "" {
		applicationSelector = fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)
	}

	// retrieve all the deployment configs that are associated with this application
	dcList, err := client.GetDeploymentConfigsFromSelector(applicationSelector)
	if err != nil {
		return ComponentList{}, errors.Wrapf(err, "unable to list components")
	}

	var components []Component

	// extract the labels we care about from each component
	for _, elem := range dcList {
		component, err := GetComponent(client, elem.Labels[componentlabels.ComponentLabel], applicationName, client.Namespace)
		if err != nil {
			return ComponentList{}, errors.Wrap(err, "Unable to get component")
		}
		components = append(components, component)

	}

	compoList := getMachineReadableFormatForList(components)
	return compoList, nil
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
		return "", "", errors.Wrapf(err, "unable to get source path for component %s", componentName)
	}

	sourcePath := deploymentConfig.ObjectMeta.Annotations[componentSourceURLAnnotation]
	sourceType := deploymentConfig.ObjectMeta.Annotations[ComponentSourceTypeAnnotation]

	if !validateSourceType(sourceType) {
		return "", "", fmt.Errorf("unsupported component source type %s", sourceType)
	}

	glog.V(4).Infof("Source for component %s is %s (%s)", componentName, sourcePath, sourceType)
	return sourceType, sourcePath, nil
}

// Update updates the requested component
// Parameters:
//	client: occlient instance
//	componentSettings: Component configuration
//	newSource: Location of component source resolved to absolute path
//	stdout: io pipe to write logs to
// Returns:
//	errors if any
func Update(client *occlient.Client, componentSettings config.LocalConfigInfo, newSource string, stdout io.Writer) error {
	// STEP 1. Create the common Object Meta for updating.

	componentName := componentSettings.GetName()
	applicationName := componentSettings.GetApplication()
	newSourceType := componentSettings.GetSourceType()
	newSourceRef := componentSettings.GetRef()
	componentImageType := componentSettings.GetType()
	cmpPorts := componentSettings.GetPorts()

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
	annotations[ComponentSourceTypeAnnotation] = string(newSourceType)

	// Parse componentImageType before adding to labels
	imageNS, imageName, imageTag, _, err := occlient.ParseImageName(componentImageType)
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

	// Retrieve the current DC in order to obtain what the current inputPorts are..
	currentDC, err := client.GetDeploymentConfigFromName(commonObjectMeta.Name)
	if err != nil {
		return errors.Wrapf(err, "unable to get DeploymentConfig %s", commonObjectMeta.Name)
	}

	foundCurrentDCContainer, err := occlient.FindContainer(currentDC.Spec.Template.Spec.Containers, commonObjectMeta.Name)
	if err != nil {
		return errors.Wrapf(err, "Unable to find container %s", commonObjectMeta.Name)
	}

	ports := foundCurrentDCContainer.Ports
	if len(cmpPorts) > 0 {
		ports, err = util.GetContainerPortsFromStrings(cmpPorts)
		if err != nil {
			return errors.Wrapf(err, "failed to apply component config %+v to component %s", componentSettings, commonObjectMeta.Name)
		}
	}

	commonImageMeta := occlient.CommonImageMeta{
		Namespace: imageNS,
		Name:      imageName,
		Tag:       imageTag,
		Ports:     ports,
	}

	// Generate the new DeploymentConfig
	resourceLimits := occlient.FetchContainerResourceLimits(foundCurrentDCContainer)
	resLts, err := occlient.GetResourceRequirementsFromCmpSettings(componentSettings)
	if err != nil {
		return errors.Wrap(err, "failed to update component")
	}
	if resLts != nil {
		resourceLimits = *resLts
	}

	updateComponentParams := occlient.UpdateComponentParams{
		CommonObjectMeta:  commonObjectMeta,
		ImageMeta:         commonImageMeta,
		ResourceLimits:    resourceLimits,
		EnvVars:           envVars,
		DcRollOutWaitCond: occlient.IsDCRolledOut,
		ExistingDC:        currentDC,
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
		bc, err := client.CreateBuildConfig(commonObjectMeta, componentImageType, newSource, newSourceRef, envVars)
		if err != nil {
			return errors.Wrapf(err, "unable to update BuildConfig  for %s component", componentName)
		}

		// Update / replace the current DeploymentConfig with a Git one (not SupervisorD!)
		glog.V(4).Infof("Updating the DeploymentConfig %s image to %s", namespacedOpenShiftObject, bc.Spec.Output.To.Name)

		// Update the image for git deployment to the BC built component image
		updateComponentParams.ImageMeta.Name = bc.Spec.Output.To.Name

		err = client.UpdateDCToGit(
			updateComponentParams,
			oldSourceType != string(config.GIT),
		)
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
		sourceURL := util.GenFileURL(newSource)
		annotations[componentSourceURLAnnotation] = sourceURL
		updateComponentParams.CommonObjectMeta.Annotations = annotations

		// Need to delete the old BuildConfig
		err = client.DeleteBuildConfig(commonObjectMeta)

		if err != nil {
			return errors.Wrapf(err, "unable to delete BuildConfig for %s component", componentName)
		}

		// Update the DeploymentConfig
		err = client.UpdateDCToSupervisor(
			updateComponentParams,
			newSourceType == config.LOCAL,
			true,
		)
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

			bc, err := client.GetBuildConfigFromName(namespacedOpenShiftObject)
			if err != nil {
				return errors.Wrap(err, "unable to get the BuildConfig file")
			}

			// Update the current DeploymentConfig with all config applied
			glog.V(4).Infof("Updating the DeploymentConfig %s image to %s", namespacedOpenShiftObject, bc.Spec.Output.To.Name)

			// Update the image for git deployment to the BC built component image
			updateComponentParams.ImageMeta.Name = bc.Spec.Output.To.Name

			err = client.UpdateDCToGit(
				updateComponentParams,
				oldSourceType != string(config.GIT),
			)
			if err != nil {
				return errors.Wrapf(err, "unable to update DeploymentConfig image for %s component", componentName)
			}

			// Build it
			err = Build(client, componentName, applicationName, true, stdout)

		} else if newSourceType == "local" || newSourceType == "binary" {

			// Update the sourceURL
			sourceURL := util.GenFileURL(newSource)
			annotations[componentSourceURLAnnotation] = sourceURL
			updateComponentParams.CommonObjectMeta.Annotations = annotations

			// Update the DeploymentConfig
			err = client.UpdateDCToSupervisor(
				updateComponentParams,
				newSourceType == config.LOCAL,
				false,
			)
			if err != nil {
				return errors.Wrapf(err, "unable to update DeploymentConfig for %s component", componentName)
			}
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
	for _, component := range componentList.Items {
		if component.Name == componentName {
			return true, nil
		}
	}
	return false, nil
}

// GetComponent provides component definition
func GetComponent(client *occlient.Client, componentName string, applicationName string, projectName string) (component Component, err error) {
	// Component Type
	componentType, err := GetComponentType(client, componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to get source type")
	}
	// Source
	_, path, err := GetComponentSource(client, componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to get source path")
	}
	// URL
	urlList, err := urlpkg.List(client, componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to get url list")
	}
	var urls []string
	for _, url := range urlList.Items {
		urls = append(urls, url.Name)
	}

	// Storage
	appStore, err := storage.List(client, componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to get storage list")
	}
	var storage []string
	for _, store := range appStore.Items {
		storage = append(storage, store.Name)
	}
	// Environment Variables
	DC, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to get DC list")
	}
	envVars, err := client.GetEnvVarsFromDC(DC)
	if err != nil {
		return component, errors.Wrap(err, "unable to get envVars list")
	}
	var filteredEnv []corev1.EnvVar
	for _, env := range envVars {
		if !strings.Contains(env.Name, "ODO") {
			filteredEnv = append(filteredEnv, env)
		}
	}

	if err != nil {
		return component, errors.Wrap(err, "unable to get envVars list")
	}

	linkedServices := make([]string, 0, 5)
	linkedComponents := make(map[string][]string)
	linkedSecretNames, err := GetComponentLinkedSecretNames(client, componentName, applicationName)
	if err != nil {
		return component, errors.Wrap(err, "unable to list linked secrets")
	}
	for _, secretName := range linkedSecretNames {
		secret, err := client.GetSecret(secretName, projectName)
		if err != nil {
			return component, errors.Wrapf(err, "unable to get info about secret %s", secretName)
		}
		componentName, containsComponentLabel := secret.Labels[componentlabels.ComponentLabel]
		if containsComponentLabel {
			if port, ok := secret.Annotations[occlient.ComponentPortAnnotationName]; ok {
				linkedComponents[componentName] = append(linkedComponents[componentName], port)
			}
		} else {
			linkedServices = append(linkedServices, secretName)
		}
	}

	component = getMachineReadableFormat(componentName, componentType)
	component.Spec.Source = path
	component.Spec.URL = urls
	component.Spec.Storage = storage
	component.Spec.Env = filteredEnv
	component.Status.LinkedComponents = linkedComponents
	component.Status.LinkedServices = linkedServices

	return component, nil
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

func getMachineReadableFormat(componentName, componentType string) Component {
	return Component{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Component",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentName,
		},
		Spec: ComponentSpec{
			Type: componentType,
		},
		Status: ComponentStatus{},
	}

}

// getMachineReadableFormatForList returns list of components in machine readable format
func getMachineReadableFormatForList(components []Component) ComponentList {
	if len(components) == 0 {
		components = []Component{}
	}
	return ComponentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ListMeta: metav1.ListMeta{},
		Items:    components,
	}

}
