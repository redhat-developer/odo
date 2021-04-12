package common

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetGroup returns the group the command belongs to
func GetGroup(dc v1.Command) *v1.CommandGroup {
	switch {
	case dc.Composite != nil:
		return dc.Composite.Group
	case dc.Exec != nil:
		return dc.Exec.Group
	case dc.Apply != nil:
		return dc.Apply.Group
	case dc.Custom != nil:
		return dc.Custom.Group

	default:
		return nil
	}
}

// GetExecComponent returns the component of the exec command
func GetExecComponent(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.Component
	}

	return ""
}

// GetExecCommandLine returns the command line of the exec command
func GetExecCommandLine(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.CommandLine
	}

	return ""
}

// GetExecWorkingDir returns the working dir of the exec command
func GetExecWorkingDir(dc v1.Command) string {
	if dc.Exec != nil {
		return dc.Exec.WorkingDir
	}

	return ""
}

// GetCommandType returns the command type of a given command
func GetCommandType(command v1.Command) (v1.CommandType, error) {
	switch {
	case command.Apply != nil:
		return v1.ApplyCommandType, nil
	case command.Composite != nil:
		return v1.CompositeCommandType, nil
	case command.Exec != nil:
		return v1.ExecCommandType, nil
	case command.Custom != nil:
		return v1.CustomCommandType, nil

	default:
		return "", fmt.Errorf("unknown command type")
	}
}
