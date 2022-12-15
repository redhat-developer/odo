package add

import (
	"fmt"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/util"
)

const (
	// RecommendedCommandName is the recommended registry command name
	RecommendedCommandName = "add"
)

var registryDesc = ktemplates.LongDesc(`Add value to an array of items`)

// NewCmdAdd implements the registry configuration command
func NewCmdAdd(name, fullName string) *cobra.Command {
	registryCmd := NewCmdRegistry(registryCommandName, util.GetFullName(fullName, registryCommandName))

	addCmd := &cobra.Command{
		Use:   name,
		Short: registryDesc,
		Long:  registryDesc,
		Example: fmt.Sprintf("%s\n",
			registryCmd.Example,
		),
	}

	addCmd.AddCommand(registryCmd)
	addCmd.SetUsageTemplate(util.CmdUsageTemplate)
	util.SetCommandGroup(addCmd, util.MainGroup)

	return addCmd
}
