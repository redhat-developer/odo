package preference

import (
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# For viewing the current preference value
   %[1]s
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct{}

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
func (o *ViewOptions) Run(cmd *cobra.Command) (err error) {

	cfg, err := preference.New()

	if err != nil {
		return err
	}

	if log.IsJSON() {
		prefDef := preference.NewPreferenceList(*cfg)
		machineoutput.OutputSuccess(prefDef)

		return
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
	fmt.Fprintln(w, "UpdateNotification", "\t", showBlankIfNil(cfg.OdoSettings.UpdateNotification))
	fmt.Fprintln(w, "NamePrefix", "\t", showBlankIfNil(cfg.OdoSettings.NamePrefix))
	fmt.Fprintln(w, "Timeout", "\t", showBlankIfNil(cfg.OdoSettings.Timeout))
	fmt.Fprintln(w, "BuildTimeout", "\t", showBlankIfNil(cfg.OdoSettings.BuildTimeout))
	fmt.Fprintln(w, "PushTimeout", "\t", showBlankIfNil(cfg.OdoSettings.PushTimeout))
	fmt.Fprintln(w, "Ephemeral", "\t", showBlankIfNil(cfg.OdoSettings.Ephemeral))
	fmt.Fprintln(w, "ConsentTelemetry", "\t", showBlankIfNil(cfg.OdoSettings.ConsentTelemetry))

	w.Flush()
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
	preferenceViewCmd := &cobra.Command{
		Use:         name,
		Short:       "View current preference values",
		Long:        "View current preference values",
		Example:     fmt.Sprintf(fmt.Sprint("\n", viewExample), fullName),
		Annotations: map[string]string{"machineoutput": "json"},

		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return preferenceViewCmd
}
