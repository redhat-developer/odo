package project

import (
	"fmt"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended project command name
const RecommendedCommandName = "project"

// NewCmdProject implements the project odo command
func NewCmdProject(name, fullName string) *cobra.Command {

	projectCreateCmd := NewCmdProjectCreate(createRecommendedCommandName, odoutil.GetFullName(fullName, createRecommendedCommandName))
	projectSetCmd := NewCmdProjectSet(setRecommendedCommandName, odoutil.GetFullName(fullName, setRecommendedCommandName))
	projectListCmd := NewCmdProjectList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	projectDeleteCmd := NewCmdProjectDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	projectGetCmd := NewCmdProjectGet(getRecommendedCommandName, odoutil.GetFullName(fullName, getRecommendedCommandName))

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
			if len(args) > 0 && args[0] != getRecommendedCommandName && args[0] != setRecommendedCommandName {
				projectSetCmd.Run(cmd, args)
			} else {
				projectGetCmd.Run(cmd, args)
			}
		},
	}

	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add a defined annotation in order to appear in the help menu
	projectCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(projectSetCmd, completion.ProjectNameCompletionHandler)
	completion.RegisterCommandHandler(projectDeleteCmd, completion.ProjectNameCompletionHandler)

	return projectCmd
}
