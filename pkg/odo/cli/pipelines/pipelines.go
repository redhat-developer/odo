package pipelines

import (
	"fmt"

	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"
)

// RecommendedPipelinesCommandName is the recommended pipelines command name.
const RecommendedCommandName = "pipelines"

// NewCmdComponent implements the component odo command
func NewCmdComponent(name, fullName string) *cobra.Command {
	bootstrapCmd := NewCmdBootstrap(BootstrapRecommendedCommandName, odoutil.GetFullName(fullName, BootstrapRecommendedCommandName))
	var pipelinesCmd = &cobra.Command{
		Use:   name,
		Short: "Manage pipelines",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, BootstrapRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	pipelinesCmd.Flags().AddFlagSet(bootstrapCmd.Flags())

	pipelinesCmd.AddCommand(bootstrapCmd)
	pipelinesCmd.Annotations = map[string]string{"command": "main"}
	pipelinesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return pipelinesCmd
}
