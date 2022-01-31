package component

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/parser"
	v1 "k8s.io/api/apps/v1"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/service"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/storage"
	urlpkg "github.com/redhat-developer/odo/pkg/url"
)

// fillEmptyFields fills any fields that are empty in the ComponentFullDescription
func (cmp *Component) fillEmptyFields(componentName string, applicationName string, projectName string) {
	// fix missing names in case it in not in description
	if len(cmp.Name) <= 0 {
		cmp.Name = componentName
	}

	if len(cmp.Namespace) <= 0 {
		cmp.Namespace = projectName
	}

	if len(cmp.Kind) <= 0 {
		cmp.Kind = "Component"
	}

	if len(cmp.APIVersion) <= 0 {
		cmp.APIVersion = apiVersion
	}

	if len(cmp.Spec.App) <= 0 {
		cmp.Spec.App = applicationName
	}
}

// NewComponentFullDescriptionFromClientAndLocalConfigProvider gets the complete description of the component from cluster
func NewComponentFullDescriptionFromClientAndLocalConfigProvider(client kclient.ClientInterface, envInfo *envinfo.EnvSpecificInfo, componentName string, applicationName string, projectName string, context string) (*Component, error) {
	// Read component from the devfile
	componentDesc, devfileObj, err := GetComponentFromDevfile(envInfo)
	if err != nil {
		return &componentDesc, err
	}

	var deployment *v1.Deployment

	// user has access to the cluster
	hasAccessToCluster := client != nil
	// if local component is the same as user asked component
	userMatchesDevfile := componentName == componentDesc.Name

	if hasAccessToCluster {
		var componentDescFromCluster Component
		componentDescFromCluster, err = getRemoteComponentMetadata(client, componentName, applicationName, false, false)
		// TODO: Move these to functions?
		componentFoundInCluster := err == nil

		// component was not found on the cluster
		if !componentFoundInCluster {
			// if local component is not the same as user asked component,
			// then return a simple component with user provided information with state as to Unknown
			if !userMatchesDevfile {
				// TODO: Move this to function
				componentDesc = Component{}
				componentDesc.Status.State = StateTypeUnknown
				devfileObj = parser.DevfileObj{}
			} else {
				// since local component and user asked component are same, we can assume that the component has not been pushed yet,
				// hence set it's state to Not Pushed
				componentDesc.Status.State = StateTypeNotPushed
			}
		} else {
			// component was found on the cluster

			// if the user asked component was found on the cluster, but is not the same as local component,
			// then set local component obtained from the devfile to reflect the remote component;
			// also nullify the devfileObj since it becomes irrelevant to the information user wants
			if !userMatchesDevfile {
				componentDesc = componentDescFromCluster
				devfileObj = parser.DevfileObj{}
			} else {
				// if both remote and local component are same, then assign the necessary remote data into the local data
				componentDesc.Status.State = componentDescFromCluster.Status.State
				componentDesc.Annotations = componentDescFromCluster.Annotations
				componentDesc.Labels = componentDescFromCluster.Labels
				componentDesc.CreationTimestamp = componentDescFromCluster.CreationTimestamp
				componentDesc.Spec.Env = append(componentDesc.Spec.Env, componentDescFromCluster.Spec.Env...)
				componentDesc.Spec.Type = componentDescFromCluster.Spec.Type
				componentDesc.Status.LinkedServices = componentDescFromCluster.Status.LinkedServices
			}
			// Obtain the deployment to correctly instantiate the Storage and URL client and get information
			deployment, err = client.GetOneDeployment(componentName, applicationName)
			if err != nil {
				// ideally the error should not occur since we have already established that the component exists on the cluster
				return &componentDesc, err
			}
		}
	} else {
		// user does not have access to the cluster

		// if local component is not the same as user asked component,
		// then return a simple component with user provided information with state as to Unknown
		if !userMatchesDevfile {
			componentDesc = Component{}
			componentDesc.Status.State = StateTypeUnknown
		} else {
			// if the user asked component is same as the local component, then set it's status to Not Pushed
			componentDesc.Status.State = StateTypeNotPushed
		}
	}

	// Fill empty fields (especially APIVersion, and Kind)
	componentDesc.fillEmptyFields(componentName, applicationName, projectName)

	if componentDesc.Status.State == StateTypeUnknown {
		return &componentDesc, nil
	}

	envInfo.SetDevfileObj(devfileObj)

	// Obtain the information (about Storage, URL, and Links) because that is not parsed from the devfile so far
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
	if client != nil {
		// we assume if there was an error then the cluster is not connected
		routeSupported, _ = client.IsRouteSupported()
	}

	urlClient := urlpkg.NewClient(urlpkg.ClientOptions{
		LocalConfigProvider: envInfo,
		Client:              client,
		IsRouteSupported:    routeSupported,
		Deployment:          deployment,
	})

	urls, err = urlClient.List()
	if err != nil {
		log.Warningf("URLs couldn't not be retrieved: %v", err)
	}
	// We explicitly do not fill in Spec.URL for describe because,
	// it is essentially a list of URL names, which seems redundant when the detailed URLSpec is available
	componentDesc.Spec.URLSpec = urls.Items

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
	componentDesc.Spec.StorageSpec = storages.Items

	return &componentDesc, nil

}

// Print prints the complete information of component onto stdout (Note: long term this function should not need to access any parameters, but just print the information in struct)
func (cmp *Component) Print(client kclient.ClientInterface) error {
	// TODO: remove the need to client here print should just deal with printing
	log.Describef("Component Name: ", cmp.GetName())
	log.Describef("Type: ", cmp.Spec.Type)

	// Env
	if cmp.Spec.Env != nil {

		// Retrieve all the environment variables
		var output string
		for _, env := range cmp.Spec.Env {
			output += fmt.Sprintf(" · %v=%v\n", env.Name, env.Value)
		}

		// Cut off the last newline and output
		if len(output) > 0 {
			output = output[:len(output)-1]
			log.Describef("Environment Variables:\n", output)
		}

	}

	// Storage
	if len(cmp.Spec.StorageSpec) > 0 {

		// Gather the output
		var output string
		for _, store := range cmp.Spec.StorageSpec {
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
	if len(cmp.Spec.URLSpec) > 0 {
		var output string
		// if the component is not pushed
		for _, componentURL := range cmp.Spec.URLSpec {
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
	if len(cmp.Status.LinkedServices) > 0 {

		// Gather the output
		var output string
		for _, linkedService := range cmp.Status.LinkedServices {

			if linkedService.SecretName == "" {
				output += fmt.Sprintf(" · %s\n", linkedService.ServiceName)
				continue
			}

			// Let's also get the secrets / environment variables that are being passed in.. (if there are any)
			secrets, err := client.GetSecret(linkedService.SecretName, cmp.GetNamespace())
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
