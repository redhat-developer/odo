package component

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	v1 "k8s.io/api/apps/v1"

	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/service"

	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	"github.com/redhat-developer/odo/pkg/util"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const componentRandomNamePartsMaxLen = 12
const componentNameMaxRetries = 3
const componentNameMaxLen = -1
const NotAvailable = "Not available"

const apiVersion = "odo.dev/v1alpha1"

type componentClient struct {
	client kclient.ClientInterface
}

func NewClient(client kclient.ClientInterface) Client {
	return componentClient{
		client: client,
	}
}

// GetComponentFullDescription gets the complete description of the component from cluster
func (c componentClient) GetComponentFullDescription(envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*Component, error) {
	componentDesc, devfileObj, err := getComponentFromDevfile(envInfo)
	if err != nil {
		return &componentDesc, err
	}
	envInfo.SetDevfileObj(devfileObj)

	componentDesc.Status.State = StateTypeUnknown
	if c.client != nil {
		componentDesc.Status.State = getComponentState(componentName, applicationName, c.client)
	}
	// If the component is not pushed, and devfile.yaml is not accessible, return early
	if componentDesc.Status.State == StateTypeNotPushed && devfileObj.Data == nil {
		componentDesc.Name = componentName
		return &componentDesc, nil
	}

	// If devfile.yaml is accessible, or if the component is pushed, continue
	var deployment *v1.Deployment
	if componentDesc.Status.State == StateTypePushed {
		componentDescFromCluster, e := getRemoteComponentMetadata(componentName, applicationName, false, false, c.client)
		if e != nil {
			return &componentDesc, e
		}
		deployment, err = c.client.GetOneDeployment(componentName, applicationName)
		if err != nil {
			return &componentDesc, err
		}
		componentDesc.Spec.Env = componentDescFromCluster.Spec.Env
		componentDesc.Spec.Type = componentDescFromCluster.Spec.Type
		componentDesc.Status.LinkedServices = componentDescFromCluster.Status.LinkedServices
	}

	// Fill empty fields
	componentDesc.fillEmptyFields(componentName, applicationName, projectName)

	// Get links
	var configLinks []string
	configLinks, err = service.ListDevfileLinks(devfileObj, context)
	if err != nil {
		return &componentDesc, err
	}
	for _, link := range configLinks {
		found := false
		for _, linked := range componentDesc.Status.LinkedServices {
			if linked.ServiceName == link {
				found = true
				break
			}
		}
		if !found {
			componentDesc.Status.LinkedServices = append(componentDesc.Status.LinkedServices, SecretMount{
				ServiceName: link,
			})
		}
	}

	// Obtain URLs information
	var urls urlpkg.URLList
	var routeSupported bool
	if c.client != nil {
		// we assume if there was an error then the cluster is not connected
		routeSupported, _ = c.client.IsRouteSupported()
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		LocalConfigProvider: envInfo,
		Client:              c.client,
		IsRouteSupported:    routeSupported,
		Deployment:          deployment,
	})

	urls, err = urlClient.List()
	if err != nil {
		log.Warningf("URLs couldn't not be retrieved: %v", err)
	}
	// We explicitly do not fill in Spec.URL for describe
	componentDesc.Spec.URLSpec = urls.Items

	// Obtain Storage information
	var storages storage.StorageList
	storageClient := storage.NewClient(storage.ClientOptions{
		Client:              c.client,
		LocalConfigProvider: envInfo,
		Deployment:          deployment,
	})

	// We explicitly do not fill in Spec.Storage for describe
	storages, err = storageClient.List()
	if err != nil {
		log.Warningf("Storages couldn't not be retrieved: %v", err)
	}
	componentDesc.Spec.StorageSpec = storages.Items

	return &componentDesc, nil
}

