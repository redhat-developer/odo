package remove

import (
	"fmt"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/util"
)

const (
	// RecommendedCommandName is the recommended registry command name
	RecommendedCommandName = "remove"
)

var registryDesc = ktemplates.LongDesc(`Configure devfile registry`)

// NewCmdRemove implements the registry configuration command
func NewCmdRemove(name, fullName string) *cobra.Command {
	registryCmd := NewCmdRegistry(registryCommandName, util.GetFullName(fullName, registryCommandName))

	removeCmd := &cobra.Command{
		Use:   name,
		Short: registryDesc,
		Long:  registryDesc,
		Example: fmt.Sprintf("%s",
			registryCmd.Example,
		),
	}

	removeCmd.AddCommand(registryCmd)
	removeCmd.SetUsageTemplate(util.CmdUsageTemplate)
	removeCmd.Annotations = map[string]string{"command": "main"}

	return removeCmd
}
