package catalog

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/redhat-developer/odo/pkg/segment"

	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/zalando/go-keyring"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"

	indexSchema "github.com/devfile/registry-support/index/generator/schema"
	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/pkg/errors"
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/util"
)

// GetDevfileRegistries gets devfile registries from preference file,
// if registry name is specified return the specific registry, otherwise return all registries
func GetDevfileRegistries(registryName string) ([]Registry, error) {
	var devfileRegistries []Registry

	cfg, err := preference.New()
	if err != nil {
		return nil, err
	}

	hasName := len(registryName) != 0
	if cfg.OdoSettings.RegistryList != nil {
		registryList := *cfg.OdoSettings.RegistryList
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
func getRegistryDevfiles(registry Registry) ([]DevfileComponentType, error) {
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
	secure, err := registryUtil.IsSecure(registry.Name)
	if err != nil {
		return nil, err
	}
	if secure {
		token, e := keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, registry.Name), registryUtil.RegistryUser)
		if e != nil {
			return nil, errors.Wrap(e, "unable to get secure registry credential from keyring")
		}
		request.Token = token
	}

	cfg, err := preference.New()
	if err != nil {
		return nil, err
	}

	jsonBytes, err := util.HTTPGetRequest(request, cfg.GetRegistryCacheTime())
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
		jsonBytes, err := util.HTTPGetRequest(request, cfg.GetRegistryCacheTime())
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

// ListDevfileComponents lists all the available devfile components
func ListDevfileComponents(registryName string) (DevfileComponentTypeList, error) {
	catalogDevfileList := &DevfileComponentTypeList{}
	var err error

	// TODO: consider caching registry information for better performance since it should be fairly stable over time
	// Get devfile registries
	catalogDevfileList.DevfileRegistries, err = GetDevfileRegistries(registryName)
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
			registryDevfiles, err := getRegistryDevfiles(registry)
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

// SearchComponent searches for the component
//TODO: Fix this to return devfile components
func SearchComponent(client kclient.ClientInterface, name string) ([]string, error) {
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

// ListOperatorServices fetches a list of Operators from the cluster and
// returns only those Operators which are successfully installed on the cluster
func ListOperatorServices(client kclient.ClientInterface) (*olm.ClusterServiceVersionList, error) {
	var csvList olm.ClusterServiceVersionList

	// first check for CSV support
	csvSupport, err := client.IsCSVSupported()
	if !csvSupport || err != nil {
		return &csvList, err
	}

	allCsvs, err := client.ListClusterServiceVersions()
	if err != nil {
		return &csvList, err
	}

	// now let's filter only those csvs which are successfully installed
	csvList.TypeMeta = allCsvs.TypeMeta
	csvList.ListMeta = allCsvs.ListMeta
	for _, csv := range allCsvs.Items {
		if csv.Status.Phase == "Succeeded" {
			csvList.Items = append(csvList.Items, csv)
		}
	}

	return &csvList, nil
}
