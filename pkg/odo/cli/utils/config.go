package utils

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/log"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// configurationCmd represents the app command
var configurationCmd = &cobra.Command{
	Use:   "config",
	Short: "Modifies configuration settings",
	Long: `Modifies Odo specific configuration settings within the config file.

Available Parameters:
UpdateNotification - Controls if an update notification is shown or not (true or false)
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix
Timeout - Timeout (in seconds) for OpenShift server connection check`,
	Example: fmt.Sprintf("%s\n%s\n",
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

var configurationSetCmd = &cobra.Command{
	Use:   "set",
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
   odo utils config set timeout 20
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

var configurationViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration values",
	Long:  "View current configuration values",
	Example: `  # For viewing the current configuration
   odo utils config view
  `,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New()
		if err != nil {
			odoutil.CheckError(err, "unable to view configuration")
		}
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
		fmt.Fprintln(w, "UpdateNotification", "\t", cfg.GetUpdateNotification())
		fmt.Fprintln(w, "NamePrefix", "\t", cfg.GetNamePrefix())
		fmt.Fprintln(w, "Timeout", "\t", cfg.GetTimeout())
		w.Flush()
	},
}
