package experimental

import (
	"github.com/openshift/odo/pkg/odo/cli/experimental/debug"
	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended experimental command name
const RecommendedCommandName = "experimental"

var ExperimentalLongDesc = `Contains all the experimental commands which are not yet perfectly stable`

func NewCmdExperimental(name, fullName string) *cobra.Command {

	debugCmd := debug.NewCmdDebug(debug.RecommendedCommandName, util.GetFullName(fullName, debug.RecommendedCommandName))
	experimentalCmd := &cobra.Command{
		Use:     name,
		Short:   "Experimental commands",
		Long:    ExperimentalLongDesc,
		Aliases: []string{"e"},
	}

	experimentalCmd.SetUsageTemplate(util.CmdUsageTemplate)
	experimentalCmd.AddCommand(debugCmd)
	experimentalCmd.Annotations = map[string]string{"command": "main"}
	return experimentalCmd
}
