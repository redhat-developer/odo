package util

import (
	"github.com/spf13/cobra"
)

const (
	// ProjectFlagName is the name of the flag allowing a user to specify which project to operate on
	ProjectFlagName = "project"
	// ComponentFlagName is the name of the flag allowing a user to specify which component to operate on
	ComponentFlagName = "component"
	// OutputFlagName is the name of the flag allowing user to specify output format
	OutputFlagName = "o"
	// ContextFlagName is the name of the flag allowing a user to specify the location of the component settings
	ContextFlagName = "context"
	// ComponentNameFlagName is the name of the flag allowing a user to specify which component to operate on
	ComponentNameFlagName = "name"
)

// AddContextFlag adds `context` flag to given cobra command
func AddContextFlag(cmd *cobra.Command, setValueTo *string) {
	helpMessage := "Use given context directory as a source for component settings"
	if setValueTo != nil {
		cmd.Flags().StringVar(setValueTo, ContextFlagName, "", helpMessage)
	} else {
		cmd.Flags().String(ContextFlagName, "", helpMessage)
	}
}
