package catalog

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"

	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	registryLibrary "github.com/devfile/registry-support/registry-library/library"

	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/preference/registry/util"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type CatalogClient struct {
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
}

func NewCatalogClient(fsys filesystem.Filesystem, preferenceClient preference.Client) *CatalogClient {
	return &CatalogClient{
		fsys:             fsys,
		preferenceClient: preferenceClient,
	}
}

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func (o *CatalogClient) GetDevfileRegistries(registryName string) ([]Registry, error) {
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

// ListDevfileComponents lists all the available devfile components
func (o *CatalogClient) ListDevfileComponents(registryName string) (DevfileComponentTypeList, error) {
	catalogDevfileList := &DevfileComponentTypeList{}
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
	registrySlice := make([][]DevfileComponentType, len(catalogDevfileList.DevfileRegistries))
	for regPriority, reg := range catalogDevfileList.DevfileRegistries {
		// Load the devfile registry index.json
		registry := reg                 // Needed to prevent the lambda from capturing the value
		registryPriority := regPriority // Needed to prevent the lambda from capturing the value
		retrieveRegistryIndices.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
			registryDevfiles, err := getRegistryDevfiles(o.preferenceClient, registry)
			if err != nil {
				log.Warningf("Registry %s is not set up properly with error: %v, please check the registry URL and credential (refer `odo registry update --help`)\n", registry.Name, err)
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

// GetStarterProjectsNames returns the list of starter projects in a devfile,
// by temporarily downloading the devile
func (o *CatalogClient) GetStarterProjectsNames(details DevfileComponentType) ([]string, error) {
	tmpDir, err := o.fsys.TempDir("", "odoinit")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = o.fsys.RemoveAll(tmpDir)
	}()

	err = registryLibrary.PullStackFromRegistry(details.Registry.URL, details.Name, tmpDir, segment.GetRegistryOptions())
	if err != nil {
		return nil, err
	}

	devObj, err := devfile.ParseAndValidateFromFile(filepath.Join(tmpDir, "devfile.yaml"))
	if err != nil {
		return nil, err
	}
	starterProjects, err := devObj.Data.GetStarterProjects(parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(starterProjects))
	for _, starterProject := range starterProjects {
		names = append(names, starterProject.Name)
	}
	return names, err
}

// SearchComponent searches for the component
//TODO: Fix this to return devfile components
func (o *CatalogClient) SearchComponent(client kclient.ClientInterface, name string) ([]string, error) {
	//var result []string
	//componentList, err := ListDevfileComponents(client)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to list components")
	//}
	//
	//// do a partial search in all the components
	//for _, component := range componentList.Items {
	//	// we only show components that contain the search term and that have at least non-hidden tag
	//	// since a component with all hidden tags is not shown in the odo catalog list components either
	//	if strings.Contains(component.ObjectMeta.Name, name) && len(component.Spec.NonHiddenTags) > 0 {
	//		result = append(result, component.ObjectMeta.Name)
	//	}
	//}

	return []string{}, nil
}

// convertURL converts GitHub regular URL to GitHub raw URL, do nothing if the URL is not GitHub URL
// For example:
// GitHub regular URL: https://github.com/elsony/devfile-registry/tree/johnmcollier-crw
// GitHub raw URL: https://raw.githubusercontent.com/elsony/devfile-registry/johnmcollier-crw
func convertURL(URL string) (string, error) {
	url, err := url.Parse(URL)
	if err != nil {
		return "", err
	}

	if strings.Contains(url.Host, "github") && !strings.Contains(url.Host, "raw") {
		// Convert path part of the URL
		URLSlice := strings.Split(URL, "/")
		if len(URLSlice) > 2 && URLSlice[len(URLSlice)-2] == "tree" {
			// GitHub raw URL doesn't have "tree" structure in the URL, need to remove it
			URL = strings.Replace(URL, "/tree", "", 1)
		} else {
			// Add "master" branch for GitHub raw URL by default if branch is not specified
			URL = URL + "/master"
		}

		// Convert host part of the URL
		if url.Host == "github.com" {
			URL = strings.Replace(URL, "github.com", "raw.githubusercontent.com", 1)
		} else {
			URL = strings.Replace(URL, url.Host, "raw."+url.Host, 1)
		}
	}

	return URL, nil
}

const indexPath = "/devfiles/index.json"

// getRegistryDevfiles retrieves the registry's index devfile entries
func getRegistryDevfiles(preferenceClient preference.Client, registry Registry) ([]DevfileComponentType, error) {
	if !strings.Contains(registry.URL, "github") {
		// OCI-based registry
		devfileIndex, err := registryLibrary.GetRegistryIndex(registry.URL, segment.GetRegistryOptions(), indexSchema.StackDevfileType)
		if err != nil {
			return nil, err
		}
		return createRegistryDevfiles(registry, devfileIndex)
	}
	// Github-based registry
	URL, err := convertURL(registry.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to convert URL %s", registry.URL)
	}
	registry.URL = URL
	indexLink := registry.URL + indexPath
	request := util.HTTPRequestParams{
		URL: indexLink,
	}

	secure := registryUtil.IsSecure(preferenceClient, registry.Name)
	if secure {
		token, e := keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, registry.Name), registryUtil.RegistryUser)
		if e != nil {
			return nil, errors.Wrap(e, "unable to get secure registry credential from keyring")
		}
		request.Token = token
	}

	jsonBytes, err := util.HTTPGetRequest(request, preferenceClient.GetRegistryCacheTime())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to download the devfile index.json from %s", indexLink)
	}

	var devfileIndex []indexSchema.Schema
	err = json.Unmarshal(jsonBytes, &devfileIndex)
	if err != nil {
		if err := util.CleanDefaultHTTPCacheDir(); err != nil {
			log.Warning("Error while cleaning up cache dir.")
		}
		// we try once again
		jsonBytes, err := util.HTTPGetRequest(request, preferenceClient.GetRegistryCacheTime())
		if err != nil {
			return nil, errors.Wrapf(err, "unable to download the devfile index.json from %s", indexLink)
		}

		err = json.Unmarshal(jsonBytes, &devfileIndex)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal the devfile index.json from %s", indexLink)
		}
	}
	return createRegistryDevfiles(registry, devfileIndex)
}

func createRegistryDevfiles(registry Registry, devfileIndex []indexSchema.Schema) ([]DevfileComponentType, error) {
	registryDevfiles := make([]DevfileComponentType, 0, len(devfileIndex))
	for _, devfileIndexEntry := range devfileIndex {
		stackDevfile := DevfileComponentType{
			Name:        devfileIndexEntry.Name,
			DisplayName: devfileIndexEntry.DisplayName,
			Description: devfileIndexEntry.Description,
			Link:        devfileIndexEntry.Links["self"],
			Registry:    registry,
			Language:    devfileIndexEntry.Language,
			Tags:        devfileIndexEntry.Tags,
		}
		registryDevfiles = append(registryDevfiles, stackDevfile)
	}

	return registryDevfiles, nil
}
