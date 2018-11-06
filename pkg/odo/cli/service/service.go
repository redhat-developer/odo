package service

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// RecommendedCommandName is the recommended service command name
const RecommendedCommandName = "service"

var serviceLongDesc = ktemplates.LongDesc(`
Perform service catalog operations, limited to template service broker and OpenShift Ansible Broker only.`)

// NewCmdService implements the odo service command
func NewCmdService(name, fullName string) *cobra.Command {
	serviceCreateCmd := NewCmdServiceCreate(createRecommendedCommandName, fullName+" "+createRecommendedCommandName)
	serviceListCmd := NewCmdServiceList(listRecommendedCommandName, fullName+" "+listRecommendedCommandName)
	serviceDeleteCmd := NewCmdServiceDelete(deleteRecommendedCommandName, fullName+" "+deleteRecommendedCommandName)
	serviceCmd := &cobra.Command{
		Use:   name,
		Short: "Perform service catalog operations",
		Long:  serviceLongDesc,
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

	//Adding `--project` flag
	completion.AddProjectFlag(serviceCreateCmd)
	completion.AddProjectFlag(serviceDeleteCmd)
	completion.AddProjectFlag(serviceListCmd)

	//Adding `--application` flag
	completion.AddApplicationFlag(serviceCreateCmd)
	completion.AddApplicationFlag(serviceDeleteCmd)
	completion.AddApplicationFlag(serviceListCmd)

	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	completion.RegisterCommandHandler(serviceDeleteCmd, completion.ServiceCompletionHandler)

	return serviceCmd
}
