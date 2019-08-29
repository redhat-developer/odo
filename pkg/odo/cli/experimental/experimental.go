package experimental

import (
	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended experimental command name
const RecommendedCommandName = "experimental"

var ExperimentalLongDesc = `Contains all the experimental commands which are not yet perfectly stable`

func NewCmdExperimental(name, fullName string) *cobra.Command {
	experimentalCmd := &cobra.Command{
		Use:     name,
		Short:   "Experimental commands",
		Long:    ExperimentalLongDesc,
		Aliases: []string{"e"},
	}

	experimentalCmd.SetUsageTemplate(util.CmdUsageTemplate)
	experimentalCmd.Annotations = map[string]string{"command": "main"}
	return experimentalCmd
}
