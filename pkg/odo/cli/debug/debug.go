package debug

import (
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended debug command name
const RecommendedCommandName = "debug"

var DebugLongDesc = `Warning - Debug is currently in tech preview and hence is subject to change in future.

Debug allows you to remotely debug your application`

func NewCmdDebug(name, fullName string) *cobra.Command {

	portforwardCmd := NewCmdPortForward(portforwardCommandName, util.GetFullName(fullName, portforwardCommandName))
	infoCmd := NewCmdInfo(infoCommandName, util.GetFullName(fullName, infoCommandName))

	debugCmd := &cobra.Command{
		Use:     name,
		Short:   "Debug commands",
		Long:    DebugLongDesc,
		Aliases: []string{"d"},
	}

	debugCmd.SetUsageTemplate(util.CmdUsageTemplate)
	debugCmd.AddCommand(portforwardCmd)
	debugCmd.AddCommand(infoCmd)
	debugCmd.Annotations = map[string]string{"command": "main"}

	return debugCmd
}
