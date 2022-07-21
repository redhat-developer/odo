package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cli/create/namespace"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended namespace command name
const RecommendedCommandName = "create"

// NewCmdCreate implements the namespace odo command
func NewCmdCreate(name, fullName string) *cobra.Command {

	namespaceCreateCmd := namespace.NewCmdNamespaceCreate(namespace.RecommendedCommandName, odoutil.GetFullName(fullName, namespace.RecommendedCommandName))
	createCmd := &cobra.Command{
		Use:   name + " [options]",
		Short: "Perform create operation",
		Long:  "Perform create operation",
		Example: fmt.Sprintf("%s\n",
			namespaceCreateCmd.Example,
		),
		Annotations: map[string]string{"command": "management"},
	}

	createCmd.AddCommand(namespaceCreateCmd)

	// Add a defined annotation in order to appear in the help menu
	createCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return createCmd
}
