package backend

import (
	"context"
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/registry"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/v2/pkg/util"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const (
	FLAG_NAME             = "name"
	FLAG_DEVFILE          = "devfile"
	FLAG_DEVFILE_REGISTRY = "devfile-registry"
	FLAG_STARTER          = "starter"
	FLAG_DEVFILE_PATH     = "devfile-path"
	FLAG_DEVFILE_VERSION  = "devfile-version"
)

// FlagsBackend is a backend that will extract all needed information from flags passed to the command
type FlagsBackend struct {
	registryClient registry.Client
}

var _ InitBackend = (*FlagsBackend)(nil)

func NewFlagsBackend(registryClient registry.Client) *FlagsBackend {
	return &FlagsBackend{
		registryClient: registryClient,
	}
}

func (o *FlagsBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	if flags[FLAG_NAME] == "" {
		return errors.New("missing --name parameter: please add --name <name> to specify a name for the component")
	}
	if flags[FLAG_DEVFILE] == "" && flags[FLAG_DEVFILE_PATH] == "" {
		return errors.New("either --devfile or --devfile-path parameter should be specified")
	}
	if flags[FLAG_DEVFILE] != "" && flags[FLAG_DEVFILE_PATH] != "" {
		return errors.New("only one of --devfile or --devfile-path parameter should be specified")
	}

	registryName := flags[FLAG_DEVFILE_REGISTRY]
	if registryName != "" {
		registries, err := o.registryClient.GetDevfileRegistries(registryName)
		if err != nil {
			return err
		}
		if len(registries) == 0 {
			//revive:disable:error-strings This is a top-level error message displayed as is to the end user
			return fmt.Errorf(`Registry %q not found in the list of devfile registries.
Please use 'odo preference <add/remove> registry'' command to configure devfile registries or add an in-cluster registry (see https://devfile.io/docs/2.2.0/deploying-a-devfile-registry).`,
				registryName)
			//revive:enable:error-strings
		}
		for _, r := range registries {
			isGithubRegistry, err := registry.IsGithubBasedRegistry(r.URL)
			if err != nil {
				return err
			}
			if r.Name == registryName && isGithubRegistry {
				return &registry.ErrGithubRegistryNotSupported{}
			}
		}
	}

	if flags[FLAG_DEVFILE_PATH] != "" && registryName != "" {
		return errors.New("--devfile-registry parameter cannot be used with --devfile-path")
	}

	err := dfutil.ValidateK8sResourceName("name", flags[FLAG_NAME])
	if err != nil {
		return err
	}

	empty, err := location.DirIsEmpty(fs, dir)
	if err != nil {
		return err
	}
	if !empty && flags[FLAG_STARTER] != "" {
		return errors.New("--starter parameter cannot be used when the directory is not empty")
	}

	return nil
}

func (o *FlagsBackend) SelectDevfile(ctx context.Context, flags map[string]string, _ filesystem.Filesystem, _ string) (*api.DetectionResult, error) {
	return &api.DetectionResult{
		Devfile:         flags[FLAG_DEVFILE],
		DevfileRegistry: flags[FLAG_DEVFILE_REGISTRY],
		DevfilePath:     flags[FLAG_DEVFILE_PATH],
		DevfileVersion:  flags[FLAG_DEVFILE_VERSION],
	}, nil
}

func (o *FlagsBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	starter := flags[FLAG_STARTER]
	if starter == "" {
		return nil, nil
	}
	projects, err := devfile.Data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	var prj v1alpha2.StarterProject
	for _, prj = range projects {
		if prj.Name == starter {
			return &prj, nil
		}
	}
	return nil, fmt.Errorf("starter project %q not found in devfile", starter)
}

func (o *FlagsBackend) PersonalizeName(_ parser.DevfileObj, flags map[string]string) (string, error) {
	if validK8sNameErr := dfutil.ValidateK8sResourceName("name", flags[FLAG_NAME]); validK8sNameErr != nil {
		return "", validK8sNameErr
	}
	return flags[FLAG_NAME], nil

}

func (o FlagsBackend) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) (parser.DevfileObj, error) {
	return devfileobj, nil
}

func (o FlagsBackend) HandleApplicationPorts(devfileobj parser.DevfileObj, ports []int, flags map[string]string) (parser.DevfileObj, error) {
	// Currently not supported, but this will be done in a separate issue: https://github.com/redhat-developer/odo/issues/6211
	return devfileobj, nil
}
