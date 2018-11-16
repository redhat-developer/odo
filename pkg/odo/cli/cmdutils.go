package cli

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
)

func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(util.ProjectFlagName, "", "Project, defaults to active project")
	completion.RegisterCommandFlagHandler(cmd, "project", completion.ProjectNameCompletionHandler)
}

func AddComponentFlag(cmd *cobra.Command) {
	cmd.Flags().String(util.ComponentFlagName, "", "Component, defaults to active component.")
}

func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(util.ApplicationFlagName, "", "Application, defaults to active application")
}
