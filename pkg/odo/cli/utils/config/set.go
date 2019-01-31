package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/spf13/cobra"
	"strings"
)

const setCommandName = "set"

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {

	configurationSetCmd := &cobra.Command{
		Use:   name,
		Short: "Set a value in odo config file",
		Long: `Set an individual value in the Odo configuration file

Available Parameters:
UpdateNotification - Controls if an update notification is shown or not (true or false)
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix.
Timeout - Timeout (in seconds) for OpenShift server connection check`,
		Example: `
   # Set a configuration value
   odo utils config set UpdateNotification false
   odo utils config set NamePrefix "app"
   odo utils config set Timeout 20
	`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("Please provide a parameter name and value")
			} else if len(args) > 2 {
				return fmt.Errorf("Only one value per parameter is allowed")
			} else {
				return nil
			}

		}, RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New()
			if err != nil {
				return errors.Wrapf(err, "unable to set configuration")
			}
			return cfg.SetConfiguration(strings.ToLower(args[0]), args[1])
		},
	}

	return configurationSetCmd
}
