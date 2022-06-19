package registry

import (
	// Built-in packages
	"fmt"

	// Third-party packages
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/redhat-developer/odo/pkg/odo/util"
)

const (
	// RecommendedCommandName is the recommended registry command name
	RecommendedCommandName = "registry"
)

var registryDesc = ktemplates.LongDesc(`Configure devfile registry`)

// NewCmdRegistry implements the registry configuration command
func NewCmdRegistry(name, fullName string) *cobra.Command {
	registryAddCmd := NewCmdAdd(addCommandName, util.GetFullName(fullName, addCommandName))
	registryListCmd := NewCmdList(listCommandName, util.GetFullName(fullName, listCommandName))
	registryDeleteCmd := NewCmdDelete(deleteCommandName, util.GetFullName(fullName, deleteCommandName))

	registryCmd := &cobra.Command{
		Use:   name,
		Short: registryDesc,
		Long:  registryDesc,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			registryAddCmd.Example,
			registryListCmd.Example,
			registryDeleteCmd.Example,
		),
	}

	registryCmd.AddCommand(registryAddCmd, registryListCmd, registryDeleteCmd)
	registryCmd.SetUsageTemplate(util.CmdUsageTemplate)
	registryCmd.Annotations = map[string]string{"command": "main"}

	return registryCmd
}
