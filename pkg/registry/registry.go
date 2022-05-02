package registry

import (
	"sync"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/registry-library/library"
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/preference/registry/util"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type RegistryClient struct {
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
}

func NewRegistryClient(fsys filesystem.Filesystem, preferenceClient preference.Client) RegistryClient {
	return RegistryClient{
		fsys:             fsys,
		preferenceClient: preferenceClient,
	}
}

// PullStackFromRegistry pulls stack from registry with all stack resources (all media types) to the destination directory
func (o RegistryClient) PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error {
	return library.PullStackFromRegistry(registry, stack, destDir, options)
}

// DownloadFileInMemory uses the url to download the file and return bytes
func (o RegistryClient) DownloadFileInMemory(params dfutil.HTTPRequestParams) ([]byte, error) {
	return util.DownloadFileInMemory(params)
}

// DownloadStarterProject downloads a starter project referenced in devfile
// This will first remove the content of the contextDir
func (o RegistryClient) DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error {
	return component.DownloadStarterProject(starterProject, decryptedToken, contextDir, verbose)
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func (o RegistryClient) GetDevfileRegistries(registryName string) ([]Registry, error) {
	var devfileRegistries []Registry

	hasName := len(registryName) != 0
	if o.preferenceClient.RegistryList() != nil {
		registryList := *o.preferenceClient.RegistryList()
		// Loop backwards here to ensure the registry display order is correct (display latest newly added registry firstly)
		for i := len(registryList) - 1; i >= 0; i-- {
			registry := registryList[i]
			if hasName {
				if registryName == registry.Name {
					reg := Registry{
						Name:   registry.Name,
						URL:    registry.URL,
						Secure: registry.Secure,
					}
					devfileRegistries = append(devfileRegistries, reg)
					return devfileRegistries, nil
				}
			} else {
				reg := Registry{
					Name:   registry.Name,
					URL:    registry.URL,
					Secure: registry.Secure,
				}
				devfileRegistries = append(devfileRegistries, reg)
			}
		}
	} else {
		return nil, nil
	}

	return devfileRegistries, nil
}

// ListDevfileStacks lists all the available devfile stacks in devfile registry
func (o RegistryClient) ListDevfileStacks(registryName string) (DevfileStackList, error) {
	catalogDevfileList := &DevfileStackList{}
	var err error

	// TODO: consider caching registry information for better performance since it should be fairly stable over time
	// Get devfile registries
	catalogDevfileList.DevfileRegistries, err = o.GetDevfileRegistries(registryName)
	if err != nil {
		return *catalogDevfileList, err
	}
	if catalogDevfileList.DevfileRegistries == nil {
		return *catalogDevfileList, nil
	}

	// first retrieve the indices for each registry, concurrently
	devfileIndicesMutex := &sync.Mutex{}
	retrieveRegistryIndices := util.NewConcurrentTasks(len(catalogDevfileList.DevfileRegistries))

	// The 2D slice index is the priority of the registry (highest priority has highest index)
	// and the element is the devfile slice that belongs to the registry
	registrySlice := make([][]DevfileStack, len(catalogDevfileList.DevfileRegistries))
	for regPriority, reg := range catalogDevfileList.DevfileRegistries {
		// Load the devfile registry index.json
		registry := reg                 // Needed to prevent the lambda from capturing the value
		registryPriority := regPriority // Needed to prevent the lambda from capturing the value
		retrieveRegistryIndices.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
			registryDevfiles, err := getRegistryStacks(o.preferenceClient, registry)
			if err != nil {
				log.Warningf("Registry %s is not set up properly with error: %v, please check the registry URL and credential (refer `odo preference registry update --help`)\n", registry.Name, err)
				return
			}

			devfileIndicesMutex.Lock()
			registrySlice[registryPriority] = registryDevfiles
			devfileIndicesMutex.Unlock()
		}})
	}
	if err := retrieveRegistryIndices.Run(); err != nil {
		return *catalogDevfileList, err
	}

	for _, registryDevfiles := range registrySlice {
		catalogDevfileList.Items = append(catalogDevfileList.Items, registryDevfiles...)
	}

	return *catalogDevfileList, nil
}

// const indexPath = "/devfiles/index.json" :Not sure if this should be removed but it is not used for now

// getRegistryStacks retrieves the registry's index devfile stack entries
func getRegistryStacks(preferenceClient preference.Client, registry Registry) ([]DevfileStack, error) {
	if !registryUtil.IsGithubBasedRegistry(registry.URL) {
		// OCI-based registry
		devfileIndex, err := library.GetRegistryIndex(registry.URL, segment.GetRegistryOptions(), indexSchema.StackDevfileType)
		if err != nil {
			return nil, err
		}
		return createRegistryDevfiles(registry, devfileIndex)
	}
	return nil, registryUtil.ErrGithubRegistryNotSupported
}

func createRegistryDevfiles(registry Registry, devfileIndex []indexSchema.Schema) ([]DevfileStack, error) {
	registryDevfiles := make([]DevfileStack, 0, len(devfileIndex))
	for _, devfileIndexEntry := range devfileIndex {
		stackDevfile := DevfileStack{
			Name:        devfileIndexEntry.Name,
			DisplayName: devfileIndexEntry.DisplayName,
			Description: devfileIndexEntry.Description,
			Link:        devfileIndexEntry.Links["self"],
			Registry:    registry,
			Language:    devfileIndexEntry.Language,
			Tags:        devfileIndexEntry.Tags,
			ProjectType: devfileIndexEntry.ProjectType,
		}
		registryDevfiles = append(registryDevfiles, stackDevfile)
	}

	return registryDevfiles, nil
}
