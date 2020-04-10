package pipelines

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/cli/pipelines/environment"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended pipelines command name.
const RecommendedCommandName = "pipelines"

// NewCmdPipelines creates a new pipelines command
func NewCmdPipelines(name, fullName string) *cobra.Command {

	initCmd := NewCmdInit(InitRecommendedCommandName, odoutil.GetFullName(fullName, InitRecommendedCommandName))
	envCmd := environment.NewCmdEnv(environment.EnvRecommendedCommandName, odoutil.GetFullName(fullName, environment.EnvRecommendedCommandName))

	var pipelinesCmd = &cobra.Command{
		Use:   name,
		Short: "Manage pipelines",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, InitRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	pipelinesCmd.Flags().AddFlagSet(initCmd.Flags())
	pipelinesCmd.AddCommand(initCmd)
	pipelinesCmd.AddCommand(envCmd)

	pipelinesCmd.Annotations = map[string]string{"command": "main"}
	pipelinesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return pipelinesCmd
}
