package add

import (
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cli/add/binding"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended add command name
const RecommendedCommandName = "add"

// NewCmdAdd implements the odo add command
func NewCmdAdd(name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   name,
		Short: "Add resources to devfile",
	}

	bindingCmd := binding.NewCmdBinding(binding.BindingRecommendedCommandName, util.GetFullName(fullName, binding.BindingRecommendedCommandName), testClientset)
	createCmd.AddCommand(bindingCmd)
	util.SetCommandGroup(createCmd, util.ManagementGroup)
	createCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return createCmd
}
