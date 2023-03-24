package init

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	dfutil "github.com/devfile/library/v2/pkg/util"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/init/backend"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type InitClient struct {
	// Backends
	flagsBackend       *backend.FlagsBackend
	interactiveBackend *backend.InteractiveBackend
	alizerBackend      *backend.AlizerBackend

	// Clients
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	registryClient   registry.Client
}

var _ Client = (*InitClient)(nil)

func NewInitClient(fsys filesystem.Filesystem, preferenceClient preference.Client, registryClient registry.Client, alizerClient alizer.Client) *InitClient {
	// We create the asker client and the backends here and not at the CLI level, as we want to hide these details to the CLI
	askerClient := asker.NewSurveyAsker()
	return &InitClient{
		flagsBackend:       backend.NewFlagsBackend(registryClient),
		interactiveBackend: backend.NewInteractiveBackend(askerClient, registryClient, alizerClient),
		alizerBackend:      backend.NewAlizerBackend(askerClient, alizerClient),
		fsys:               fsys,
		preferenceClient:   preferenceClient,
		registryClient:     registryClient,
	}
}

// GetFlags gets the flag specific to init operation so that it can correctly decide on the backend to be used
// It ignores all the flags except the ones specific to init operation, for e.g. verbosity flag
func (o *InitClient) GetFlags(flags map[string]string) map[string]string {
	initFlags := map[string]string{}
	for flag, value := range flags {
		if flag == backend.FLAG_NAME || flag == backend.FLAG_DEVFILE || flag == backend.FLAG_DEVFILE_REGISTRY || flag == backend.FLAG_STARTER || flag == backend.FLAG_DEVFILE_PATH || flag == backend.FLAG_DEVFILE_VERSION {
			initFlags[flag] = value
		}
	}
	return initFlags
}

// Validate calls Validate method of the adequate backend
func (o *InitClient) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags, fs, dir)
}

// SelectDevfile calls SelectDevfile methods of the adequate backend
func (o *InitClient) SelectDevfile(ctx context.Context, flags map[string]string, fs filesystem.Filesystem, dir string) (*api.DetectionResult, error) {
	var backend backend.InitBackend

	empty, err := location.DirIsEmpty(fs, dir)
	if err != nil {
		return nil, err
	}
	if empty && len(flags) == 0 {
		backend = o.interactiveBackend
	} else if len(flags) == 0 {
		backend = o.alizerBackend
	} else {
		backend = o.flagsBackend
	}
	location, err := backend.SelectDevfile(ctx, flags, fs, dir)
	if err != nil || location == nil {
		if backend == o.alizerBackend {
			// Fallback to the Interactive Mode if Alizer could not determine the Devfile.
			if err != nil {
				log.Warningf("Could not determine a Devfile based on the files in the current directory: %v", err)
			}
			return o.interactiveBackend.SelectDevfile(ctx, flags, fs, dir)
		}
		if err != nil {
			return nil, err
		}
		return nil, errors.New("unable to determine the devfile location")
	}

	return location, err
}

