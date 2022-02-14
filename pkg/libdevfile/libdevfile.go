package libdevfile

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

type Handler interface {
	ApplyImage(image v1alpha2.Component) error
	ApplyKubernetes(kubernetes v1alpha2.Component) error
}

// Deploy executes the default Deploy command of the devfile
func Deploy(devfileObj parser.DevfileObj, handler Handler) error {
	devfileCommand, err := getDefaultCommand(devfileObj, v1alpha2.DeployCommandGroupKind)
	if err != nil {
		return err
	}

	return executeCommand(devfileObj, devfileCommand, handler)
}

// getDefaultCommand returns the default command of the given kind in the devfile.
// If only one command of the kind exists, it is returned, even if it is not marked as default
func getDefaultCommand(devfileObj parser.DevfileObj, kind v1alpha2.CommandGroupKind) (v1alpha2.Command, error) {
	groupCmds, err := devfileObj.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: kind,
		},
	})
	if err != nil {
		return v1alpha2.Command{}, err
	}
	if len(groupCmds) == 0 {
		return v1alpha2.Command{}, NewNoCommandFoundError(kind)
	}
	if len(groupCmds) > 1 {
		var found bool
		var foundGroupCmd v1alpha2.Command
		for _, groupCmd := range groupCmds {
			group := common.GetGroup(groupCmd)
			if group == nil {
				continue
			}
			if group.IsDefault != nil && *group.IsDefault {
				if found {
					return v1alpha2.Command{}, NewMoreThanOneDefaultCommandFoundError(kind)
				}
				found = true
				foundGroupCmd = groupCmd
			}
		}
		if !found {
			return v1alpha2.Command{}, NewNoDefaultCommandFoundError(kind)
		}
		return foundGroupCmd, nil
	}
	return groupCmds[0], nil
}

// executeCommand executes a specific command of a devfile
func executeCommand(devfileObj parser.DevfileObj, command v1alpha2.Command, handler Handler) error {
	cmd, err := newCommand(devfileObj, command)
	if err != nil {
		return err
	}
	return cmd.Execute(handler)
}
