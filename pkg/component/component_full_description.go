package component

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/service"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/storage"
	urlpkg "github.com/openshift/odo/pkg/url"
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
func (cfd *ComponentFullDescription) loadStoragesFromClientAndLocalConfig(client *occlient.Client, configProvider localConfigProvider.LocalConfigProvider, componentName string, applicationName string, componentDesc *Component) error {
	var storages storage.StorageList
	var err error
	var derefClient occlient.Client

	if client != nil {
		derefClient = *client
	}

	// if component is pushed call ListWithState which gets storages from localconfig and cluster
	// this result is already in mc readable form
	if componentDesc.Status.State == StateTypePushed {
		storageClient := storage.NewClient(storage.ClientOptions{
			OCClient:            derefClient,
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
	//fix missing names in case it in not in description
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
func NewComponentFullDescriptionFromClientAndLocalConfigProvider(client *occlient.Client, envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*ComponentFullDescription, error) {
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
	if state == StateTypePushed {
		componentDescFromCluster, err := getRemoteComponentMetadata(client, componentName, applicationName, false, false)
		if err != nil {
			return cfd, err
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

	var derefClient occlient.Client

	if client != nil {
		derefClient = *client
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		LocalConfigProvider: configProvider,
		OCClient:            derefClient,
		IsRouteSupported:    routeSupported,
	})

	urls, err = urlClient.List()
	if err != nil {
		log.Warningf("URLs couldn't not be retrieved: %v", err)
	}
	cfd.Spec.URL = urls

	err = cfd.loadStoragesFromClientAndLocalConfig(client, configProvider, componentName, applicationName, &componentDesc)
	if err != nil {
		return cfd, err
	}

	return cfd, nil
}

// Print prints the complete information of component onto stdout (Note: long term this function should not need to access any parameters, but just print the information in struct)
func (cfd *ComponentFullDescription) Print(client *occlient.Client) error {
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
			output += fmt.Sprintf(" · %v of size %v mounted to %v\n", store.Name, store.Spec.Size, store.Spec.Path)
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
			secrets, err := client.GetKubeClient().GetSecret(linkedService.SecretName, cfd.GetNamespace())
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
