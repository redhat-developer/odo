package component

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
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

	// Get a list of all the unique volume names
	componentAliasToVolumes := utils.GetVolumes(a.Devfile)
	var uniqueVolumes []common.Volume
	processedVolumes := make(map[string]bool)
	for _, volumes := range componentAliasToVolumes {
		for _, vol := range volumes {
			if _, ok := processedVolumes[*vol.Name]; !ok {
				processedVolumes[*vol.Name] = true
				uniqueVolumes = append(uniqueVolumes, vol)
			}
		}
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(uniqueVolumes)
	if err != nil {
		return err
	}
	volumeNameToPVCName := stoAdapter.(storage.HelperAdapter).GetVolumeNameToPVCName()
	// Add PVC and Volume Mounts to the podTemplateSpec
	err = kclient.AddPVCAndVolumeMount(podTemplateSpec, volumeNameToPVCName, componentAliasToVolumes)
	if err != nil {
		return err
	}

	deploymentSpec := kclient.GenerateDeploymentSpec(*podTemplateSpec)

	glog.V(3).Infof("Creating deployment %v", deploymentSpec.Template.GetName())
	glog.V(3).Infof("The component name is %v", componentName)

	var deployment *appsv1.Deployment
	if utils.ComponentExists(a.Client, componentName) {
		glog.V(3).Info("The component already exists, attempting to update it")
		deployment, err = a.Client.UpdateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully updated component %v", componentName)
	} else {
		deployment, err = a.Client.CreateDeployment(*deploymentSpec)
		if err != nil {
			return err
		}
		glog.V(3).Infof("Successfully created component %v", componentName)
	}

	// Get owner reference of the deployment
	ownerReference := kclient.GenerateOwnerReference(deployment)

	// Update component pvcs with the owner references
	err = a.Client.UpdateStorageOwnerReference(volumeNameToPVCName, ownerReference)
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
