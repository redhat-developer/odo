package storage

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
)

const (
	// OdoSourceVolume is the constant containing the name of the emptyDir volume containing the project source
	OdoSourceVolume = "odo-projects"

	// OdoSourceVolumeSize specifies size for odo source volume.
	OdoSourceVolumeSize = "2Gi"

	apiVersion = "odo.dev/v1alpha1"
)

// generic contains information required for all the Storage clients
type generic struct {
	appName       string
	componentName string
	localConfig   localConfigProvider.LocalConfigProvider
}

type ClientOptions struct {
	OCClient            occlient.Client
	LocalConfigProvider localConfigProvider.LocalConfigProvider
	Deployment          *v1.Deployment
}

type Client interface {
	Create(Storage) error
	Delete(string) error
	ListFromCluster() (StorageList, error)
	List() (StorageList, error)
}

// NewClient gets the appropriate Storage client based on the parameters
func NewClient(options ClientOptions) Client {
	var genericInfo generic

	if options.LocalConfigProvider != nil {
		genericInfo = generic{
			appName:       options.LocalConfigProvider.GetApplication(),
			componentName: options.LocalConfigProvider.GetName(),
			localConfig:   options.LocalConfigProvider,
		}
	}

	if options.Deployment != nil {
		genericInfo.appName = options.Deployment.Labels[applabels.ApplicationLabel]
		genericInfo.componentName = options.Deployment.Labels[labels.ComponentLabel]
	}

	return kubernetesClient{
		generic:    genericInfo,
		client:     options.OCClient,
		deployment: options.Deployment,
	}
}

// Push creates and deletes the required Storage
// it compares the local storage against the storage on the cluster
func Push(client Client, configProvider localConfigProvider.LocalConfigProvider) error {
	// list all the storage in the cluster
	storageClusterList := StorageList{}

	storageClusterList, err := client.ListFromCluster()
	if err != nil {
		return err
	}
	storageClusterNames := make(map[string]Storage)
	for _, storage := range storageClusterList.Items {
		storageClusterNames[storage.Name] = storage
	}

	// list all the storage in the config
	storageConfigNames := make(map[string]Storage)

	localStorage, err := configProvider.ListStorage()
	if err != nil {
		return err
	}
	for _, storage := range ConvertListLocalToMachine(localStorage).Items {
		storageConfigNames[storage.Name] = storage
	}

	// find storage to delete
	for storageName, storage := range storageClusterNames {
		val, ok := storageConfigNames[storageName]
		if !ok {
			// delete the pvc
			err = client.Delete(storage.Name)
			if err != nil {
				return err
			}
			log.Successf("Deleted storage %v from %v", storage.Name, configProvider.GetName())
			continue
		} else if storage.Name == val.Name {
			if val.Spec.Size != storage.Spec.Size {
				return errors.Errorf("config mismatch for storage with the same name %s", storage.Name)
			}
		}
	}

	// find storage to create
	for storageName, storage := range storageConfigNames {
		_, ok := storageClusterNames[storageName]
		if !ok {
			err := client.Create(storage)
			if err != nil {
				return err
			}
			log.Successf("Added storage %v to %v", storage.Name, configProvider.GetName())
		}
	}

	return err
}
