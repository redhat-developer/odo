package storage

import (
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/log"
	v1 "k8s.io/api/apps/v1"
)

const (
	// OdoSourceVolume is the constant containing the name of the emptyDir volume containing the project source
	OdoSourceVolume = "odo-projects"

	// OdoSupervisordVolume is the constant containing the name of the emptyDir volume containing Supervisord shared data
	OdoSupervisordVolume = "odo-supervisord-shared-data"

	// OdoSourceVolumeSize specifies size for odo source volume.
	OdoSourceVolumeSize = "2Gi"
)

// generic contains information required for all the Storage clients
type generic struct {
	appName             string
	componentName       string
	localConfigProvider localConfigProvider.LocalConfigProvider
}

type ClientOptions struct {
	Client              kclient.ClientInterface
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
			appName:             options.LocalConfigProvider.GetApplication(),
			componentName:       options.LocalConfigProvider.GetName(),
			localConfigProvider: options.LocalConfigProvider,
		}
	}

	if options.Deployment != nil {
		genericInfo.appName = options.Deployment.Labels[applabels.ApplicationLabel]
		genericInfo.componentName = options.Deployment.Labels[labels.ComponentKubernetesInstanceLabel]
	}

	return kubernetesClient{
		generic:    genericInfo,
		client:     options.Client,
		deployment: options.Deployment,
	}
}

// Push creates and deletes the required persistent storages and returns the list of ephemeral storages
// it compares the local storage against the storage on the cluster
func Push(client Client, configProvider localConfigProvider.LocalConfigProvider) (ephemerals map[string]Storage, _ error) {
	// list all the storage in the cluster
	storageClusterList := StorageList{}

	storageClusterList, err := client.ListFromCluster()
	if err != nil {
		return nil, err
	}
	storageClusterNames := make(map[string]Storage)
	for _, storage := range storageClusterList.Items {
		storageClusterNames[storage.Name] = storage
	}

	// list the persistent storages in the config
	persistentConfigNames := make(map[string]Storage)
	// list the ephemeral storages
	ephemeralConfigNames := make(map[string]Storage)

	localStorage, err := configProvider.ListStorage()
	if err != nil {
		return nil, err
	}
	for _, storage := range ConvertListLocalToMachine(localStorage).Items {
		if storage.Spec.Ephemeral == nil || storage.Spec.Ephemeral != nil && !*storage.Spec.Ephemeral {
			persistentConfigNames[storage.Name] = storage
		} else {
			ephemeralConfigNames[storage.Name] = storage
		}
	}

	// find storage to delete
	for storageName, storage := range storageClusterNames {
		val, ok := persistentConfigNames[storageName]
		if !ok {
			// delete the pvc
			err = client.Delete(storage.Name)
			if err != nil {
				return nil, err
			}
			log.Successf("Deleted storage %v from %v", storage.Name, configProvider.GetName())
			continue
		} else if storage.Name == val.Name {
			if val.Spec.Size != storage.Spec.Size {
				return nil, errors.Errorf("config mismatch for storage with the same name %s", storage.Name)
			}
		}
	}

	// find storage to create
	for storageName, storage := range persistentConfigNames {
		_, ok := storageClusterNames[storageName]
		if ok {
			continue
		}
		if e := client.Create(storage); e != nil {
			return nil, e
		}
		log.Successf("Added storage %v to %v", storage.Name, configProvider.GetName())
	}

	return ephemeralConfigNames, nil
}
