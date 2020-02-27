package storage

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/utils"
	"github.com/openshift/odo/pkg/kclient"
)

// New instantiantes a storage adapter
func New(adapterContext common.AdapterContext, client kclient.Client) Adapter {
	return Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a storage adapter implementation for Kubernetes
type Adapter struct {
	Client kclient.Client
	common.AdapterContext
}

// Start creates the component pvc storage if it does not exist and adds them to the pod template spec
func (a Adapter) Start(podTemplateSpec *corev1.PodTemplateSpec) (err error) {
	componentName := a.ComponentName

	componentAliasToVolumes := utils.GetVolumes(a.Devfile)

	// Get a list of all the unique volume names
	var uniqueVolumes []string
	processedVolumes := make(map[string]bool)
	for _, volumes := range componentAliasToVolumes {
		for _, vol := range volumes {
			if _, ok := processedVolumes[*vol.Name]; !ok {
				processedVolumes[*vol.Name] = true
				uniqueVolumes = append(uniqueVolumes, *vol.Name)
			}
		}
	}

	// createComponentStorage creates PVC from the unique Devfile volumes if it does not exist and returns a map of volume name to the PVC created
	volumeNameToPVC, err := CreateComponentStorage(&a.Client, uniqueVolumes, componentName)
	if err != nil {
		return err
	}

	// Add PVC and Volume Mounts to the podTemplateSpec
	err = kclient.AddPVCAndVolumeMount(podTemplateSpec, volumeNameToPVC, componentAliasToVolumes)
	if err != nil {
		return err
	}

	return
}
