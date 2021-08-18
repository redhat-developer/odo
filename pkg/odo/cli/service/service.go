package service

import (
	"fmt"

	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"

	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended service command name
const RecommendedCommandName = "service"

var serviceLongDesc = ktemplates.LongDesc(`Perform service related operations`)

// NewCmdService implements the odo service command
func NewCmdService(name, fullName string) *cobra.Command {
	serviceCreateCmd := NewCmdServiceCreate(createRecommendedCommandName, util.GetFullName(fullName, createRecommendedCommandName))
	serviceListCmd := NewCmdServiceList(listRecommendedCommandName, util.GetFullName(fullName, listRecommendedCommandName))
	serviceDeleteCmd := NewCmdServiceDelete(deleteRecommendedCommandName, util.GetFullName(fullName, deleteRecommendedCommandName))
	serviceCmd := &cobra.Command{
		Use:   name,
		Short: "Perform service related operations",
		Long:  serviceLongDesc,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			serviceCreateCmd.Example,
			serviceDeleteCmd.Example,
			serviceListCmd.Example),
		Args: cobra.RangeArgs(1, 3),
	}
	// Add a defined annotation in order to appear in the help menu
	serviceCmd.Annotations = map[string]string{"command": "main"}
	serviceCmd.SetUsageTemplate(util.CmdUsageTemplate)
	serviceCmd.AddCommand(serviceCreateCmd, serviceDeleteCmd, serviceListCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(serviceCreateCmd)
	projectCmd.AddProjectFlag(serviceDeleteCmd)
	projectCmd.AddProjectFlag(serviceListCmd)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(serviceCreateCmd)
	appCmd.AddApplicationFlag(serviceDeleteCmd)
	appCmd.AddApplicationFlag(serviceListCmd)

	return serviceCmd
}
