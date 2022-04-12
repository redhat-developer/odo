package backend

import (
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const (
	FLAG_NAME             = "name"
	FLAG_DEVFILE          = "devfile"
	FLAG_DEVFILE_REGISTRY = "devfile-registry"
	FLAG_STARTER          = "starter"
	FLAG_DEVFILE_PATH     = "devfile-path"
)

// FlagsBackend is a backend that will extract all needed information from flags passed to the command
type FlagsBackend struct {
	preferenceClient preference.Client
}

func NewFlagsBackend(preferenceClient preference.Client) *FlagsBackend {
	return &FlagsBackend{
		preferenceClient: preferenceClient,
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

	if flags[FLAG_DEVFILE_REGISTRY] != "" && !o.preferenceClient.RegistryNameExists(flags[FLAG_DEVFILE_REGISTRY]) {
		return fmt.Errorf("registry %q not found in the list of devfile registries. Please use `odo registry` command to configure devfile registries", flags[FLAG_DEVFILE_REGISTRY])
	}

	if flags[FLAG_DEVFILE_PATH] != "" && flags[FLAG_DEVFILE_REGISTRY] != "" {
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

func (o *FlagsBackend) SelectDevfile(flags map[string]string, _ filesystem.Filesystem, _ string) (*alizer.DevfileLocation, error) {
	return &alizer.DevfileLocation{
		Devfile:         flags[FLAG_DEVFILE],
		DevfileRegistry: flags[FLAG_DEVFILE_REGISTRY],
		DevfilePath:     flags[FLAG_DEVFILE_PATH],
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

func (o *FlagsBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {
	return flags[FLAG_NAME], nil
}

func (o FlagsBackend) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) (parser.DevfileObj, error) {
	return devfileobj, nil
}
