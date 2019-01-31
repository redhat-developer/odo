package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
)

const RecommendedCommandName = "config"

// NewCmdConfiguration implements the utils config odo command
func NewCmdConfiguration(name, fullName string) *cobra.Command {
	configurationViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	configurationSetCmd := NewCmdSet(setCommandName, util.GetFullName(fullName, setCommandName))
	configurationCmd := &cobra.Command{
		Use:   name,
		Short: "Modifies configuration settings",
		Long: `Modifies Odo specific configuration settings within the config file.

Available Parameters:
UpdateNotification - Controls if an update notification is shown or not (true or false)
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix
Timeout - Timeout (in seconds) for OpenShift server connection check`,
		Example: fmt.Sprintf("%s\n%s",
			configurationViewCmd.Example,
			configurationSetCmd.Example),
		Aliases: []string{"configuration"},
		// 'odo utils config' is the same as 'odo utils config --help'
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) >= 1 && args[0] != "view" && args[0] != "set" {
				log.Errorf(`Unknown command, use "set" or "view"`)
				os.Exit(1)
			}
			return nil
		}, Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 && args[0] == "set" {
				configurationSetCmd.Run(cmd, args)
			} else if len(args) > 0 && args[0] == "view" {
				configurationViewCmd.Run(cmd, args)
			} else {
				cmd.Help()
			}
		},
	}

	configurationCmd.AddCommand(configurationViewCmd, configurationSetCmd)

	configurationCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return configurationCmd
}
