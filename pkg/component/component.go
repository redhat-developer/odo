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

// NewComponentFullDescriptionFromClientAndLocalConfigProvider gets the complete description of the component from cluster
func (c componentClient) NewComponentFullDescriptionFromClientAndLocalConfigProvider(envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error) {
	cfd := &ComponentFullDescription{}
	var state State
	if c.client == nil {
		state = StateTypeUnknown
	} else {
		state = c.GetComponentState(componentName, applicationName)
	}
	var componentDesc Component
	var devfileObj parser.DevfileObj
	var err error
	var configLinks []string
	componentDesc, devfileObj, err = c.GetComponentFromDevfile(envInfo)
	if err != nil {
		return cfd, err
	}
	configLinks, err = service.ListDevfileLinks(devfileObj, context)

	if err != nil {
		return cfd, err
	}
	err = cfd.copyFromComponentDesc(&componentDesc)
	if err != nil {
		return cfd, err
	}
	cfd.Status.State = state
	if state == StateTypePushed {
		componentDescFromCluster, e := c.getRemoteComponentMetadata(componentName, applicationName, false, false)
		if e != nil {
			return cfd, e
		}
		cfd.Spec.Env = componentDescFromCluster.Spec.Env
		cfd.Spec.Type = componentDescFromCluster.Spec.Type
		cfd.Status.LinkedServices = componentDescFromCluster.Status.LinkedServices
	}

	for _, link := range configLinks {
		found := false
		for _, linked := range cfd.Status.LinkedServices {
			if linked.ServiceName == link {
				found = true
				break
			}
		}
		if !found {
			cfd.Status.LinkedServices = append(cfd.Status.LinkedServices, SecretMount{
				ServiceName: link,
			})
		}
	}

	cfd.fillEmptyFields(componentDesc, componentName, applicationName, projectName)

	var urls urlpkg.URLList

	var routeSupported bool
	var e error
	if c.client == nil {
		routeSupported = false
	} else {
		routeSupported, e = c.client.IsRouteSupported()
		if e != nil {
			// we assume if there was an error then the cluster is not connected
			routeSupported = false
		}
	}

	var configProvider localConfigProvider.LocalConfigProvider
	envInfo.SetDevfileObj(devfileObj)
	configProvider = envInfo

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		LocalConfigProvider: configProvider,
		Client:              c.client,
		IsRouteSupported:    routeSupported,
	})

	urls, err = urlClient.List()
	if err != nil {
		log.Warningf("URLs couldn't not be retrieved: %v", err)
	}
	cfd.Spec.URL = urls

	var storages storage.StorageList
	storages, err = c.loadStoragesFromClientAndLocalConfig(configProvider, &componentDesc)
	if err != nil {
		return cfd, err
	}
	cfd.Spec.Storage = storages

	return cfd, nil
}

