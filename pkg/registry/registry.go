package registry

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/pkg/util"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/registry-library/library"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
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

var _ Client = (*RegistryClient)(nil)

func NewRegistryClient(fsys filesystem.Filesystem, preferenceClient preference.Client) RegistryClient {
	return RegistryClient{
		fsys:             fsys,
		preferenceClient: preferenceClient,
	}
}

// PullStackFromRegistry pulls stack from registry with all stack resources (all media types) to the destination directory
func (o RegistryClient) PullStackFromRegistry(registry string, stack string, destDir string, options library.RegistryOptions) error {
	klog.V(3).Infof("sending telemetry data: %#v", options.Telemetry)
	return library.PullStackFromRegistry(registry, stack, destDir, options)
}

// DownloadFileInMemory uses the url to download the file and return bytes
func (o RegistryClient) DownloadFileInMemory(params dfutil.HTTPRequestParams) ([]byte, error) {
	return util.DownloadFileInMemory(params)
}

// DownloadStarterProject downloads a starter project referenced in devfile
// This will first remove the content of the contextDir
func (o RegistryClient) DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) error {
	return DownloadStarterProject(starterProject, decryptedToken, contextDir, verbose)
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func (o RegistryClient) GetDevfileRegistries(registryName string) ([]api.Registry, error) {
	var devfileRegistries []api.Registry

	hasName := len(registryName) != 0
	if o.preferenceClient.RegistryList() != nil {
		registryList := o.preferenceClient.RegistryList()
		for _, registry := range registryList {
			if hasName {
				if registryName == registry.Name {
					reg := api.Registry{
						Name:   registry.Name,
						URL:    registry.URL,
						Secure: registry.Secure,
					}
					devfileRegistries = append(devfileRegistries, reg)
					return devfileRegistries, nil
				}
			} else {
				reg := api.Registry{
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
func (o RegistryClient) ListDevfileStacks(ctx context.Context, registryName, devfileFlag, filterFlag string, detailsFlag bool) (DevfileStackList, error) {
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
	registrySlice := make([][]api.DevfileStack, len(catalogDevfileList.DevfileRegistries))
	for regPriority, reg := range catalogDevfileList.DevfileRegistries {
		// Load the devfile registry index.json
		registry := reg                 // Needed to prevent the lambda from capturing the value
		registryPriority := regPriority // Needed to prevent the lambda from capturing the value
		retrieveRegistryIndices.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
			registryDevfiles, err := getRegistryStacks(ctx, o.preferenceClient, registry)
			if err != nil {
				log.Warningf("Registry %s is not set up properly with error: %v, please check the registry URL, and credential and remove add the registry again (refer to `odo preference add registry --help`)\n", registry.Name, err)
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

		devfiles := []api.DevfileStack{}

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
				devfileData, err := o.retrieveDevfileDataFromRegistry(ctx, devfile.Registry.Name, devfile.Name)
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
func getRegistryStacks(ctx context.Context, preferenceClient preference.Client, registry api.Registry) ([]api.DevfileStack, error) {
	isGithubregistry, err := IsGithubBasedRegistry(registry.URL)
	if err != nil {
		return nil, err
	}
	if isGithubregistry {
		return nil, &ErrGithubRegistryNotSupported{}
	}
	// OCI-based registry
	options := segment.GetRegistryOptions(ctx)
	options.NewIndexSchema = true
	devfileIndex, err := library.GetRegistryIndex(registry.URL, options, indexSchema.StackDevfileType)
	if err != nil {
		// Fallback to the "old" index
		klog.V(3).Infof("error while accessing the v2index endpoint for registry %s (%s) => falling back to the old index endpoint: %v",
			registry.Name, registry.URL, err)
		options.NewIndexSchema = false
		devfileIndex, err = library.GetRegistryIndex(registry.URL, options, indexSchema.StackDevfileType)
		if err != nil {
			return nil, err
		}
	}
	return createRegistryDevfiles(registry, devfileIndex)
}

func createRegistryDevfiles(registry api.Registry, devfileIndex []indexSchema.Schema) ([]api.DevfileStack, error) {
	registryDevfiles := make([]api.DevfileStack, 0, len(devfileIndex))
	for _, devfileIndexEntry := range devfileIndex {
		stackDevfile := api.DevfileStack{
			Name:                   devfileIndexEntry.Name,
			DisplayName:            devfileIndexEntry.DisplayName,
			Description:            devfileIndexEntry.Description,
			Registry:               registry,
			Language:               devfileIndexEntry.Language,
			Tags:                   devfileIndexEntry.Tags,
			ProjectType:            devfileIndexEntry.ProjectType,
			DefaultStarterProjects: devfileIndexEntry.StarterProjects,
			DefaultVersion:         devfileIndexEntry.Version,
		}
		for _, v := range devfileIndexEntry.Versions {
			if v.Default {
				// There should be only 1 default version. But if there is more than one, the last one will be used.
				stackDevfile.DefaultVersion = v.Version
				stackDevfile.DefaultStarterProjects = v.StarterProjects
			}
			stackDevfile.Versions = append(stackDevfile.Versions, api.DevfileStackVersion{
				IsDefault:       v.Default,
				Version:         v.Version,
				SchemaVersion:   v.SchemaVersion,
				StarterProjects: v.StarterProjects,
			})
		}
		sort.Slice(stackDevfile.Versions, func(i, j int) bool {
			vi, err := semver.Make(stackDevfile.Versions[i].Version)
			if err != nil {
				return false
			}
			vj, err := semver.Make(stackDevfile.Versions[j].Version)
			if err != nil {
				return false
			}
			return vi.LT(vj)
		})

		registryDevfiles = append(registryDevfiles, stackDevfile)
	}

	return registryDevfiles, nil
}

func (o RegistryClient) retrieveDevfileDataFromRegistry(ctx context.Context, registryName string, devfileName string) (api.DevfileData, error) {

	// Create random temporary file
	tmpFile, err := ioutil.TempDir("", "odo")
	if err != nil {
		return api.DevfileData{}, err
	}
	defer os.Remove(tmpFile)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry
	registryOptions := segment.GetRegistryOptions(ctx)
	registryOptions.NewIndexSchema = true
	// Get the file and save it to the temporary file
	// Why do we do that?
	// 1. We need to get the file from the registry
	// 2. The devfile api library does not support saving in memory
	// 3. We need to get the file from the registry and save it to the temporary file
	// 4. We need to read the file from the temporary file, unmarshal it and then return the devfile data
	for _, reg = range registries {
		if reg.Name == registryName {
			err = o.PullStackFromRegistry(reg.URL, devfileName, tmpFile, registryOptions)
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