// GetComponentNames retrieves the names of the components in the specified application
func (c componentClient) GetComponentNames(applicationName string) ([]string, error) {
	components, err := getPushedComponents(applicationName, c.client)
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

// List returns the devfile component matching a selector.
// The selector could be about selecting components part of an application.
// There are helpers in "applabels" package for this.
func (c componentClient) List(selector string) (ComponentList, error) {

	var deploymentList []v1.Deployment
	var components []Component

	// retrieve all the deployments that are associated with this application
	deploymentList, err := c.client.GetDeploymentFromSelector(selector)
	if err != nil {
		return ComponentList{}, errors.Wrapf(err, "unable to list components")
	}

	// create a list of object metadata based on the component and application name (extracted from Deployment labels)
	for _, elem := range deploymentList {
		component, err := getRemoteComponentMetadata(elem.Labels[componentlabels.ComponentLabel], elem.Labels[applabels.ApplicationLabel], true, true, c.client)
		if err != nil {
			return ComponentList{}, errors.Wrap(err, "Unable to get component")
		}

		if !reflect.ValueOf(component).IsZero() {
			// This is a workaround to avoid having Storage and URL specs in the JSON output;
			// We do not want this information in the output; and JSON marshaller will omit this since it's empty
			component.Spec.URLSpec = []url.URL{}
			component.Spec.StorageSpec = []storage.Storage{}
			components = append(components, component)
		}

	}

	compoList := newComponentList(components)
	return compoList, nil
}

// ListComponentsInPath lists all the components in a path
func (c componentClient) ListComponentsInPath(paths []string) ([]Component, error) {
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
				if c.client != nil {
					c.client.SetNamespace(data.GetNamespace())
					deployment, err := c.client.GetOneDeployment(comp.Name, comp.Spec.App)
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
// The second returned parameter is the error that might occur while execution
func (c componentClient) Exists(componentName, applicationName string) (bool, error) {
	deploymentName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to create namespaced name")
	}
	deployment, _ := c.client.GetDeploymentByName(deploymentName)
	if deployment != nil {
		return true, nil
	}
	return false, nil
}

// GetLinkedServicesSecretData gets the secrets/environment variables that are passed by the linkedsecret
func (c componentClient) GetLinkedServicesSecretData(namespace, secretName string) (map[string][]byte, error) {
	secrets, err := c.client.GetSecret(secretName, namespace)
	if err != nil {
		return map[string][]byte{}, err
	}
	return secrets.Data, nil
}

// getRemoteComponentMetadata provides component metadata from the cluster
func getRemoteComponentMetadata(componentName string, applicationName string, getUrls, getStorage bool, kubeclient kclient.ClientInterface) (Component, error) {
	fromCluster, err := getPushedComponent(componentName, applicationName, kubeclient)
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
	err = setLinksServiceNames(kubeclient, linkedSecrets, componentlabels.GetSelector(componentName, applicationName))
	if err != nil {
		return Component{}, fmt.Errorf("unable to get name of services: %w", err)
	}
	component.Status.LinkedServices = linkedSecrets

	// Annotations
	component.Annotations = fromCluster.GetAnnotations()

	// Labels
	component.Labels = fromCluster.GetLabels()

	component.Namespace = kubeclient.GetCurrentNamespace()
	component.Spec.App = applicationName
	component.Spec.Env = filteredEnv
	component.Status.State = StateTypePushed

	return component, nil
}

func newPushedComponent(applicationName string, p provider, kubeclient kclient.ClientInterface, storageClient storage.Client, urlClient url.Client) PushedComponent {
	return &defaultPushedComponent{
		application:   applicationName,
		provider:      p,
		client:        kubeclient,
		storageClient: storageClient,
		urlClient:     urlClient,
	}
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

func getComponentState(componentName, applicationName string, kubeclient kclient.ClientInterface) State {
	// first check if a deployment exists
	pushedComponent, err := getPushedComponent(componentName, applicationName, kubeclient)
	if err != nil {
		return StateTypeUnknown
	}
	if pushedComponent != nil {
		return StateTypePushed
	}
	return StateTypeNotPushed
}

// getPushedComponents retrieves a map of PushedComponents from the cluster, keyed by their name
func getPushedComponents(applicationName string, kubeclient kclient.ClientInterface) (map[string]PushedComponent, error) {
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)

	deploymentList, err := kubeclient.ListDeployments(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}
	res := make(map[string]PushedComponent, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		deployment := d
		storageClient := storage.NewClient(storage.ClientOptions{
			Client:     kubeclient,
			Deployment: &deployment,
		})

		urlClient := url.NewClient(url.ClientOptions{
			Client:     kubeclient,
			Deployment: &deployment,
		})
		comp := newPushedComponent(applicationName, &devfileComponent{d: d}, kubeclient, storageClient, urlClient)
		res[comp.GetName()] = comp
	}

	return res, nil
}

// getPushedComponent returns an abstraction over the cluster representation of the component
func getPushedComponent(componentName, applicationName string, kubeclient kclient.ClientInterface) (PushedComponent, error) {
	d, err := kubeclient.GetOneDeployment(componentName, applicationName)
	if err != nil {
		if isIgnorableError(err) {
			return nil, nil
		}
		return nil, err
	}
	storageClient := storage.NewClient(storage.ClientOptions{
		Client:     kubeclient,
		Deployment: d,
	})

	urlClient := url.NewClient(url.ClientOptions{
		Client:     kubeclient,
		Deployment: d,
	})
	return newPushedComponent(applicationName, &devfileComponent{d: *d}, kubeclient, storageClient, urlClient), nil
}

// setLinksServiceNames sets the service name of the links from the info in ServiceBindingRequests present in the cluster
func setLinksServiceNames(kubeclient kclient.ClientInterface, linkedSecrets []SecretMount, selector string) error {
	ok, err := kubeclient.IsServiceBindingSupported()
	if err != nil {
		return fmt.Errorf("unable to check if service binding is supported: %w", err)
	}

	serviceBindings := map[string]string{}
	if ok {
		// service binding operator is installed on the cluster
		list, err := kubeclient.ListDynamicResource(kclient.ServiceBindingGroup, kclient.ServiceBindingVersion, kclient.ServiceBindingResource)
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
			serviceObj := services[0]
			if serviceObj.Kind == "Service" {
				serviceBindings[sbr.Status.Secret] = serviceObj.Name
			} else {
				serviceBindings[sbr.Status.Secret] = serviceObj.Kind + "/" + serviceObj.Name
			}
		}
	} else {
		// service binding operator is not installed
		// get the secrets instead of the service binding objects to retrieve the link data
		secrets, err := kubeclient.ListSecrets(selector)
		if err != nil {
			return err
		}

		// get the services to get their names against the component names
		services, err := kubeclient.ListServices("")
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

// getComponentFromDevfile extracts component's metadata from the specified env info if it exists
func getComponentFromDevfile(info *envinfo.EnvSpecificInfo) (Component, parser.DevfileObj, error) {
	if info.Exists() {
		devfileObj, err := parser.Parse(info.GetDevfilePath())
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		component, err := getComponentFrom(info, GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))
		if err != nil {
			return Component{}, parser.DevfileObj{}, err
		}
		components, err := devfileObj.Data.GetComponents(parsercommon.DevfileOptions{})
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

		return component, devfileObj, nil
	}
	return Component{}, parser.DevfileObj{}, nil
}

// CheckDefaultProject errors out if the project resource is supported and the value is "default"
func (c componentClient) CheckDefaultProject(name string) error {
	// Check whether resource "Project" is supported
	projectSupported, err := c.client.IsProjectSupported()

	if err != nil {
		return errors.Wrap(err, "resource project validation check failed.")
	}

	if projectSupported && name == "default" {
		return &DefaultProjectError{}
	}
	return nil
}
