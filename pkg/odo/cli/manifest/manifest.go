package manifest

import (
	"fmt"

	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended manifest command name.
const RecommendedCommandName = "manifest"

// NewCmdManifest implements the component odo command
func NewCmdManifest(name, fullName string) *cobra.Command {

	initCmd := NewCmdInit(InitRecommendedCommandName, odoutil.GetFullName(fullName, InitRecommendedCommandName))
	var manifestCmd = &cobra.Command{
		Use:   name,
		Short: "Manifest operations",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, InitRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	manifestCmd.Flags().AddFlagSet(initCmd.Flags())
	manifestCmd.AddCommand(initCmd)

	manifestCmd.Annotations = map[string]string{"command": "main"}
	manifestCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return manifestCmd
}
