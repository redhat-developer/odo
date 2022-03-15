package component

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/devfile/location"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/service"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"github.com/redhat-developer/odo/pkg/util"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1
const NotAvailable = "Not available"
const apiVersion = "odo.dev/v1alpha1"

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
func GetDefaultComponentName(cfg preference.Client, componentPath string, componentType string, existingComponentList ComponentList) (string, error) {
	var prefix string
	var err error

	// Get component names from component list
	var existingComponentNames []string
	for _, component := range existingComponentList.Items {
		existingComponentNames = append(existingComponentNames, component.Name)
	}

	// Create a random generated name for the component to use within Kubernetes
	prefix, err = GetComponentDir(componentPath)
	if err != nil {
		return "", errors.Wrap(err, "unable to generate random component name")
	}
	prefix = util.TruncateString(prefix, componentRandomNamePartsMaxLen)

	// Generate unique name for the component using prefix and unique random suffix
	componentName, err := dfutil.GetRandomName(
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

// ApplyConfig applies the component config onto component deployment
// Parameters:
//	client: kclient instance
//	componentConfig: Component configuration
//	envSpecificInfo: Component environment specific information, available if uses devfile
// Returns:
//	err: Errors if any else nil
func ApplyConfig(client kclient.ClientInterface, envSpecificInfo envinfo.EnvSpecificInfo) (err error) {
	isRouteSupported := false
	isRouteSupported, err = client.IsRouteSupported()
	if err != nil {
		isRouteSupported = false
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		Client:              client,
		IsRouteSupported:    isRouteSupported,
		LocalConfigProvider: &envSpecificInfo,
	})

	return urlpkg.Push(urlpkg.PushParameters{
		LocalConfigProvider: &envSpecificInfo,
		URLClient:           urlClient,
		IsRouteSupported:    isRouteSupported,
	})
}

// ListDevfileComponents returns the devfile component matching a selector.
// The selector could be about selecting components part of an application.
// There are helpers in "applabels" package for this.
func ListDevfileComponents(client kclient.ClientInterface, selector string) (ComponentList, error) {

	var deploymentList []v1.Deployment
	var components []Component

	// retrieve all the deployments that are associated with this application
	deploymentList, err := client.GetDeploymentFromSelector(selector)
	if err != nil {
		return ComponentList{}, errors.Wrapf(err, "unable to list components")
	}

	// create a list of object metadata based on the component and application name (extracted from Deployment labels)
	for _, elem := range deploymentList {
		component, err := GetComponent(client, elem.Labels[componentlabels.KubernetesInstanceLabel], elem.Labels[applabels.ApplicationLabel], client.GetCurrentNamespace())
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

// List lists all the devfile components in active application
func List(client kclient.ClientInterface, applicationSelector string) (ComponentList, error) {
	devfileList, err := ListDevfileComponents(client, applicationSelector)
	if err != nil {
		return ComponentList{}, nil
	}
	return newComponentList(devfileList.Items), nil
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

// GetProjectTypeFromDevfileMetadata returns component type from the devfile metadata
func GetProjectTypeFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var projectType string
	if metadata.ProjectType != "" {
		projectType = metadata.ProjectType
	} else {
		projectType = NotAvailable
	}
	return projectType
}

// GetLanguageFromDevfileMetadata returns component type from the devfile metadata
func GetLanguageFromDevfileMetadata(metadata devfile.DevfileMetadata) string {
	var language string
	if metadata.Language != "" {
		language = metadata.Language
	} else {
		language = NotAvailable
	}
	return language
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
					Path: location.DevfileLocation(dir),
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
func Exists(client kclient.ClientInterface, componentName, applicationName string) (bool, error) {
	deploymentName, err := dfutil.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to create namespaced name")
	}
	deployment, _ := client.GetDeploymentByName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

// GetComponentState ...
func GetComponentState(client kclient.ClientInterface, componentName, applicationName string) string {
	// Check to see if the deployment has been pushed or not
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
func GetComponent(client kclient.ClientInterface, componentName string, applicationName string, projectName string) (component Component, err error) {
	return getRemoteComponentMetadata(client, componentName, applicationName, true, true)
}

// getRemoteComponentMetadata provides component metadata from the cluster
func getRemoteComponentMetadata(client kclient.ClientInterface, componentName string, applicationName string, getUrls, getStorage bool) (Component, error) {
	fromCluster, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil || fromCluster == nil {
		return Component{}, errors.Wrapf(err, "unable to get remote metadata for %s component", componentName)
	}

	// Component Type
	componentType, err := fromCluster.GetType()
	if err != nil {
		return Component{}, errors.Wrap(err, "unable to get source type")
	}

	// init component
	component := newComponentWithType(componentName, componentType)

	// URL
	if getUrls {
		urls, e := fromCluster.GetURLs()
		if e != nil {
			return Component{}, e
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
		appStore, e := fromCluster.GetStorage()
		if e != nil {
			return Component{}, errors.Wrap(e, "unable to get storage list")
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

	// Mark the component status as pushed
	component.Status.State = componentlabels.ComponentPushedName

	// Labels
	component.Labels = fromCluster.GetLabels()
	component.Namespace = client.GetCurrentNamespace()
	component.Spec.App = applicationName
	component.Spec.Env = filteredEnv

	return component, nil
}

// setLinksServiceNames sets the service name of the links from the info in ServiceBindingRequests present in the cluster
func setLinksServiceNames(client kclient.ClientInterface, linkedSecrets []SecretMount, selector string) error {
	ok, err := client.IsServiceBindingSupported()
	if err != nil {
		return fmt.Errorf("unable to check if service binding is supported: %w", err)
	}

	serviceBindings := map[string]string{}
	if ok {
		// service binding operator is installed on the cluster
		list, err := client.ListDynamicResource(kclient.ServiceBindingGroup, kclient.ServiceBindingVersion, kclient.ServiceBindingResource)
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
		secrets, err := client.ListSecrets(selector)
		if err != nil {
			return err
		}

		// get the services to get their names against the component names
		services, err := client.ListServices("")
		if err != nil {
			return err
		}

		serviceCompMap := make(map[string]string)
		for _, gotService := range services {
			serviceCompMap[gotService.Labels[componentlabels.KubernetesInstanceLabel]] = gotService.Name
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

// GetOnePod gets a pod using the component and app name
func GetOnePod(client kclient.ClientInterface, componentName string, appName string) (*corev1.Pod, error) {
	return client.GetOnePodFromSelector(componentlabels.GetSelector(componentName, appName))
}

// ComponentExists checks whether a deployment by the given name exists in the given app
func ComponentExists(client kclient.ClientInterface, name string, app string) (bool, error) {
	deployment, err := client.GetOneDeployment(name, app)
	if _, ok := err.(*kclient.DeploymentNotFoundError); ok {
		klog.V(2).Infof("Deployment %s not found for belonging to the %s app ", name, app)
		return false, nil
	}
	return deployment != nil, err
}

// Log returns log from component
func Log(client kclient.ClientInterface, componentName string, appName string, follow bool, command v1alpha2.Command) (io.ReadCloser, error) {

	pod, err := GetOnePod(client, componentName, appName)
	if err != nil {
		return nil, errors.Errorf("the component %s doesn't exist on the cluster", componentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, errors.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return client.GetPodLogs(pod.Name, containerName, follow)
}

// Delete deletes the component
func Delete(kubeClient kclient.ClientInterface, devfileObj parser.DevfileObj, componentName string, appName string, labels map[string]string, show bool, wait bool) error {
	if labels == nil {
		return fmt.Errorf("cannot delete with labels being nil")
	}
	log.Printf("Gathering information for component: %q", componentName)
	podSpinner := log.Spinner("Checking status for component")
	defer podSpinner.End(false)

	pod, err := GetOnePod(kubeClient, componentName, appName)
	if kerrors.IsForbidden(err) {
		klog.V(2).Infof("Resource for %s forbidden", componentName)
		// log the error if it failed to determine if the component exists due to insufficient RBACs
		podSpinner.End(false)
		log.Warningf("%v", err)
		return nil
	} else if e, ok := err.(*kclient.PodNotFoundError); ok {
		podSpinner.End(false)
		log.Warningf("%v", e)
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "unable to determine if component %s exists", componentName)
	}

	podSpinner.End(true)

	// if there are preStop events, execute them before deleting the deployment
	if libdevfile.HasPreStopEvents(devfileObj) {
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("unable to execute preStop events, pod for component %s is not running", componentName)
		}
		log.Infof("\nExecuting %s event commands for component %s", libdevfile.PreStop, componentName)
		err = libdevfile.ExecPreStopEvents(devfileObj, componentName, NewExecHandler(kubeClient, pod.Name, show))
		if err != nil {
			return err
		}
	}

	log.Infof("\nDeleting component %s", componentName)
	spinner := log.Spinner("Deleting Kubernetes resources for component")
	defer spinner.End(false)

	err = kubeClient.Delete(labels, wait)
	if err != nil {
		return err
	}

	spinner.End(true)
	log.Successf("Successfully deleted component")
	return nil
}
