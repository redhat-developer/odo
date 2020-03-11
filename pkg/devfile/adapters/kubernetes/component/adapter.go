package component

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/glog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/openshift/odo/pkg/kclient"
)

// New instantiantes a component adapter
func New(adapterContext common.AdapterContext, client kclient.Client) common.ComponentAdapter {
	return Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client kclient.Client
	common.AdapterContext
}

// Create generates the Kubernetes resources, adds the devfile storage and
// updates the component if a matching component exists or creates one if it doesn't exist
func (a Adapter) Create() (err error) {
	componentName := a.ComponentName

	labels := map[string]string{
		"component": componentName,
	}

	containers := utils.GetContainers(a.Devfile)
	if len(containers) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	objectMeta := kclient.CreateObjectMeta(componentName, a.Client.Namespace, labels, nil)
	podTemplateSpec := kclient.GeneratePodTemplateSpec(objectMeta, containers)

	componentAliasToVolumes := utils.GetVolumes(a.Devfile)

	var uniqueStorages []common.Storage
	volumeNameToPVCName := make(map[string]string)
	processedVolumes := make(map[string]bool)

	// Get a list of all the unique volume names and generate their PVC names
	for _, volumes := range componentAliasToVolumes {
		for _, vol := range volumes {
			if _, ok := processedVolumes[*vol.Name]; !ok {
				processedVolumes[*vol.Name] = true

				// Generate the PVC Names
				glog.V(3).Infof("Generating PVC name for %v", *vol.Name)
				generatedPVCName, err := storage.GeneratePVCNameFromDevfileVol(*vol.Name, componentName)
				if err != nil {
					return err
				}

				// Check if we have an existing PVC with the labels, overwrite the generated name with the existing name if present
				existingPVCName, err := storage.GetExistingPVC(&a.Client, *vol.Name, componentName)
				if err != nil {
					return err
				}
				if len(existingPVCName) > 0 {
					glog.V(3).Infof("Found an existing PVC for %v, PVC %v will be re-used", *vol.Name, existingPVCName)
					generatedPVCName = existingPVCName
				}

				pvc := common.Storage{
					Name:   generatedPVCName,
					Volume: vol,
				}
				uniqueStorages = append(uniqueStorages, pvc)
				volumeNameToPVCName[*vol.Name] = generatedPVCName
			}
		}
	}

	// Add PVC and Volume Mounts to the podTemplateSpec
	err = kclient.AddPVCAndVolumeMount(podTemplateSpec, volumeNameToPVCName, componentAliasToVolumes)
	if err != nil {
		return err
	}

	deploymentSpec := kclient.GenerateDeploymentSpec(*podTemplateSpec)

	glog.V(3).Infof("Creating deployment %v", deploymentSpec.Template.GetName())
	glog.V(3).Infof("The component name is %v", componentName)

	if utils.ComponentExists(a.Client, componentName) {
		glog.V(3).Info("The component already exists, attempting to update it")
		_, err = a.Client.UpdateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully updated component %v", componentName)
	} else {
		_, err = a.Client.CreateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully created component %v", componentName)
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(uniqueStorages)
	if err != nil {
		return err
	}

	podSelector := fmt.Sprintf("component=%s", componentName)
	watchOptions := metav1.ListOptions{
		LabelSelector: podSelector,
	}

	_, err = a.Client.WaitAndGetPod(watchOptions, corev1.PodRunning, "Waiting for component to start")
	return err
}
