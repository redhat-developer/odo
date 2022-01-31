package component

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/service"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentFullDescriptionSpec represents the complete description of the component
type ComponentFullDescriptionSpec struct {
	App     string              `json:"app,omitempty"`
	Type    string              `json:"type,omitempty"`
	Source  string              `json:"source,omitempty"`
	URL     urlpkg.URLList      `json:"urls,omitempty"`
	Storage storage.StorageList `json:"storages,omitempty"`
	Env     []corev1.EnvVar     `json:"env,omitempty"`
	Ports   []string            `json:"ports,omitempty"`
}

// ComponentFullDescription describes a component fully
type ComponentFullDescription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentFullDescriptionSpec `json:"spec,omitempty"`
	Status            ComponentStatus              `json:"status,omitempty"`
}

// copyFromComponentDescription copies over all fields from Component that can be copied
func (cfd *ComponentFullDescription) copyFromComponentDesc(component *Component) error {
	d, err := json.Marshal(component)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, cfd)
}

// loadStoragesFromClientAndLocalConfig collects information about storages both locally and from the cluster.
func (cfd *ComponentFullDescription) loadStoragesFromClientAndLocalConfig(client kclient.ClientInterface, configProvider localConfigProvider.LocalConfigProvider, componentName string, applicationName string, componentDesc *Component) error {
	var storages storage.StorageList
	var err error

	// if component is pushed call ListWithState which gets storages from localconfig and cluster
	// this result is already in mc readable form
	if componentDesc.Status.State == StateTypePushed {
		storageClient := storage.NewClient(storage.ClientOptions{
			Client:              client,
			LocalConfigProvider: configProvider,
		})

		storages, err = storageClient.List()
		if err != nil {
			return err
		}
	} else {
		// otherwise simply fetch storagelist locally
		storageLocal, err := configProvider.ListStorage()
		if err != nil {
			return err
		}
		// convert to machine readable format
		storages = storage.ConvertListLocalToMachine(storageLocal)
	}
	cfd.Spec.Storage = storages
	return nil
}

// fillEmptyFields fills any fields that are empty in the ComponentFullDescription
func (cfd *ComponentFullDescription) fillEmptyFields(componentDesc Component, componentName string, applicationName string, projectName string) {
	// fix missing names in case it is in not in description
	if len(cfd.Name) <= 0 {
		cfd.Name = componentName
	}

	if len(cfd.Namespace) <= 0 {
		cfd.Namespace = projectName
	}

	if len(cfd.Kind) <= 0 {
		cfd.Kind = "Component"
	}

	if len(cfd.APIVersion) <= 0 {
		cfd.APIVersion = apiVersion
	}

	if len(cfd.Spec.App) <= 0 {
		cfd.Spec.App = applicationName
	}
	cfd.Spec.Ports = componentDesc.Spec.Ports
}

// NewComponentFullDescriptionFromClientAndLocalConfigProvider gets the complete description of the component from cluster
func NewComponentFullDescriptionFromClientAndLocalConfigProvider(client kclient.ClientInterface, envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error) {
	cfd := &ComponentFullDescription{}
	var state State
	if client == nil {
		state = StateTypeUnknown
	} else {
		state = GetComponentState(client, componentName, applicationName)
	}
	var componentDesc Component
	var devfile devfileParser.DevfileObj
	var err error
	var configLinks []string
	componentDesc, devfile, err = GetComponentFromDevfile(envInfo)
	if err != nil {
		return cfd, err
	}
	configLinks, err = service.ListDevfileLinks(devfile, context)

	if err != nil {
		return cfd, err
	}
	err = cfd.copyFromComponentDesc(&componentDesc)
	if err != nil {
		return cfd, err
	}
	cfd.Status.State = state
	var deployment *v1.Deployment
	if state == StateTypePushed {
		componentDescFromCluster, e := getRemoteComponentMetadata(client, componentName, applicationName, false, false)
		if e != nil {
			return cfd, e
		}
		cfd.Spec.Env = componentDescFromCluster.Spec.Env
		cfd.Spec.Type = componentDescFromCluster.Spec.Type
		cfd.Status.LinkedServices = componentDescFromCluster.Status.LinkedServices
		// Obtain the deployment to correctly instantiate the Storage and URL client and get information
		deployment, err = client.GetOneDeployment(componentName, applicationName)
		if err != nil {
			// ideally the error should not occur since we have already established that the component exists on the cluster
			return cfd, err
		}
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
	if client == nil {
		routeSupported = false
	} else {
		routeSupported, e = client.IsRouteSupported()
		if e != nil {
			// we assume if there was an error then the cluster is not connected
			routeSupported = false
		}
	}

	var configProvider localConfigProvider.LocalConfigProvider
	envInfo.SetDevfileObj(devfile)
	configProvider = envInfo

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		LocalConfigProvider: configProvider,
		Client:              client,
		IsRouteSupported:    routeSupported,
		Deployment:          deployment,
	})

	urls, err = urlClient.List()
	if err != nil {
		log.Warningf("URLs couldn't not be retrieved: %v", err)
	}
	cfd.Spec.URL = urls

	// Obtain Storage information
	var storages storage.StorageList
	storageClient := storage.NewClient(storage.ClientOptions{
		Client:              client,
		LocalConfigProvider: envInfo,
		Deployment:          deployment,
	})

	// We explicitly do not fill in Spec.Storage for describe because,
	// it is essentially a list of Storage names, which seems redundant when the detailed StorageSpec is available
	storages, err = storageClient.List()
	if err != nil {
		log.Warningf("Storages couldn't not be retrieved: %v", err)
	}
	cfd.Spec.Storage = storages

	return cfd, nil
}

