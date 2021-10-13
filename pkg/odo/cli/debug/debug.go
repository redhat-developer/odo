package debug

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// RecommendedCommandName is the recommended debug command name
	RecommendedCommandName = "debug"
)

var debugLongDesc = ktemplates.LongDesc(`Debug allows you to remotely debug your application.`)

func NewCmdDebug(name, fullName string) *cobra.Command {

	portforwardCmd := NewCmdPortForward(portforwardCommandName, util.GetFullName(fullName, portforwardCommandName))
	infoCmd := NewCmdInfo(infoCommandName, util.GetFullName(fullName, infoCommandName))

	debugCmd := &cobra.Command{
		Use:   name,
		Short: "Debug commands",
		Example: fmt.Sprintf("%s\n\n%s",
			portforwardCmd.Example,
			infoCmd.Example),
		Long:    debugLongDesc,
		Aliases: []string{"d"},
	}

	debugCmd.SetUsageTemplate(util.CmdUsageTemplate)
	debugCmd.AddCommand(portforwardCmd)
	debugCmd.AddCommand(infoCmd)
	debugCmd.Annotations = map[string]string{"command": "main"}

	return debugCmd
}
