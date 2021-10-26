package registry

import (
	// Built-in packages
	"fmt"

	// Third-party packages
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/openshift/odo/v2/pkg/odo/util"
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
	registryUpdateCmd := NewCmdUpdate(updateCommandName, util.GetFullName(fullName, updateCommandName))
	registryDeleteCmd := NewCmdDelete(deleteCommandName, util.GetFullName(fullName, deleteCommandName))

	registryCmd := &cobra.Command{
		Use:   name,
		Short: registryDesc,
		Long:  registryDesc,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
			registryAddCmd.Example,
			registryListCmd.Example,
			registryUpdateCmd.Example,
			registryDeleteCmd.Example,
		),
	}

	registryCmd.AddCommand(registryAddCmd, registryListCmd, registryUpdateCmd, registryDeleteCmd)
	registryCmd.SetUsageTemplate(util.CmdUsageTemplate)
	registryCmd.Annotations = map[string]string{"command": "main"}

	return registryCmd
}
