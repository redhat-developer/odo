package backend

import (
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
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

func (o *FlagsBackend) Validate(flags map[string]string) error {
	if len(flags) == 0 {
		return nil
	}

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

	err := util.ValidateK8sResourceName("name", flags[FLAG_NAME])
	if err != nil {
		return err
	}
	return nil
}

func (o *FlagsBackend) SelectDevfile(flags map[string]string) (bool, *DevfileLocation, error) {
	if len(flags) == 0 {
		return false, nil, nil
	}
	return true, &DevfileLocation{
		Devfile:         flags[FLAG_DEVFILE],
		DevfileRegistry: flags[FLAG_DEVFILE_REGISTRY],
		DevfilePath:     flags[FLAG_DEVFILE_PATH],
	}, nil
}

func (o *FlagsBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (bool, *v1alpha2.StarterProject, error) {
	if len(flags) == 0 {
		return false, nil, nil
	}
	starter := flags[FLAG_STARTER]
	projects, err := devfile.Data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return true, nil, err
	}
	var prj v1alpha2.StarterProject
	for _, prj = range projects {
		if prj.Name == starter {
			return true, &prj, nil
		}
	}
	return true, nil, fmt.Errorf("starter project %q not found in devfile", starter)
}

func (o *FlagsBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (bool, error) {
	if len(flags) == 0 {
		return false, nil
	}
	return true, devfile.SetMetadataName(flags[FLAG_NAME])
}
