package genericclioptions

import (
	"github.com/spf13/cobra"
)

const (
	// SkipConnectionCheckFlagName is the name of the global flag used to skip connection check in the client
	SkipConnectionCheckFlagName = "skip-connection-check"
	// ProjectFlagName is the name of the flag allowing a user to specify which project to operate on
	ProjectFlagName = "project"
	// ApplicationFlagName is the name of the flag allowing a user to specify which application to operate on
	ApplicationFlagName = "app"
	// ComponentFlagName is the name of the flag allowing a user to specify which component to operate on
	ComponentFlagName = "component"
)

func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(ProjectFlagName, "", "Project, defaults to active project")
}

func AddComponentFlag(cmd *cobra.Command) {
	cmd.Flags().String(ComponentFlagName, "", "Component, defaults to active component.")
}

func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(ApplicationFlagName, "", "Application, defaults to active application")
}
