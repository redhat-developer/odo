package component

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/devfile/location"

	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/service"
	urlpkg "github.com/redhat-developer/odo/pkg/url"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const NotAvailable = "Not available"

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

// ListDevfileStacks returns the devfile component matching a selector.
// The selector could be about selecting components part of an application.
// There are helpers in "applabels" package for this.
func ListDevfileStacks(client kclient.ClientInterface, selector string) (ComponentList, error) {

	var deploymentList []v1.Deployment
	var components []Component

	// retrieve all the deployments that are associated with this application
	deploymentList, err := client.GetDeploymentFromSelector(selector)
	if err != nil {
		return ComponentList{}, fmt.Errorf("unable to list components: %w", err)
	}

	// create a list of object metadata based on the component and application name (extracted from Deployment labels)
	for _, elem := range deploymentList {
		component, err := GetComponent(client, elem.Labels[componentlabels.KubernetesInstanceLabel], elem.Labels[applabels.ApplicationLabel], client.GetCurrentNamespace())
		if err != nil {
			return ComponentList{}, fmt.Errorf("Unable to get component: %w", err)
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
	devfileList, err := ListDevfileStacks(client, applicationSelector)
	if err != nil {
		return ComponentList{}, nil
	}
	return newComponentList(devfileList.Items), nil
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

func ListDevfileStacksInPath(client kclient.ClientInterface, paths []string) ([]Component, error) {
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
		return false, fmt.Errorf("unable to create namespaced name: %w", err)
	}
	deployment, _ := client.GetDeploymentByName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

// GetComponent provides component definition
func GetComponent(client kclient.ClientInterface, componentName string, applicationName string, projectName string) (component Component, err error) {
	return getRemoteComponentMetadata(client, componentName, applicationName, true, true)
}

// getRemoteComponentMetadata provides component metadata from the cluster
func getRemoteComponentMetadata(client kclient.ClientInterface, componentName string, applicationName string, getUrls, getStorage bool) (Component, error) {
	fromCluster, err := GetPushedComponent(client, componentName, applicationName)
	if err != nil {
		return Component{}, fmt.Errorf("unable to get remote metadata for %s component: %w", componentName, err)
	}
	if fromCluster == nil {
		return Component{}, nil
	}

	// Component Type
	componentType, err := fromCluster.GetType()
	if err != nil {
		return Component{}, fmt.Errorf("unable to get source type: %w", err)
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
			return Component{}, fmt.Errorf("unable to get storage list: %w", e)
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
	component.Status.State = StateTypePushed

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
		return nil, fmt.Errorf("the component %s doesn't exist on the cluster", componentName)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("unable to show logs, component is not in running state. current status=%v", pod.Status.Phase)
	}

	containerName := command.Exec.Component

	return client.GetPodLogs(pod.Name, containerName, follow)
}
