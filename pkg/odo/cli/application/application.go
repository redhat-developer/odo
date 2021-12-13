package application

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended app command name
const RecommendedCommandName = "app"

// NewCmdApplication implements the odo application command
func NewCmdApplication(name, fullName string) *cobra.Command {
	delete := NewCmdDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	describe := NewCmdDescribe(describeRecommendedCommandName, odoutil.GetFullName(fullName, describeRecommendedCommandName))
	list := NewCmdList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	applicationCmd := &cobra.Command{
		Use:   name,
		Short: "Perform application operations",
		Long:  `Performs application operations related to your project.`,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			delete.Example,
			describe.Example,
			list.Example),
		Aliases: []string{"application"},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	applicationCmd.AddCommand(delete, describe, list)

	// Add a defined annotation in order to appear in the help menu
	applicationCmd.Annotations = map[string]string{"command": "main"}
	applicationCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return applicationCmd
}

// AddApplicationFlag adds a `app` flag to the given cobra command
// Also adds a completion handler to the flag
func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(util.ApplicationFlagName, "", "Application, defaults to active application")
	completion.RegisterCommandFlagHandler(cmd, "app", completion.AppCompletionHandler)
}
