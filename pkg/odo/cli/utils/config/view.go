package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/config"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

const viewCommandName = "view"

func NewCmdView(name, fullName string) *cobra.Command {
	configurationViewCmd := &cobra.Command{
		Use:   viewCommandName,
		Short: "View current configuration values",
		Long:  "View current configuration values",
		Example: `  # For viewing the current configuration
   odo utils config view
  `,
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.New()
			if err != nil {
				odoutil.LogErrorAndExit(err, "unable to view configuration")
			}
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
			fmt.Fprintln(w, "UpdateNotification", "\t", cfg.GetUpdateNotification())
			fmt.Fprintln(w, "NamePrefix", "\t", cfg.GetNamePrefix())
			fmt.Fprintln(w, "Timeout", "\t", cfg.GetTimeout())
			w.Flush()
		},
	}

	return configurationViewCmd
}
