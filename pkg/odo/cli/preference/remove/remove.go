package remove

import (
	"fmt"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
)

const (
	// RecommendedCommandName is the recommended registry command name
	RecommendedCommandName = "remove"
)

var registryDesc = ktemplates.LongDesc(`Remove value from an array of items`)

// NewCmdRemove implements the registry configuration command
func NewCmdRemove(name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	registryCmd := NewCmdRegistry(registryCommandName, util.GetFullName(fullName, registryCommandName), testClientset)

	removeCmd := &cobra.Command{
		Use:   name,
		Short: registryDesc,
		Long:  registryDesc,
		Example: fmt.Sprintf("%s\n",
			registryCmd.Example,
		),
	}

	removeCmd.AddCommand(registryCmd)
	removeCmd.SetUsageTemplate(util.CmdUsageTemplate)
	util.SetCommandGroup(removeCmd, util.MainGroup)

	return removeCmd
}
