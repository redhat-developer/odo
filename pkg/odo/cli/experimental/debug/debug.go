package debug

import (
	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended debug command name
const RecommendedCommandName = "debug"

var DebugLongDesc = `Debugging related functions`

func NewCmdDebug(name, fullName string) *cobra.Command {
	debugCmd := &cobra.Command{
		Use:     name,
		Short:   "Debug commands",
		Long:    DebugLongDesc,
		Aliases: []string{"e"},
	}

	debugCmd.SetUsageTemplate(util.CmdUsageTemplate)
	debugCmd.Annotations = map[string]string{"command": "main"}
	return debugCmd
}
