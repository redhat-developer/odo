package registry

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	dfutil "github.com/devfile/library/v2/pkg/util"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	"github.com/devfile/registry-support/registry-library/library"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type RegistryClient struct {
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	kubeClient       kclient.ClientInterface
}

var _ Client = (*RegistryClient)(nil)

const (
	CONFLICT_DIR_NAME = "CONFLICT_STARTER_PROJECT"
)

func NewRegistryClient(fsys filesystem.Filesystem, preferenceClient preference.Client, kubeClient kclient.ClientInterface) RegistryClient {
	return RegistryClient{
		fsys:             fsys,
		preferenceClient: preferenceClient,
		kubeClient:       kubeClient,
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
// There are 3 cases to consider here:
// Case 1: If there is devfile in the starterproject, replace all the contents of contextDir with that of the starterproject; warn about this
// Case 2: If there is no devfile, and there is no conflict between the contents of contextDir and starterproject, then copy the contents of the starterproject into contextDir.
// Case 3: If there is no devfile, and there is conflict between the contents of contextDir and starterproject, copy contents of starterproject into a dir named CONFLICT_STARTER_PROJECT; warn about this
func (o RegistryClient) DownloadStarterProject(starterProject *devfilev1.StarterProject, decryptedToken string, contextDir string, verbose bool) (containsDevfile bool, err error) {
	// Let the project be downloaded in a temp directory
	starterProjectTmpDir, err := o.fsys.TempDir("", "odostarterproject")
	if err != nil {
		return containsDevfile, err
	}
	defer func() {
		err = o.fsys.RemoveAll(starterProjectTmpDir)
		if err != nil {
			klog.V(2).Infof("failed to delete temporary starter project dir %s; cause: %s", starterProjectTmpDir, err.Error())
		}
	}()
	err = DownloadStarterProject(o.fsys, starterProject, decryptedToken, starterProjectTmpDir, verbose)
	if err != nil {
		return containsDevfile, err
	}

	// Case 1: If there is devfile in the starterproject, replace all the contents of contextDir with that of the starterproject; warn about this
	if containsDevfile, err = location.DirectoryContainsDevfile(o.fsys, starterProjectTmpDir); err != nil {
		return containsDevfile, err
	}
	if containsDevfile {
		fmt.Println()
		log.Warning("A Devfile is present inside the starter project; replacing the entire content of the current directory with the starter project")
		err = removeDirectoryContents(contextDir, o.fsys)
		if err != nil {
			return containsDevfile, fmt.Errorf("failed to delete contents of the current directory; cause %w", err)
		}
		return containsDevfile, util.CopyDirWithFS(starterProjectTmpDir, contextDir, o.fsys)
	}

	// Case 2: If there is no devfile, and there is no conflict between the contents of contextDir and starterproject, then copy the contents of the starterproject into contextDir.
	// Case 3: If there is no devfile, and there is conflict between the contents of contextDir and starterproject, copy contents of starterproject into a dir named CONFLICT_STARTER_PROJECT; warn about this
	var conflictingFiles []string
	conflictingFiles, err = getConflictingFiles(starterProjectTmpDir, contextDir, o.fsys)
	if err != nil {
		return containsDevfile, err
	}

	// Case 2
	if len(conflictingFiles) == 0 {
		return containsDevfile, util.CopyDirWithFS(starterProjectTmpDir, contextDir, o.fsys)
	}

	// Case 3
	conflictingDirPath := filepath.Join(contextDir, CONFLICT_DIR_NAME)
	err = o.fsys.MkdirAll(conflictingDirPath, 0750)
	if err != nil {
		return containsDevfile, err
	}

	err = util.CopyDirWithFS(starterProjectTmpDir, conflictingDirPath, o.fsys)
	if err != nil {
		return containsDevfile, err
	}
	fmt.Println()
	log.Warningf("There are conflicting files (%s) between starter project and the current directory, hence the starter project has been copied to %s", strings.Join(conflictingFiles, ", "), conflictingDirPath)

	return containsDevfile, nil
}

// removeDirectoryContents attempts to remove dir contents without deleting the directory itself, unlike os.RemoveAll() method
func removeDirectoryContents(path string, fsys filesystem.Filesystem) error {
	dir, err := fsys.ReadDir(path)
	if err != nil {
		return err
	}
	for _, f := range dir {
		// a bit of cheating by using absolute file name to make sure this works with a fake filesystem, especially a memory map which is used by our unit tests
		// memorymap's Name() method trims the full path and returns just the file name, which then becomes impossible to find by the RemoveAll method that looks for prefix
		// See: https://github.com/redhat-developer/odo/blob/d717421494f746a5cb12da135f561d12750935f3/vendor/github.com/spf13/afero/memmap.go#L282
		absFileName := filepath.Join(path, f.Name())
		err = fsys.RemoveAll(absFileName)
		if err != nil {
			return fmt.Errorf("failed to remove %s; cause: %w", absFileName, err)
		}
	}

	return nil
}

// getConflictingFiles fetches the contents of the two directories in question and compares them to check for conflicting files.
// it returns a list of conflicting files (if any) along with an error (if any).
func getConflictingFiles(spDir, contextDir string, fsys filesystem.Filesystem) (conflictingFiles []string, err error) {
	var (
		contextDirMap = map[string]struct{}{}
	)
	// walk through the contextDir, trim the file path from the file name and append it to a map
	err = fsys.Walk(contextDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to fetch contents of dir %s; cause: %w", contextDirMap, err)
		}
		if info.IsDir() {
			return nil
		}
		path = strings.TrimPrefix(path, contextDir)
		contextDirMap[path] = struct{}{}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk %s dir; cause: %w", contextDir, err)
	}

	// walk through the starterproject dir, trim the file path from file name, and check if it exists in the contextDir map;
	// if it does, it is a conflicting file, hence append it to the conflictingFiles list.
	err = fsys.Walk(spDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to fetch contents of dir %s; cause: %w", spDir, err)
		}
		if info.IsDir() {
			return nil
		}
		path = strings.TrimPrefix(path, spDir)
		if _, ok := contextDirMap[path]; ok {
			conflictingFiles = append(conflictingFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk %s dir; cause: %w", spDir, err)
	}

	return conflictingFiles, nil
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func (o RegistryClient) GetDevfileRegistries(registryName string) ([]api.Registry, error) {
	var allRegistries []api.Registry

	if o.kubeClient != nil {
		clusterRegistries, err := o.kubeClient.GetRegistryList()
		if err != nil {
			// #6636 : errors should not be blocking
			klog.V(3).Infof("failed to get Devfile registries from the cluster: %v", err)
		} else {
			allRegistries = append(allRegistries, clusterRegistries...)
		}
	}
	allRegistries = append(allRegistries, o.preferenceClient.RegistryList()...)

	hasName := registryName != ""
	var result []api.Registry
	for _, registry := range allRegistries {
		if hasName {
			if registryName == registry.Name {
				reg := api.Registry{
					Name:   registry.Name,
					URL:    registry.URL,
					Secure: registry.Secure,
				}
				result = append(result, reg)
				return result, nil
			}
			continue
		}
		reg := api.Registry{
			Name:   registry.Name,
			URL:    registry.URL,
			Secure: registry.Secure,
		}
		result = append(result, reg)
	}

	return result, nil
}

// ListDevfileStacks lists all the available devfile stacks in devfile registry
// When `withDevfileContent` and `detailsFlag` are both true, another HTTP call is executed to download the Devfile
func (o RegistryClient) ListDevfileStacks(ctx context.Context, registryName, devfileFlag, filterFlag string, detailsFlag bool, withDevfileContent bool) (DevfileStackList, error) {
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
			registryDevfiles, err := getRegistryStacks(ctx, registry)
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
	for priorityNumber, registryDevfiles := range registrySlice {

		devfiles := []api.DevfileStack{}

		for _, devfile := range registryDevfiles {

			// Add the "priority" of the registry to the devfile
			devfile.Registry.Priority = priorityNumber

			if filterFlag != "" {
				containsArch := func(s string) bool {
					for _, arch := range devfile.Architectures {
						if strings.Contains(arch, s) {
							return true
						}
					}
					return false
				}
				if !strings.Contains(devfile.Name, filterFlag) && !strings.Contains(devfile.Description, filterFlag) && !containsArch(filterFlag) {
					continue
				}
			}

			if devfileFlag != "" {
				if devfileFlag != devfile.Name {
					continue
				}
			}

			// We are fetching the Devfile content only when `--details` and `-o json` flags are used
			if detailsFlag && withDevfileContent {
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
func getRegistryStacks(ctx context.Context, registry api.Registry) ([]api.DevfileStack, error) {
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
			Architectures:          devfileIndexEntry.Architectures,
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
				CommandGroups:   v.CommandGroups,
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
	tmpFile, err := o.fsys.TempDir("", "odo")
	if err != nil {
		return api.DevfileData{}, err
	}
	defer func() {
		err = o.fsys.RemoveAll(tmpFile)
		if err != nil {
			klog.V(2).Infof("failed to delete temporary starter project dir %s; cause: %s", tmpFile, err.Error())
		}
	}()

	registries, err := o.GetDevfileRegistries(registryName)
	if err != nil {
		return api.DevfileData{}, err
	}
	registryOptions := segment.GetRegistryOptions(ctx)
	registryOptions.NewIndexSchema = true
	// Get the file and save it to the temporary file
	// Why do we do that?
	// 1. We need to get the file from the registry
	// 2. The devfile api library does not support saving in memory
	// 3. We need to get the file from the registry and save it to the temporary file
	// 4. We need to read the file from the temporary file, unmarshal it and then return the devfile data
	for _, reg := range registries {
		if reg.Name == registryName {
			err = o.PullStackFromRegistry(reg.URL, devfileName, tmpFile, registryOptions)
			if err != nil {
				return api.DevfileData{}, err
			}
		}
	}

	// Get the devfile yaml file from the directory
	devfileYamlFile := location.DevfileFilenamesProvider(o.fsys, tmpFile)

	// Parse and validate the file and return the devfile data
	devfileObj, err := devfile.ParseAndValidateFromFile(path.Join(tmpFile, devfileYamlFile), "", true)
	if err != nil {
		return api.DevfileData{}, err
	}

	// Convert DevfileObj to DevfileData
	// use api.GetDevfileData to get supported features
	devfileData, err := api.GetDevfileData(devfileObj)
	if err != nil {
		return api.DevfileData{}, err
	}

	return *devfileData, nil
}
