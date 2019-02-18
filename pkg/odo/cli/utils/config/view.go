package config

import (
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# For viewing the current local configuration
   %[1]s

   For viewing the current global configuration
   %[1]s --global
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	configGlobalFlag bool
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
	if o.configGlobalFlag {

		cfg, err := config.NewGlobalConfig()

		if err != nil {
			util.LogErrorAndExit(err, "")
		}
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
		fmt.Fprintln(w, "UpdateNotification", "\t", showBlankIfNil(cfg.OdoSettings.UpdateNotification))
		fmt.Fprintln(w, "NamePrefix", "\t", showBlankIfNil(cfg.OdoSettings.NamePrefix))
		fmt.Fprintln(w, "Timeout", "\t", showBlankIfNil(cfg.OdoSettings.Timeout))
		w.Flush()
	} else {
		cfg, err := config.NewLocalConfig()
		if err != nil {
			util.LogErrorAndExit(err, "")
		}
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
		fmt.Fprintln(w, "ComponentType", "\t", showBlankIfNil(cfg.ComponentSettings.ComponentType))
		fmt.Fprintln(w, "ComponentName", "\t", showBlankIfNil(cfg.ComponentSettings.ComponentName))
		fmt.Fprintln(w, "MinMemory", "\t", showBlankIfNil(cfg.ComponentSettings.MinMemory))
		fmt.Fprintln(w, "MaxMemory", "\t", showBlankIfNil(cfg.ComponentSettings.MaxMemory))
		fmt.Fprintln(w, "Ignore", "\t", showBlankIfNil(cfg.ComponentSettings.Ignore))
		fmt.Fprintln(w, "MinCPU", "\t", showBlankIfNil(cfg.ComponentSettings.MinCPU))
		fmt.Fprintln(w, "MaxCPU", "\t", showBlankIfNil(cfg.ComponentSettings.MaxCPU))
		w.Flush()
	}
	return
}

func showBlankIfNil(intf interface{}) interface{} {
	imm := reflect.ValueOf(intf)

	// if the value is nil then we should return a blank string
	if imm.IsNil() {
		return ""
	}

	// if its a pointer then we should de-ref it because we cant de-ref an interface{}
	if imm.Kind() == reflect.Ptr {
		return imm.Elem().Interface()
	}

	return intf
}

// NewCmdView implements the config view odo command
func NewCmdView(name, fullName string) *cobra.Command {
	o := NewViewOptions()
	configurationViewCmd := &cobra.Command{
		Use:     name,
		Short:   "View current configuration values",
		Long:    "View current configuration values",
		Example: fmt.Sprintf(fmt.Sprint("\n", viewExample), fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	configurationViewCmd.Flags().BoolVarP(&o.configGlobalFlag, "global", "g", false, "Use the global config file")
	return configurationViewCmd
}
