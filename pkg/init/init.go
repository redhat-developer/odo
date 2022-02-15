package init

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/init/backend"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type InitClient struct {
	// Backends
	flagsBackend       *backend.FlagsBackend
	interactiveBackend *backend.InteractiveBackend

	// Clients
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	registryClient   registry.Client
}

func NewInitClient(fsys filesystem.Filesystem, preferenceClient preference.Client, registryClient registry.Client) *InitClient {
	return &InitClient{
		flagsBackend:       backend.NewFlagsBackend(preferenceClient),
		interactiveBackend: backend.NewInteractiveBackend(asker.NewSurveyAsker(), catalog.NewCatalogClient(fsys, preferenceClient)),
		fsys:               fsys,
		preferenceClient:   preferenceClient,
		registryClient:     registryClient,
	}
}

// Validate calls Validate method of the adequate backend
func (o *InitClient) Validate(flags map[string]string) error {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
}

// SelectDevfile calls SelectDevfile methods of the adequate backend
func (o *InitClient) SelectDevfile(flags map[string]string) (*backend.DevfileLocation, error) {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectDevfile(flags)
}

func (o *InitClient) DownloadDevfile(devfileLocation *backend.DevfileLocation, destDir string) (string, error) {
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
		params := dfutil.HTTPRequestParams{
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

// SelectStarterProject calls SelectStarterProject methods of the adequate backend
func (o *InitClient) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectStarterProject(devfile, flags)
}

func (o *InitClient) DownloadStarterProject(starter *v1alpha2.StarterProject, dest string) error {
	downloadSpinner := log.Spinnerf("Downloading starter project %q", starter.Name)
	err := o.registryClient.DownloadStarterProject(starter, "", dest, false)
	if err != nil {
		downloadSpinner.End(false)
		return err
	}
	downloadSpinner.End(true)
	return nil
}

// PersonalizeName calls PersonalizeName methods of the adequate backend
func (o *InitClient) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	err := backend.PersonalizeName(devfile, flags)
	return err
}

func (o *InitClient) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) error {
	return o.interactiveBackend.PersonalizeDevfileConfig(devfileobj)
}
