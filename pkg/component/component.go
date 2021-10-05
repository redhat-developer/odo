package component

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/devfile/api/v2/pkg/devfile"

	v1 "k8s.io/api/apps/v1"

	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/service"

	"github.com/pkg/errors"
	"k8s.io/klog"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/sync"
	urlpkg "github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// componentSourceURLAnnotation is an source url from which component was build
const componentSourceURLAnnotation = "app.openshift.io/vcs-uri"
const ComponentSourceTypeAnnotation = "app.kubernetes.io/component-source-type"
const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1
const NotAvailable = "Not available"

const apiVersion = "odo.dev/v1alpha1"

var validSourceTypes = map[string]bool{
	"git":    true,
	"local":  true,
	"binary": true,
}

type componentAdapter struct {
	client occlient.Client
}

func (a componentAdapter) ExecCMDInContainer(componentInfo common.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return a.client.GetKubeClient().ExecCMDInContainer(componentInfo.ContainerName, componentInfo.PodName, cmd, stdout, stderr, stdin, tty)
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (a componentAdapter) ExtractProjectToComponent(componentInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {
	return a.client.GetKubeClient().ExtractProjectToComponent(componentInfo.ContainerName, componentInfo.PodName, targetPath, stdin)
}

// GetComponentDir returns source repo name
// Parameters:
//		path: source path
// Returns: directory name
func GetComponentDir(path string) (string, error) {
	retVal := ""
	if path != "" {
		retVal = filepath.Base(path)
	} else {
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
func GetDefaultComponentName(componentPath string, componentType string, existingComponentList ComponentList) (string, error) {
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
		prefix, err = GetComponentDir(componentPath)
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
		fmt.Sprintf("%s-%s", componentType, prefix),
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
	return validSourceTypes[sourceType]
}

// CreateFromGit inputPorts is the array containing the string port values
// inputPorts is the array containing the string port values
// envVars is the array containing the environment variables
func CreateFromGit(client *occlient.Client, params occlient.CreateArgs) error {

	// Create the labels
	labels := componentlabels.GetLabels(params.Name, params.ApplicationName, true)

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

	return nil
}

// GetComponentPorts provides slice of ports used by the component in the form port_no/protocol
func GetComponentPorts(client *occlient.Client, componentName string, applicationName string) (ports []string, err error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
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

	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
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

	// Create the labels to be used
	labels := componentlabels.GetLabels(params.Name, params.ApplicationName, true)

	// Parse componentImageType before adding to labels
	_, imageName, imageTag, _, err := occlient.ParseImageName(params.ImageName)
	if err != nil {
		return errors.Wrap(err, "unable to parse image name")
	}

	// save component type as label
	labels[componentlabels.ComponentTypeLabel] = imageName
	labels[componentlabels.ComponentTypeVersion] = imageTag

	// save source path as annotation
	annotations := map[string]string{}
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

	if params.Wait {
		// if wait flag is present then extract the podselector
		// use the podselector for calling WaitAndGetPod
		selectorLabels, err := util.NamespaceOpenShiftObject(labels[componentlabels.ComponentLabel], labels["app"])
		if err != nil {
			return err
		}

		podSelector := fmt.Sprintf("deploymentconfig=%s", selectorLabels)
		_, err = client.GetKubeClient().WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, "Waiting for component to start")
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// Delete whole component
func Delete(client *occlient.Client, wait bool, componentName, applicationName string) error {

	// Loading spinner
	s := log.Spinnerf("Deleting component %s", componentName)
	defer s.End(false)

	labels := componentlabels.GetLabels(componentName, applicationName, false)
	err := client.Delete(labels, wait)
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
	return retVal
}

// ApplyConfig applies the component config onto component dc
// Parameters:
//	client: occlient instance
//	componentConfig: Component configuration
//	envSpecificInfo: Component environment specific information, available if uses devfile
//  cmpExist: true if components exists in the cluster
//  isS2I: Legacy option. Set as true if you want to use the old S2I method as it differentiates slightly.
// Returns:
//	err: Errors if any else nil
func ApplyConfig(client *occlient.Client, envSpecificInfo envinfo.EnvSpecificInfo, stdout io.Writer, cmpExist bool, isS2I bool) (err error) {

	var configProvider localConfigProvider.LocalConfigProvider

	configProvider = &envSpecificInfo

	isRouteSupported := false
	isRouteSupported, err = client.IsRouteSupported()
	if err != nil {
		isRouteSupported = false
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		OCClient:            *client,
		IsRouteSupported:    isRouteSupported,
		LocalConfigProvider: configProvider,
	})

	return urlpkg.Push(urlpkg.PushParameters{
		LocalConfigProvider: configProvider,
		URLClient:           urlClient,
		IsRouteSupported:    isRouteSupported,
	})
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
//	show determines whether or not to show the log (passed in by po.show argument within /cmd)
// Returns
//	Error if any
func PushLocal(client *occlient.Client, componentName string, applicationName string, path string, out io.Writer, files []string, delFiles []string, isForcePush bool, globExps []string, show bool) error {
	klog.V(4).Infof("PushLocal: componentName: %s, applicationName: %s, path: %s, files: %s, delFiles: %s, isForcePush: %+v", componentName, applicationName, path, files, delFiles, isForcePush)

	// Edge case: check to see that the path is NOT empty.
	emptyDir, err := util.IsEmpty(path)
	if err != nil {
		return errors.Wrapf(err, "Unable to check directory: %s", path)
	} else if emptyDir {
		return errors.New(fmt.Sprintf("Directory / file %s is empty", path))
	}

	// Find DeploymentConfig for component
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)
	dc, err := client.GetDeploymentConfigFromSelector(componentSelector)
	if err != nil {
		return errors.Wrap(err, "unable to get deployment for component")
	}
	// Find Pod for component
	podSelector := fmt.Sprintf("deploymentconfig=%s", dc.Name)

	// Wait for Pod to be in running state otherwise we can't sync data to it.
	pod, err := client.GetKubeClient().WaitAndGetPodWithEvents(podSelector, corev1.PodRunning, "Waiting for component to start")
	if err != nil {
		return errors.Wrapf(err, "error while waiting for pod  %s", podSelector)
	}

	// Get S2I Source/Binary Path from Pod Env variables created at the time of component create
	s2iSrcPath := getEnvFromPodEnvs(occlient.EnvS2ISrcOrBinPath, pod.Spec.Containers[0].Env)
	if s2iSrcPath == "" {
		s2iSrcPath = occlient.DefaultS2ISrcOrBinPath
	}
	targetPath := fmt.Sprintf("%s/src", s2iSrcPath)

	// Sync the files to the pod
	s := log.Spinner("Syncing files to the component")
	defer s.End(false)

	// If there are files identified as deleted, propagate them to the component pod
	if len(delFiles) > 0 {
		klog.V(4).Infof("propagating deletion of files %s to pod", strings.Join(delFiles, " "))
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

	if !isForcePush {
		if len(files) == 0 && len(delFiles) == 0 {
			// nothing to push
			s.End(true)
			return nil
		}
	}

	adapter := componentAdapter{
		client: *client,
	}

	if isForcePush || len(files) > 0 {
		klog.V(4).Infof("Copying files %s to pod", strings.Join(files, " "))
		compInfo := common.ComponentInfo{
			PodName: pod.Name,
		}
		err = sync.CopyFile(adapter, path, compInfo, targetPath, files, globExps, util.IndexerRet{})
		if err != nil {
			s.End(false)
			return errors.Wrap(err, "unable push files to pod")
		}
	}
	s.End(true)

	if show {
		s = log.SpinnerNoSpin("Building component")
	} else {
		s = log.Spinner("Building component")
	}

	// We will use the assemble-and-restart script located within the supervisord container we've created
	cmdArr := []string{"/opt/odo/bin/assemble-and-restart"}

	compInfo := common.ComponentInfo{
		PodName: pod.Name,
	}

	err = common.ExecuteCommand(adapter, compInfo, cmdArr, show, nil, nil)

	if err != nil {
		s.End(false)
		return errors.Wrap(err, "unable to execute assemble script")
	}

	s.End(true)

	return nil
}

// Build component from BuildConfig.
// If 'wait' is true than it waits for build to successfully complete.
// If 'wait' is false than this function won't return error even if build failed.
// 'show' will determine whether or not the log will be shown to the user (while building)
func Build(client *occlient.Client, componentName string, applicationName string, wait bool, stdout io.Writer, show bool) error {

	// Try to grab the preference in order to set a timeout.. but if not, we’ll use the default.
	buildTimeout := preference.DefaultBuildTimeout * time.Second
	cfg, configReadErr := preference.New()
	if configReadErr != nil {
		klog.V(4).Info(errors.Wrap(configReadErr, "unable to read config file"))
	} else {
		buildTimeout = time.Duration(cfg.GetBuildTimeout()) * time.Second
	}

	// Loading spinner
	// No loading spinner if we're showing the logging output
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
	if !log.IsDebug() && !show {
		stdout = bufio.NewWriter(&b)
	}

	if wait {

		if show {
			s = log.SpinnerNoSpin("Waiting for build to finish")
		} else {
			s = log.Spinner("Waiting for build to finish")
		}

		defer s.End(false)

		if err := client.WaitForBuildToFinish(buildName, stdout, buildTimeout); err != nil {
			return errors.Wrapf(err, "unable to build %s, error: %s", buildName, b.String())
		}
		s.End(true)
	}

	return nil
}

// Deploy deploys the component
// it starts a new deployment and wait for the new dc to be available
// desiredRevision is the desired version of the deployment config to wait for
func Deploy(client *occlient.Client, params occlient.CreateArgs, desiredRevision int64) error {

	// Loading spinner
	s := log.Spinnerf("Deploying component %s", params.Name)
	defer s.End(false)

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(params.Name, params.ApplicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// start the deployment
	// the build must be finished before this call and the new image must be successfully updated
	_, err = client.StartDeployment(namespacedOpenShiftObject)
	if err != nil {
		return errors.Wrapf(err, "unable to create DeploymentConfig for %s", namespacedOpenShiftObject)
	}

	// Watch / wait for deployment config to update annotations
	_, err = client.WaitAndGetDC(namespacedOpenShiftObject, desiredRevision, occlient.OcUpdateTimeout, occlient.IsDCRolledOut)
	if err != nil {
		return errors.Wrapf(err, "unable to wait for DeploymentConfig %s to update", namespacedOpenShiftObject)
	}

	s.End(true)

	return nil
}

// GetComponentNames retrieves the names of the components in the specified application
func GetComponentNames(client *occlient.Client, applicationName string) ([]string, error) {
	components, err := GetPushedComponents(client, applicationName)
	if err != nil {
		return []string{}, err
	}
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// ListDevfileComponents returns the devfile component matching a selector.
// The selector could be about selecting components part of an application.
// There are helpers in "applabels" package for this.
func ListDevfileComponents(client *occlient.Client, selector string) (ComponentList, error) {

	var deploymentList []v1.Deployment
	var components []Component

	// retrieve all the deployments that are associated with this application
	deploymentList, err := client.GetKubeClient().GetDeploymentFromSelector(selector)
	if err != nil {
		return ComponentList{}, errors.Wrapf(err, "unable to list components")
	}

	// create a list of object metadata based on the component and application name (extracted from Deployment labels)
	for _, elem := range deploymentList {
		component, err := GetComponent(client, elem.Labels[componentlabels.ComponentLabel], elem.Labels[applabels.ApplicationLabel], client.Namespace)
		if err != nil {
			return ComponentList{}, errors.Wrap(err, "Unable to get component")
		}

		if !reflect.ValueOf(component).IsZero() {
			components = append(components, component)
		}

	}

	compoList := newComponentList(components)
	return compoList, nil
}

// List lists all s2i and devfile components in active application
func List(client *occlient.Client, applicationSelector string) (ComponentList, error) {
	var components []Component
	devfileList, err := ListDevfileComponents(client, applicationSelector)
	if err != nil {
		return ComponentList{}, nil
	}
	components = append(components, devfileList.Items...)
	return newComponentList(components), nil
}

// GetComponentFromDevfile extracts component's metadata from the specified env info if it exists
func GetComponentFromDevfile(info *envinfo.EnvSpecificInfo) (Component, parser.DevfileObj, error) {
	if info.Exists() {
		devfile, err := parser.Parse(info.GetDevfilePath())
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		component, err := getComponentFrom(info, GetComponentTypeFromDevfileMetadata(devfile.Data.GetMetadata()))
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		components, err := devfile.Data.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		for _, cmp := range components {
			if cmp.Container != nil {
				for _, env := range cmp.Container.Env {
					component.Spec.Env = append(component.Spec.Env, corev1.EnvVar{Name: env.Name, Value: env.Value})
				}
			}
		}

		return component, devfile, nil
	}
	return Component{}, parser.DevfileObj{}, nil
}

// GetComponentTypeFromDevfileMetadata returns component type from the devfile metadata;
// it could either be projectType or language, if neither of them are set, return 'Not available'
func GetComponentTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var componentType string
	if metadata.ProjectType != "" {
		componentType = metadata.ProjectType
	} else if metadata.Language != "" {
		componentType = metadata.Language
	} else {
		componentType = NotAvailable
	}
	return componentType
}

func getComponentFrom(info localConfigProvider.LocalConfigProvider, componentType string) (Component, error) {
	if info.Exists() {
		component := newComponentWithType(info.GetName(), componentType)

		component.Namespace = info.GetNamespace()

		component.Spec = ComponentSpec{
			App:   info.GetApplication(),
			Type:  componentType,
			Ports: []string{fmt.Sprintf("%d", info.GetDebugPort())},
		}

		urls, err := info.ListURLs()
		if err != nil {
			return Component{}, err
		}
		if len(urls) > 0 {
			for _, url := range urls {
				component.Spec.URL = append(component.Spec.URL, url.Name)
			}
		}

		return component, nil
	}
	return Component{}, nil
}

func ListDevfileComponentsInPath(client kclient.ClientInterface, paths []string) ([]Component, error) {
	var components []Component
	var err error
	for _, path := range paths {
		err = filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
			// we check for .odo/env/env.yaml folder first and then find devfile.yaml, this could be changed
			// TODO: optimise this
			if f != nil && strings.Contains(f.Name(), ".odo") {
				// lets find if there is a devfile and an env.yaml
				dir := filepath.Dir(path)
				data, err := envinfo.NewEnvSpecificInfo(dir)
				if err != nil {
					return err
				}

				// if the .odo folder doesn't contain a proper env file
				if data.GetName() == "" || data.GetApplication() == "" || data.GetNamespace() == "" {
					return nil
				}

				// we just want to confirm if the devfile is correct
				_, err = parser.ParseDevfile(parser.ParserArgs{
					Path: filepath.Join(dir, "devfile.yaml"),
				})
				if err != nil {
					return err
				}
				con, _ := filepath.Abs(filepath.Dir(path))

				comp := NewComponent(data.GetName())
				comp.Status.State = StateTypeUnknown
				comp.Spec.App = data.GetApplication()
				comp.Namespace = data.GetNamespace()
				comp.Status.Context = con

				// since the config file maybe belong to a component of a different project
				if client != nil {
					client.SetNamespace(data.GetNamespace())
					deployment, err := client.GetOneDeployment(comp.Name, comp.Spec.App)
					if err != nil {
						comp.Status.State = StateTypeNotPushed
					} else if deployment != nil {
						comp.Status.State = StateTypePushed
					}
				}

				components = append(components, comp)
			}

			return nil
		})

	}
	return components, err
}

// Exists checks whether a component with the given name exists in the current application or not
// componentName is the component name to perform check for
// The first returned parameter is a bool indicating if a component with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func Exists(client *occlient.Client, componentName, applicationName string) (bool, error) {
	deploymentName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to create namespaced name")
	}
	deployment, _ := client.GetDeploymentConfigFromName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

func GetComponentState(client *occlient.Client, componentName, applicationName string) State {
	// first check if a deployment exists
	c, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil {
		return StateTypeUnknown
	}
	if c != nil {
		return StateTypePushed
	}
	return StateTypeNotPushed
}

// GetComponent provides component definition
func GetComponent(client *occlient.Client, componentName string, applicationName string, projectName string) (component Component, err error) {
	return getRemoteComponentMetadata(client, componentName, applicationName, true, true)
}

// getRemoteComponentMetadata provides component metadata from the cluster
func getRemoteComponentMetadata(client *occlient.Client, componentName string, applicationName string, getUrls, getStorage bool) (component Component, err error) {
	fromCluster, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil || fromCluster == nil {
		return Component{}, errors.Wrapf(err, "unable to get remote metadata for %s component", componentName)
	}

	// Component Type
	componentType, err := fromCluster.GetType()
	if err != nil {
		return component, errors.Wrap(err, "unable to get source type")
	}

	// init component
	component = newComponentWithType(componentName, componentType)

	// URL
	if getUrls {
		urls, err := fromCluster.GetURLs()
		if err != nil {
			return Component{}, err
		}
		component.Spec.URLSpec = urls
		urlsNb := len(urls)
		if urlsNb > 0 {
			res := make([]string, 0, urlsNb)
			for _, url := range urls {
				res = append(res, url.Name)
			}
			component.Spec.URL = res
		}
	}

	// Storage
	if getStorage {
		appStore, err := fromCluster.GetStorage()
		if err != nil {
			return Component{}, errors.Wrap(err, "unable to get storage list")
		}

		component.Spec.StorageSpec = appStore
		var storageList []string
		for _, store := range appStore {
			storageList = append(storageList, store.Name)
		}
		component.Spec.Storage = storageList
	}

	// Environment Variables
	envVars := fromCluster.GetEnvVars()
	var filteredEnv []corev1.EnvVar
	for _, env := range envVars {
		if !strings.Contains(env.Name, "ODO") {
			filteredEnv = append(filteredEnv, env)
		}
	}

	// Secrets
	linkedSecrets := fromCluster.GetLinkedSecrets()
	err = setLinksServiceNames(client, linkedSecrets, componentlabels.GetSelector(componentName, applicationName))
	if err != nil {
		return Component{}, fmt.Errorf("unable to get name of services: %w", err)
	}
	component.Status.LinkedServices = linkedSecrets

	// Annotations
	component.Annotations = fromCluster.GetAnnotations()

	// Labels
	component.Labels = fromCluster.GetLabels()

	component.Namespace = client.Namespace
	component.Spec.App = applicationName
	component.Spec.Env = filteredEnv
	component.Status.State = StateTypePushed

	return component, nil
}

// setLinksServiceNames sets the service name of the links from the info in ServiceBindingRequests present in the cluster
func setLinksServiceNames(client *occlient.Client, linkedSecrets []SecretMount, selector string) error {
	ok, err := client.GetKubeClient().IsServiceBindingSupported()
	if err != nil {
		return fmt.Errorf("unable to check if service binding is supported: %w", err)
	}

	serviceBindings := map[string]string{}
	if ok {
		// service binding operator is installed on the cluster
		list, err := client.GetKubeClient().ListDynamicResource(kclient.ServiceBindingGroup, kclient.ServiceBindingVersion, kclient.ServiceBindingResource)
		if err != nil || list == nil {
			return err
		}

		for _, u := range list.Items {
			var sbr servicebinding.ServiceBinding
			js, err := u.MarshalJSON()
			if err != nil {
				return err
			}
			err = json.Unmarshal(js, &sbr)
			if err != nil {
				return err
			}
			services := sbr.Spec.Services
			if len(services) != 1 {
				return errors.New("the ServiceBinding resource should define only one service")
			}
			service := services[0]
			if service.Kind == "Service" {
				serviceBindings[sbr.Status.Secret] = service.Name
			} else {
				serviceBindings[sbr.Status.Secret] = service.Kind + "/" + service.Name
			}
		}
	} else {
		// service binding operator is not installed
		// get the secrets instead of the service binding objects to retrieve the link data
		secrets, err := client.GetKubeClient().ListSecrets(selector)
		if err != nil {
			return err
		}

		// get the services to get their names against the component names
		services, err := client.GetKubeClient().ListServices("")
		if err != nil {
			return err
		}

		serviceCompMap := make(map[string]string)
		for _, gotService := range services {
			serviceCompMap[gotService.Labels[componentlabels.ComponentLabel]] = gotService.Name
		}

		for _, secret := range secrets {
			serviceName, serviceOK := secret.Labels[service.ServiceLabel]
			_, linkOK := secret.Labels[service.LinkLabel]
			serviceKind, serviceKindOK := secret.Labels[service.ServiceKind]
			if serviceKindOK && serviceOK && linkOK {
				if serviceKind == "Service" {
					if _, ok := serviceBindings[secret.Name]; !ok {
						serviceBindings[secret.Name] = serviceCompMap[serviceName]
					}
				} else {
					// service name is stored as kind-name in the labels
					parts := strings.SplitN(serviceName, "-", 2)
					if len(parts) < 2 {
						continue
					}

					serviceName = fmt.Sprintf("%v/%v", parts[0], parts[1])
					if _, ok := serviceBindings[secret.Name]; !ok {
						serviceBindings[secret.Name] = serviceName
					}
				}
			}
		}
	}

	for i, linkedSecret := range linkedSecrets {
		linkedSecrets[i].ServiceName = serviceBindings[linkedSecret.SecretName]
	}
	return nil
}

// GetLogs follow the DeploymentConfig logs if follow is set to true
func GetLogs(client *occlient.Client, componentName string, applicationName string, follow bool, stdout io.Writer) error {

	// Namespace the component
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	// Retrieve the logs
	err = client.DisplayDeploymentConfigLog(namespacedOpenShiftObject, follow)
	if err != nil {
		return err
	}

	return nil
}
