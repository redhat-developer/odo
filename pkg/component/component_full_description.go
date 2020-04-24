package component

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/storage"
	urlpkg "github.com/openshift/odo/pkg/url"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//ComponentFullDescriptionSpec repersents complete desciption of the component
type ComponentFullDescriptionSpec struct {
	App        string              `json:"app,omitempty"`
	Type       string              `json:"type,omitempty"`
	Source     string              `json:"source,omitempty"`
	SourceType string              `json:"sourceType,omitempty"`
	URL        urlpkg.URLList      `json:"urls,omitempty"`
	Storage    storage.StorageList `json:"storages,omitempty"`
	Env        []corev1.EnvVar     `json:"env,omitempty"`
	Ports      []string            `json:"ports,omitempty"`
}

type ComponentFullDescription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentFullDescriptionSpec `json:"spec,omitempty"`
	Status            ComponentStatus              `json:"status,omitempty"`
}

func (cfd *ComponentFullDescription) copyFromComponentDesc(component *Component) error {
	d, err := json.Marshal(component)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, cfd)
}

func (cfd *ComponentFullDescription) loadURLS(client *occlient.Client, localConfigInfo *config.LocalConfigInfo, componentName string, applicationName string) error {
	urls, err := urlpkg.List(client, localConfigInfo, componentName, applicationName)
	if err != nil {
		return err
	}

	cfd.Spec.URL = urls
	return nil
}

func (cfd *ComponentFullDescription) loadStorages(client *occlient.Client, localConfigInfo *config.LocalConfigInfo, componentName string, applicationName string, componentDesc *Component) error {
	var storages storage.StorageList
	var err error
	if componentDesc.Status.State == StateTypePushed {
		storages, err = storage.ListStorageWithState(client, localConfigInfo, componentName, applicationName)
		if err != nil {
			return err
		}
	} else {
		storageLocal, err := localConfigInfo.StorageList()
		if err != nil {
			return err
		}
		storages = storage.ConvertListLocalToMachine(storageLocal)
	}
	cfd.Spec.Storage = storages
	return nil
}

//NewComponentFullDescription gets the complete description of the component from both localconfig and cluster
func NewComponentFullDescription(client *occlient.Client, localConfigInfo *config.LocalConfigInfo, componentName string, applicationName string, projectName string) (*ComponentFullDescription, error) {
	cfd := &ComponentFullDescription{}
	state := GetComponentState(client, componentName, applicationName)
	componentDesc, err := GetComponentFromConfig(localConfigInfo)
	if err != nil {
		return cfd, err
	}
	err = cfd.copyFromComponentDesc(&componentDesc)
	if err != nil {
		return cfd, err
	}

	//fix missing names in case it in not in description
	if len(cfd.Name) <= 0 {
		cfd.Name = componentName
	}

	if state == StateTypePushed {
		componentDescFromCluster, err := GetComponent(client, componentName, applicationName, projectName)
		if err != nil {
			return cfd, err
		}
		cfd.Spec.Env = componentDescFromCluster.Spec.Env
	}

	cfd.Status.State = state

	err = cfd.loadURLS(client, localConfigInfo, componentName, applicationName)
	if err != nil {
		return cfd, err
	}
	err = cfd.loadStorages(client, localConfigInfo, componentName, applicationName, &componentDesc)
	if err != nil {
		return cfd, err
	}
	return cfd, nil
}

func (componentDesc *ComponentFullDescription) PrintInfo(client *occlient.Client, localConfigInfo *config.LocalConfigInfo) error {
	log.Describef("Component Name: ", componentDesc.GetName())
	log.Describef("Type: ", componentDesc.Spec.Type)

	// Source
	if componentDesc.Spec.Source != "" {
		log.Describef("Source: ", componentDesc.Spec.Source)
	}

	// Env
	if componentDesc.Spec.Env != nil {

		// Retrieve all the environment variables
		var output string
		for _, env := range componentDesc.Spec.Env {
			output += fmt.Sprintf(" · %v=%v\n", env.Name, env.Value)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Environment Variables:\n", output)
		}

	}

	// Storage
	if len(componentDesc.Spec.Storage.Items) > 0 {

		// Gather the output
		var output string
		for _, store := range componentDesc.Spec.Storage.Items {
			output += fmt.Sprintf(" · %v of size %v mounted to %v\n", store.Name, store.Spec.Size, store.Spec.Path)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Storage:\n", output)
		}

	}

	// URL
	if len(componentDesc.Spec.URL.Items) > 0 {
		var output string

		if !experimental.IsExperimentalModeEnabled() {
			// if the component is not pushed
			for i, componentURL := range componentDesc.Spec.URL.Items {
				if componentURL.Status.State == urlpkg.StateTypePushed {
					output += fmt.Sprintf(" · %v exposed via %v\n", urlpkg.GetURLString(componentURL.Spec.Protocol, componentURL.Spec.Host, ""), componentURL.Spec.Port)
				} else {
					var p string
					if i >= len(componentDesc.Spec.Ports) {
						p = componentDesc.Spec.Ports[len(componentDesc.Spec.Ports)-1]
					} else {
						p = componentDesc.Spec.Ports[i]
					}
					output += fmt.Sprintf(" · URL named %s will be exposed via %v\n", componentURL.Name, p)
				}
			}
		}
		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("URLs:\n", output)
		}

	}

	// Linked components
	if len(componentDesc.Status.LinkedComponents) > 0 {

		// Gather the output
		var output string
		for name, ports := range componentDesc.Status.LinkedComponents {
			if len(ports) > 0 {
				output += fmt.Sprintf(" · %v - Port(s): %v\n", name, strings.Join(ports, ","))
			} else {
				output += fmt.Sprintf(" · %v\n", name)
			}
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Linked Components:\n", output)
		}

	}

	// Linked services
	if len(componentDesc.Status.LinkedServices) > 0 {

		// Gather the output
		var output string
		for _, linkedService := range componentDesc.Status.LinkedServices {

			// Let's also get the secrets / environment variables that are being passed in.. (if there are any)
			secrets, err := client.GetSecret(linkedService, componentDesc.GetNamespace())
			if err != nil {
				return err
			}

			if len(secrets.Data) > 0 {
				// Iterate through the secrets to throw in a string
				var secretOutput string
				for i := range secrets.Data {
					secretOutput += fmt.Sprintf("    · %v\n", i)
				}

				if len(secretOutput) > 0 {
					// Cut off the last newline
					secretOutput = secretOutput[:len(secretOutput)-1]
					output += fmt.Sprintf(" · %s\n   Environment Variables:\n%s\n", linkedService, secretOutput)
				}

			} else {
				output += fmt.Sprintf(" · %s\n", linkedService)
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
