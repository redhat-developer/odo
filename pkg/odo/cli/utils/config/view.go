package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"text/tabwriter"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`  # For viewing the current configuration
   %[1]s
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {
	return &ViewOptions{}
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	return
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *ViewOptions) Run() (err error) {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
	fmt.Fprintln(w, config.UpdateNotificationSetting, "\t", cfg.GetUpdateNotification())
	fmt.Fprintln(w, config.NamePrefixSetting, "\t", cfg.GetNamePrefix())
	fmt.Fprintln(w, config.TimeoutSetting, "\t", cfg.GetTimeout())
	w.Flush()
	return
}

func NewCmdView(name, fullName string) *cobra.Command {
	o := NewViewOptions()
	configurationViewCmd := &cobra.Command{
		Use:     name,
		Short:   "View current configuration values",
		Long:    "View current configuration values",
		Example: fmt.Sprintf(viewExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}

	return configurationViewCmd
}
