package debug

import (
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended debug command name
const RecommendedCommandName = "debug"

var DebugLongDesc = `Debugging related functions`

func NewCmdDebug(name, fullName string) *cobra.Command {

	portforwardCmd := NewCmdPortForward(portforwardCommandName, util.GetFullName(fullName, portforwardCommandName))

	debugCmd := &cobra.Command{
		Use:     name,
		Short:   "Debug commands",
		Long:    DebugLongDesc,
		Aliases: []string{"d"},
	}

	debugCmd.SetUsageTemplate(util.CmdUsageTemplate)
	debugCmd.AddCommand(portforwardCmd)
	return debugCmd
}
