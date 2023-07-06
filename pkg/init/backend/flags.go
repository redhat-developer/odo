package backend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/libdevfile"
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
	FLAG_RUN_PORT         = "run-port"
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

func (o FlagsBackend) HandleApplicationPorts(devfileobj parser.DevfileObj, _ []int, flags map[string]string) (parser.DevfileObj, error) {
	d, err := setPortsForFlag(devfileobj, flags, FLAG_RUN_PORT)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	return d, nil
}

func setPortsForFlag(devfileobj parser.DevfileObj, flags map[string]string, flagName string) (parser.DevfileObj, error) {
	flagVal := flags[flagName]
	// Repeatable flags are formatted as "[val1,val2]"
	if !(strings.HasPrefix(flagVal, "[") && strings.HasSuffix(flagVal, "]")) {
		return devfileobj, nil
	}
	portsStr := flagVal[1 : len(flagVal)-1]

	var ports []int
	split := strings.Split(portsStr, ",")
	for _, s := range split {
		p, err := strconv.Atoi(s)
		if err != nil {
			return parser.DevfileObj{}, fmt.Errorf("invalid value for %s (%q): %w", flagName, s, err)
		}
		ports = append(ports, p)
	}

	var kind v1alpha2.CommandGroupKind
	switch flagName {
	case FLAG_RUN_PORT:
		kind = v1alpha2.RunCommandGroupKind
	default:
		return parser.DevfileObj{}, fmt.Errorf("unknown flag: %q", flagName)
	}

	cmd, ok, err := libdevfile.GetCommand(devfileobj, "", kind)
	if err != nil {
		return parser.DevfileObj{}, err
	}
	if !ok {
		klog.V(3).Infof("Specified %s flag will not be applied - no default (or single non-default) command found for kind %v", flagName, kind)
		return devfileobj, nil
	}
	// command must be an exec command to determine the right container component endpoints to update.
	cmdType, err := common.GetCommandType(cmd)
	if err != nil {
		return parser.DevfileObj{}, err
	}
	if cmdType != v1alpha2.ExecCommandType {
		return parser.DevfileObj{},
			fmt.Errorf("%v cannot be used with non-exec commands. Found out that command (id: %s) for kind %v is of type %q instead",
				flagName, cmd.Id, kind, cmdType)
	}

	cmp, ok, err := libdevfile.FindComponentByName(devfileobj.Data, cmd.Exec.Component)
	if err != nil {
		return parser.DevfileObj{}, err
	}
	if !ok {
		return parser.DevfileObj{}, fmt.Errorf("component not found in Devfile for exec command %q", cmd.Id)
	}
	cmpType, err := common.GetComponentType(cmp)
	if err != nil {
		return parser.DevfileObj{}, err
	}
	if cmpType != v1alpha2.ContainerComponentType {
		return parser.DevfileObj{},
			fmt.Errorf("%v cannot be used with non-container components. Found out that command (id: %s) for kind %v points to a compoenent of type %q instead",
				flagName, cmd.Id, kind, cmpType)
	}

	err = setPortsInContainerComponent(&devfileobj, &cmp, ports, false)
	if err != nil {
		return parser.DevfileObj{}, err
	}
	return devfileobj, nil
}