// Print prints the complete information of component onto stdout (Note: long term this function should not need to access any parameters, but just print the information in struct)
func (cfd *ComponentFullDescription) Print(client kclient.ClientInterface) error {
	// TODO: remove the need to client here print should just deal with printing
	log.Describef("Component Name: ", cfd.GetName())
	log.Describef("Type: ", cfd.Spec.Type)

	// Source
	if cfd.Spec.Source != "" {
		log.Describef("Source: ", cfd.Spec.Source)
	}

	// Env
	if cfd.Spec.Env != nil {

		// Retrieve all the environment variables
		var output string
		for _, env := range cfd.Spec.Env {
			output += fmt.Sprintf(" · %v=%v\n", env.Name, env.Value)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Environment Variables:\n", output)
		}

	}

	// Storage
	if len(cfd.Spec.Storage.Items) > 0 {

		// Gather the output
		var output string
		for _, store := range cfd.Spec.Storage.Items {
			eph := ""
			if store.Spec.Ephemeral != nil {
				if *store.Spec.Ephemeral {
					eph = " as ephemeral volume"
				} else {
					eph = " as persistent volume"
				}
			}
			output += fmt.Sprintf(" · %v of size %v mounted to %v%s\n", store.Name, store.Spec.Size, store.Spec.Path, eph)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Storage:\n", output)
		}

	}

	// URL
	if len(cfd.Spec.URL.Items) > 0 {
		var output string
		// if the component is not pushed
		for _, componentURL := range cfd.Spec.URL.Items {
			if componentURL.Status.State == urlpkg.StateTypePushed {
				output += fmt.Sprintf(" · %v exposed via %v\n", urlpkg.GetURLString(componentURL.Spec.Protocol, componentURL.Spec.Host, ""), componentURL.Spec.Port)
			} else {
				output += fmt.Sprintf(" · URL named %s will be exposed via %v\n", componentURL.Name, componentURL.Spec.Port)
			}
		}

		// Cut off the last newline and output
		output = output[:len(output)-1]
		log.Describef("URLs:\n", output)
	}

	// Linked services
	if len(cfd.Status.LinkedServices) > 0 {

		// Gather the output
		var output string
		for _, linkedService := range cfd.Status.LinkedServices {

			if linkedService.SecretName == "" {
				output += fmt.Sprintf(" · %s\n", linkedService.ServiceName)
				continue
			}

			// Let's also get the secrets / environment variables that are being passed in.. (if there are any)
			secrets, err := client.GetSecret(linkedService.SecretName, cfd.GetNamespace())
			if err != nil {
				return err
			}

			if len(secrets.Data) > 0 {
				// Iterate through the secrets to throw in a string
				var secretOutput string
				for i := range secrets.Data {
					if linkedService.MountVolume {
						secretOutput += fmt.Sprintf("    · %v\n", filepath.ToSlash(filepath.Join(linkedService.MountPath, i)))
					} else {
						secretOutput += fmt.Sprintf("    · %v\n", i)
					}
				}

				if len(secretOutput) > 0 {
					// Cut off the last newline
					secretOutput = secretOutput[:len(secretOutput)-1]
					if linkedService.MountVolume {
						output += fmt.Sprintf(" · %s\n   Files:\n%s\n", linkedService.ServiceName, secretOutput)
					} else {
						output += fmt.Sprintf(" · %s\n   Environment Variables:\n%s\n", linkedService.ServiceName, secretOutput)
					}
				}

			} else {
				output += fmt.Sprintf(" · %s\n", linkedService.SecretName)
			}

		}

		if len(output) > 0 {
			// Cut off the last newline and output
			output = output[:len(output)-1]
			log.Describef("Linked Services:\n", output)
		}

	}
	return nil
}

// GetComponent returns a component representation
func (cfd *ComponentFullDescription) GetComponent() Component {
	cmp := NewComponent(cfd.Name)
	cmp.Spec.App = cfd.Spec.App
	cmp.Spec.Ports = cfd.Spec.Ports
	cmp.Spec.Type = cfd.Spec.Type
	cmp.Spec.StorageSpec = cfd.Spec.Storage.Items
	cmp.Spec.URLSpec = cfd.Spec.URL.Items
	for _, url := range cfd.Spec.URL.Items {
		cmp.Spec.URL = append(cmp.Spec.URL, url.Name)
	}
	for _, storage := range cfd.Spec.Storage.Items {
		cmp.Spec.Storage = append(cmp.Spec.URL, storage.Name)
	}
	cmp.ObjectMeta.Namespace = cfd.ObjectMeta.Namespace
	cmp.Status = cfd.Status
	cmp.Spec.Env = cfd.Spec.Env
	return cmp
}
