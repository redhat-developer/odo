package manifest

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/odo/cli/manifest/environment"
	odoutil "github.com/openshift/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended manifest command name.
const RecommendedCommandName = "manifest"

// NewCmdManifest implements the component odo command
func NewCmdManifest(name, fullName string) *cobra.Command {
	var manifestCmd = &cobra.Command{
		Use:   name,
		Short: "Manifest operations",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, InitRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	initCmd := NewCmdInit(InitRecommendedCommandName, odoutil.GetFullName(fullName, InitRecommendedCommandName))
	bootstrapCmd := NewCmdBootstrap(BootstrapRecommendedCommandName, odoutil.GetFullName(fullName, BootstrapRecommendedCommandName))
	envCmd := environment.NewCmdEnv(environment.EnvRecommendedCommandName, odoutil.GetFullName(fullName, environment.EnvRecommendedCommandName))

	manifestCmd.AddCommand(initCmd)
	manifestCmd.AddCommand(bootstrapCmd)
	manifestCmd.AddCommand(envCmd)

	buildCmd := NewCmdBuild(BuildRecommendedCommandName, odoutil.GetFullName(fullName, BuildRecommendedCommandName))
	manifestCmd.AddCommand(buildCmd)

	manifestCmd.Annotations = map[string]string{"command": "main"}
	manifestCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return manifestCmd
}
