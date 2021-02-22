package common

import (
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
	case dc.VscodeLaunch != nil:
		return dc.VscodeLaunch.Group
	case dc.VscodeTask != nil:
		return dc.VscodeTask.Group
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