func (o *InitClient) DownloadDevfile(ctx context.Context, devfileLocation *api.DetectionResult, destDir string) (string, error) {
	destDevfile := filepath.Join(destDir, "devfile.yaml")
	if devfileLocation.DevfilePath != "" {
		return destDevfile, o.downloadDirect(devfileLocation.DevfilePath, destDevfile)
	} else {
		devfile := devfileLocation.Devfile
		if devfileLocation.DevfileVersion != "" {
			devfile = fmt.Sprintf("%s:%s", devfileLocation.Devfile, devfileLocation.DevfileVersion)
		}
		return destDevfile, o.downloadFromRegistry(ctx, devfileLocation.DevfileRegistry, devfile, destDir)
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
func (o *InitClient) downloadFromRegistry(ctx context.Context, registryName string, devfile string, dest string) error {
	// setting NewIndexSchema ensures that the Devfile library pulls registry based on the stack version
	registryOptions := segment.GetRegistryOptions(ctx)
	registryOptions.NewIndexSchema = true

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

	registries, err := o.registryClient.GetDevfileRegistries(registryName)
	if err != nil {
		return err
	}
	for _, reg := range registries {
		if forceRegistry && reg.Name == registryName {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, registryOptions)
			if err != nil {
				return err
			}
			downloadSpinner.End(true)
			return nil
		} else if !forceRegistry {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, registryOptions)
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
func (o *InitClient) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string, isEmptyDir bool) (*v1alpha2.StarterProject, error) {
	var backend backend.InitBackend

	if isEmptyDir && len(flags) == 0 {
		backend = o.interactiveBackend
	} else if len(flags) == 0 {
		backend = o.alizerBackend
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
func (o *InitClient) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {
	var backend backend.InitBackend

	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.PersonalizeName(devfile, flags)
}

func (o *InitClient) HandleApplicationPorts(devfileobj parser.DevfileObj, ports []int, flags map[string]string, fs filesystem.Filesystem, dir string) (parser.DevfileObj, error) {
	var backend backend.InitBackend
	onlyDevfile, err := location.DirContainsOnlyDevfile(fs, dir)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// Interactive mode since no flags are provided
	if len(flags) == 0 && !onlyDevfile {
		// Other files present in the directory; hence alizer is run
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.HandleApplicationPorts(devfileobj, ports, flags)
}

func (o *InitClient) PersonalizeDevfileConfig(devfileobj parser.DevfileObj, flags map[string]string, fs filesystem.Filesystem, dir string) (parser.DevfileObj, error) {
	var backend backend.InitBackend

	// Interactive mode since no flags are provided
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.PersonalizeDevfileConfig(devfileobj)
}

func (o *InitClient) SelectAndPersonalizeDevfile(ctx context.Context, flags map[string]string, contextDir string) (parser.DevfileObj, string, *api.DetectionResult, error) {
	devfileLocation, err := o.SelectDevfile(ctx, flags, o.fsys, contextDir)
	if err != nil {
		return parser.DevfileObj{}, "", nil, err
	}

	devfilePath, err := o.DownloadDevfile(ctx, devfileLocation, contextDir)
	if err != nil {
		return parser.DevfileObj{}, "", nil, fmt.Errorf("unable to download devfile: %w", err)
	}

	devfileObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: pointer.Bool(false)})
	if err != nil {
		return parser.DevfileObj{}, "", nil, fmt.Errorf("unable to parse devfile: %w", err)
	}

	devfileObj, err = o.HandleApplicationPorts(devfileObj, devfileLocation.ApplicationPorts, flags, o.fsys, contextDir)
	if err != nil {
		return parser.DevfileObj{}, "", nil, fmt.Errorf("unable to set application ports in devfile: %w", err)
	}

	devfileObj, err = o.PersonalizeDevfileConfig(devfileObj, flags, o.fsys, contextDir)
	if err != nil {
		return parser.DevfileObj{}, "", nil, fmt.Errorf("failed to configure devfile: %w", err)
	}
	return devfileObj, devfilePath, devfileLocation, nil
}

func (o *InitClient) InitDevfile(ctx context.Context, flags map[string]string, contextDir string,
	preInitHandlerFunc func(interactiveMode bool), newDevfileHandlerFunc func(newDevfileObj parser.DevfileObj) error) error {

	containsDevfile, err := location.DirectoryContainsDevfile(o.fsys, contextDir)
	if err != nil {
		return err
	}
	if containsDevfile {
		return nil
	}

	if preInitHandlerFunc != nil {
		preInitHandlerFunc(len(flags) == 0)
	}

	devfileObj, _, _, err := o.SelectAndPersonalizeDevfile(ctx, map[string]string{}, contextDir)
	if err != nil {
		return err
	}

	// Set the name in the devfile but do not write it yet.
	name, err := o.PersonalizeName(devfileObj, map[string]string{})
	if err != nil {
		return fmt.Errorf("failed to update the devfile's name: %w", err)
	}
	metadata := devfileObj.Data.GetMetadata()
	metadata.Name = name
	devfileObj.Data.SetMetadata(metadata)

	if newDevfileHandlerFunc != nil {
		err = newDevfileHandlerFunc(devfileObj)
	}

	return err
}
