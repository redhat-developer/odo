package registry

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/registry-library/library"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
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
func (o RegistryClient) ListDevfileStacks(registryName, devfileFlag, filterFlag string, detailsFlag bool) (DevfileStackList, error) {
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

	// Go through all the devfiles and filter based on:
	// What's in the name or description
	// The exact name of the devfile
	//
	// We also add additional details such as supported odo features (which we
	// manually http get) if the details flag has been passed in.
	for priorityNumber, registryDevfiles := range registrySlice {

		devfiles := []DevfileStack{}

		for _, devfile := range registryDevfiles {

			// Add the "priority" of the registry to the devfile
			devfile.Registry.Priority = priorityNumber

			if filterFlag != "" {
				if !strings.Contains(devfile.Name, filterFlag) && !strings.Contains(devfile.Description, filterFlag) {
					continue
				}
			}

			if devfileFlag != "" {
				if devfileFlag != devfile.Name {
					continue
				}
			}

			if detailsFlag {
				devfileData, err := o.retrieveDevfileDataFromRegistry(devfile.Registry.Name, devfile.Name)
				if err != nil {
					return *catalogDevfileList, err
				}
				devfile.DevfileData = &devfileData
			}

			devfiles = append(devfiles, devfile)
		}

		catalogDevfileList.Items = append(catalogDevfileList.Items, devfiles...)
	}

	// Sort catalogDevfileList.Items by:
	// 1. Priority of the registry (highest priority has highest index)
	// 2. Name of the devfile
	sort.Slice(catalogDevfileList.Items[:], func(i, j int) bool {
		if catalogDevfileList.Items[i].Name == catalogDevfileList.Items[j].Name {
			return catalogDevfileList.Items[i].Registry.Priority < catalogDevfileList.Items[j].Registry.Priority
		}
		return catalogDevfileList.Items[i].Name < catalogDevfileList.Items[j].Name
	})

	return *catalogDevfileList, nil
}

// getRegistryStacks retrieves the registry's index devfile stack entries
func getRegistryStacks(preferenceClient preference.Client, registry Registry) ([]DevfileStack, error) {
	isGithubregistry, err := registryUtil.IsGithubBasedRegistry(registry.URL)
	if err != nil {
		return nil, err
	}
	if isGithubregistry {
		return nil, registryUtil.ErrGithubRegistryNotSupported
	}
	// OCI-based registry
	devfileIndex, err := library.GetRegistryIndex(registry.URL, segment.GetRegistryOptions(), indexSchema.StackDevfileType)
	if err != nil {
		return nil, err
	}
	return createRegistryDevfiles(registry, devfileIndex)
}

func createRegistryDevfiles(registry Registry, devfileIndex []indexSchema.Schema) ([]DevfileStack, error) {
	registryDevfiles := make([]DevfileStack, 0, len(devfileIndex))
	for _, devfileIndexEntry := range devfileIndex {
		stackDevfile := DevfileStack{
			Name:            devfileIndexEntry.Name,
			DisplayName:     devfileIndexEntry.DisplayName,
			Description:     devfileIndexEntry.Description,
			Registry:        registry,
			Language:        devfileIndexEntry.Language,
			Tags:            devfileIndexEntry.Tags,
			ProjectType:     devfileIndexEntry.ProjectType,
			StarterProjects: devfileIndexEntry.StarterProjects,
			Version:         devfileIndexEntry.Version,
		}
		registryDevfiles = append(registryDevfiles, stackDevfile)
	}

	return registryDevfiles, nil
}

func (o RegistryClient) retrieveDevfileDataFromRegistry(registryName string, devfileName string) (api.DevfileData, error) {

	// Create random temporary file
	tmpFile, err := ioutil.TempDir("", "odo")
	if err != nil {
		return api.DevfileData{}, err
	}
	defer os.Remove(tmpFile)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry

	// Get the file and save it to the temporary file
	// Why do we do that?
	// 1. We need to get the file from the registry
	// 2. The devfile api library does not support saving in memory
	// 3. We need to get the file from the registry and save it to the temporary file
	// 4. We need to read the file from the temporary file, unmarshal it and then return the devfile data
	for _, reg = range *registries {
		if reg.Name == registryName {
			err = o.PullStackFromRegistry(reg.URL, devfileName, tmpFile, segment.GetRegistryOptions())
			if err != nil {
				return api.DevfileData{}, err
			}
		}
	}

	// Get the devfile yaml file from the directory
	devfileYamlFile := location.DevfileFilenamesProvider(tmpFile)

	// Parse and validate the file and return the devfile data
	devfileObj, err := devfile.ParseAndValidateFromFile(path.Join(tmpFile, devfileYamlFile))
	if err != nil {
		return api.DevfileData{}, err
	}

	// Convert DevfileObj to DevfileData
	// use api.GetDevfileData to get supported features
	return *api.GetDevfileData(devfileObj), nil
}
