package project

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended service command name
const RecommendedCommandName = "project"

// NewCmdProject implements the project odo command
func NewCmdProject(name, fullName string) *cobra.Command {

	projectCreateCmd := NewCmdProjectCreate(createRecommendedCommandName, fullName+" "+createRecommendedCommandName)
	projectSetCmd := NewCmdProjectSet(setRecommendedCommandName, fullName+" "+setRecommendedCommandName)
	projectListCmd := NewCmdProjectList(listRecommendedCommandName, fullName+" "+listRecommendedCommandName)
	projectDeleteCmd := NewCmdProjectDelete(deleteRecommendedCommandName, fullName+" "+deleteRecommendedCommandName)
	projectGetCmd := NewCmdProjectGet(getRecommendedCommandName, fullName+" "+getRecommendedCommandName)

	projectCmd := &cobra.Command{
		Use:   name + " [options]",
		Short: "Perform project operations",
		Long:  "Perform project operations",
		Example: fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			projectSetCmd.Example,
			projectCreateCmd.Example,
			projectListCmd.Example,
			projectDeleteCmd.Example,
			projectGetCmd.Example),
		// 'odo project' is the same as 'odo project get'
		// 'odo project <project_name>' is the same as 'odo project set <project_name>'
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 && args[0] != "get" && args[0] != "set" {
				projectSetCmd.Run(cmd, args)
			} else {
				projectGetCmd.Run(cmd, args)
			}
		},
	}

	projectCmd.Flags().AddFlagSet(projectGetCmd.Flags())
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add a defined annotation in order to appear in the help menu
	projectCmd.Annotations = map[string]string{"command": "other"}
	projectCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(projectSetCmd, completion.ProjectNameCompletionHandler)
	completion.RegisterCommandHandler(projectDeleteCmd, completion.ProjectNameCompletionHandler)

	return projectCmd
}

// AddProjectFlag adds a `project` flag to the given cobra command
// Also adds a completion handler to the flag
func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ProjectFlagName, "", "Project, defaults to active project")
	completion.RegisterCommandFlagHandler(cmd, "project", completion.ProjectNameCompletionHandler)
}
