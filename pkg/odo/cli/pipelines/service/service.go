package service

import (
	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"
)

// ServiceRecommendedCommandName is the recommended application command name.
const ServiceRecommendedCommandName = "service"

// NewCmdService implements the component odo command
func NewCmdService(name, fullName string) *cobra.Command {

	addServiceCmd := NewCmdAddService(AddServiceRecommendedCommandName, odoutil.GetFullName(fullName, AddServiceRecommendedCommandName))
	var serviceCmd = &cobra.Command{
		Use:   name,
		Short: "Add a new service to GitOps",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	serviceCmd.Flags().AddFlagSet(addServiceCmd.Flags())
	serviceCmd.AddCommand(addServiceCmd)

	serviceCmd.Annotations = map[string]string{"command": "main"}
	serviceCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return serviceCmd
}
