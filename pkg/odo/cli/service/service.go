package service

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "service"

// NewCmdService implements the odo service command.
func NewCmdService(name string) *cobra.Command {
	serviceCreateCmd := NewCmdServiceCreate(createRecommendedCommandName)
	serviceListCmd := NewCmdServiceList(listRecommendedCommandName)
	serviceDeleteCmd := NewCmdServiceDelete(deleteRecommendedCommandName)
	serviceCmd := &cobra.Command{
		Use:   name,
		Short: "Perform service catalog operations",
		Long:  ` Perform service catalog operations, Limited to template service broker only.`,
		Example: fmt.Sprintf("%s\n%s\n%s",
			serviceCreateCmd.Example,
			serviceDeleteCmd.Example,
			serviceListCmd.Example),
		Args: cobra.RangeArgs(1, 3),
	}
	// Add a defined annotation in order to appear in the help menu
	serviceCmd.Annotations = map[string]string{"command": "other"}
	serviceCmd.SetUsageTemplate(util.CmdUsageTemplate)
	serviceCmd.AddCommand(serviceCreateCmd, serviceDeleteCmd, serviceListCmd)
	return serviceCmd
}

func addProjectFlag(cmd *cobra.Command) {
	genericclioptions.AddProjectFlag(cmd)
	completion.RegisterCommandFlagHandler(cmd, "project", completion.ProjectNameCompletionHandler)
}
