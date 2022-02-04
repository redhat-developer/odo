package init

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/init/params"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

type InitClient struct {
	backends         []params.ParamsBuilder
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	registryClient   registry.Client
}

func NewInitClient(backends []params.ParamsBuilder, fsys filesystem.Filesystem, preferenceClient preference.Client, registryClient registry.Client) *InitClient {
	return &InitClient{
		backends:         backends,
		fsys:             fsys,
		preferenceClient: preferenceClient,
		registryClient:   registryClient,
	}
}

// SelectDevfile returns information about a devfile to download
func (o *InitClient) SelectDevfile(flags map[string]string) (*params.DevfileLocation, error) {
	for _, backend := range o.backends {
		if backend.IsAdequate(flags) {
			return backend.ParamsBuild()
		}
	}
	return nil, errors.New("no backend found to build init parameters. This should not happen")
}

func (o *InitClient) DownloadDevfile(devfileLocation *params.DevfileLocation, destDir string) (string, error) {
	destDevfile := filepath.Join(destDir, "devfile.yaml")
	if devfileLocation.DevfilePath != "" {
		return destDevfile, o.downloadDirect(devfileLocation.DevfilePath, destDevfile)
	} else {
		return destDevfile, o.downloadFromRegistry(devfileLocation.DevfileRegistry, devfileLocation.Devfile, destDir)
	}
}

// downloadDirect downloads a devfile at the provided URL and saves it in dest
func (o *InitClient) downloadDirect(URL string, dest string) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if strings.HasPrefix(parsedURL.Scheme, "http") {
		downloadSpinner := log.Spinnerf("Downloading devfile from %q", URL)
		defer downloadSpinner.End(false)
		params := util.HTTPRequestParams{
			URL: URL,
		}
		devfileData, err := o.registryClient.DownloadFileInMemory(params)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, devfileData, 0644)
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	} else {
		downloadSpinner := log.Spinnerf("Copying devfile from %q", URL)
		defer downloadSpinner.End(false)
		content, err := o.fsys.ReadFile(URL)
		if err != nil {
			return err
		}
		info, err := o.fsys.Stat(URL)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, content, info.Mode().Perm())
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	}

	return nil
}

// downloadFromRegistry downloads a devfile from the provided registry and saves it in dest
// If registryName is empty, will try to download the devfile from the list of registries in preferences
func (o *InitClient) downloadFromRegistry(registryName string, devfile string, dest string) error {
	var downloadSpinner *log.Status
	var forceRegistry bool
	if registryName == "" {
		downloadSpinner = log.Spinnerf("Downloading devfile %q", devfile)
		forceRegistry = false
	} else {
		downloadSpinner = log.Spinnerf("Downloading devfile %q from registry %q", devfile, registryName)
		forceRegistry = true
	}
	defer downloadSpinner.End(false)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry
	for _, reg = range *registries {
		if forceRegistry && reg.Name == registryName {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				return err
			}
			downloadSpinner.End(true)
			return nil
		} else if !forceRegistry {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				continue
			}
			downloadSpinner.End(true)
			return nil
		}
	}

	return fmt.Errorf("unable to find the registry with name %q", devfile)
}

// DownloadStarterProject downloads the starter project referenced in devfile and stores it in dest directory
// WARNING: This will first remove all the content of dest.
func (o *InitClient) DownloadStarterProject(devfile parser.DevfileObj, project string, dest string) error {
	projects, err := devfile.Data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}
	var prj v1alpha2.StarterProject
	var found bool
	for _, prj = range projects {
		if prj.Name == project {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("starter project %q does not exist in devfile", project)
	}
	downloadSpinner := log.Spinnerf("Downloading starter project %q", prj.Name)
	err = o.registryClient.DownloadStarterProject(&prj, "", dest, false)
	if err != nil {
		downloadSpinner.End(false)
		return err
	}
	downloadSpinner.End(true)
	return nil
}