// GetComponentNames retrieves the names of the components in the specified application
func (c componentClient) GetComponentNames(applicationName string) ([]string, error) {
	components, err := c.GetPushedComponents(applicationName)
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
		component, err := c.GetComponent(elem.Labels[componentlabels.ComponentLabel], elem.Labels[applabels.ApplicationLabel])
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

// GetComponentFromDevfile extracts component's metadata from the specified env info if it exists
func (c componentClient) GetComponentFromDevfile(info *envinfo.EnvSpecificInfo) (Component, parser.DevfileObj, error) {
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

func (c componentClient) ListDevfileComponentsInPath(paths []string) ([]Component, error) {
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

func (c componentClient) GetComponentState(componentName, applicationName string) State {
	// first check if a deployment exists
	pushedComponent, err := c.GetPushedComponent(componentName, applicationName)
	if err != nil {
		return StateTypeUnknown
	}
	if pushedComponent != nil {
		return StateTypePushed
	}
	return StateTypeNotPushed
}

// GetComponent provides component definition
func (c componentClient) GetComponent(componentName string, applicationName string) (component Component, err error) {
	return c.getRemoteComponentMetadata(componentName, applicationName, true, true)
}

// GetPushedComponents retrieves a map of PushedComponents from the cluster, keyed by their name
func (c componentClient) GetPushedComponents(applicationName string) (map[string]PushedComponent, error) {
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)

	deploymentList, err := c.client.ListDeployments(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list components")
	}
	res := make(map[string]PushedComponent, len(deploymentList.Items))
	for _, d := range deploymentList.Items {
		deployment := d
		storageClient := storage.NewClient(storage.ClientOptions{
			Client:     c.client,
			Deployment: &deployment,
		})

		urlClient := url.NewClient(url.ClientOptions{
			Client:     c.client,
			Deployment: &deployment,
		})
		comp := c.newPushedComponent(applicationName, &devfileComponent{d: d}, storageClient, urlClient)
		res[comp.GetName()] = comp
	}

	return res, nil
}

// GetPushedComponent returns an abstraction over the cluster representation of the component
func (c componentClient) GetPushedComponent(componentName, applicationName string) (PushedComponent, error) {
	d, err := c.client.GetOneDeployment(componentName, applicationName)
	if err != nil {
		if isIgnorableError(err) {
			return nil, nil
		}
		return nil, err
	}
	storageClient := storage.NewClient(storage.ClientOptions{
		Client:     c.client,
		Deployment: d,
	})

	urlClient := url.NewClient(url.ClientOptions{
		Client:     c.client,
		Deployment: d,
	})
	return c.newPushedComponent(applicationName, &devfileComponent{d: *d}, storageClient, urlClient), nil
}

// CheckDefaultProject errors out if the project resource is supported and the value is "default"
func (c componentClient) CheckDefaultProject(name string) error {
	// Check whether resource "Project" is supported
	projectSupported, err := c.client.IsProjectSupported()

	if err != nil {
		return errors.Wrap(err, "resource project validation check failed.")
	}

	if projectSupported && name == "default" {
		return errors.New("odo may not work as expected in the default project, please run the odo component in a non-default project")
	}
	return nil
}

// getRemoteComponentMetadata provides component metadata from the cluster
func (c componentClient) getRemoteComponentMetadata(componentName string, applicationName string, getUrls, getStorage bool) (Component, error) {
	fromCluster, err := c.GetPushedComponent(componentName, applicationName)
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
	err = c.setLinksServiceNames(linkedSecrets, componentlabels.GetSelector(componentName, applicationName))
	if err != nil {
		return Component{}, fmt.Errorf("unable to get name of services: %w", err)
	}
	component.Status.LinkedServices = linkedSecrets

	// Annotations
	component.Annotations = fromCluster.GetAnnotations()

	// Labels
	component.Labels = fromCluster.GetLabels()

	component.Namespace = c.client.GetCurrentNamespace()
	component.Spec.App = applicationName
	component.Spec.Env = filteredEnv
	component.Status.State = StateTypePushed

	return component, nil
}

// setLinksServiceNames sets the service name of the links from the info in ServiceBindingRequests present in the cluster
func (c componentClient) setLinksServiceNames(linkedSecrets []SecretMount, selector string) error {
	ok, err := c.client.IsServiceBindingSupported()
	if err != nil {
		return fmt.Errorf("unable to check if service binding is supported: %w", err)
	}

	serviceBindings := map[string]string{}
	if ok {
		// service binding operator is installed on the cluster
		list, err := c.client.ListDynamicResource(kclient.ServiceBindingGroup, kclient.ServiceBindingVersion, kclient.ServiceBindingResource)
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
		secrets, err := c.client.ListSecrets(selector)
		if err != nil {
			return err
		}

		// get the services to get their names against the component names
		services, err := c.client.ListServices("")
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

// loadStoragesFromClientAndLocalConfig collects information about storages both locally and from the cluster.
func (c componentClient) loadStoragesFromClientAndLocalConfig(configProvider localConfigProvider.LocalConfigProvider, componentDesc *Component) (storage.StorageList, error) {
	var storages storage.StorageList
	var err error

	// if component is pushed call ListWithState which gets storages from localconfig and cluster
	// this result is already in mc readable form
	if componentDesc.Status.State == StateTypePushed {
		storageClient := storage.NewClient(storage.ClientOptions{
			Client:              c.client,
			LocalConfigProvider: configProvider,
		})

		storages, err = storageClient.List()
		if err != nil {
			return storages, err
		}
	} else {
		var storageLocal []localConfigProvider.LocalStorage
		// otherwise simply fetch storagelist locally
		storageLocal, err = configProvider.ListStorage()
		if err != nil {
			return storages, err
		}
		// convert to machine readable format
		storages = storage.ConvertListLocalToMachine(storageLocal)
	}
	return storages, nil
}

func (c componentClient) newPushedComponent(applicationName string, p provider, storageClient storage.Client, urlClient url.Client) PushedComponent {
	return &defaultPushedComponent{
		application:   applicationName,
		provider:      p,
		client:        c.client,
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
