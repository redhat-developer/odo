package preference

import (
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# For viewing the current preference value
   %[1]s
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	// Clients
	prefClient preference.Client
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions(prefClient preference.Client) *ViewOptions {
	return &ViewOptions{
		prefClient: prefClient,
	}
}

func (o *ViewOptions) SetClientset(clientset *clientset.Clientset) {
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	return
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *ViewOptions) Run() (err error) {

	if log.IsJSON() {
		prefDef := o.prefClient.NewPreferenceList()
		machineoutput.OutputSuccess(prefDef)

		return
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "PARAMETER", "\t", "CURRENT_VALUE")
	fmt.Fprintln(w, "UpdateNotification", "\t", showBlankIfNil(o.prefClient.UpdateNotification()))
	fmt.Fprintln(w, "NamePrefix", "\t", showBlankIfNil(o.prefClient.NamePrefix()))
	fmt.Fprintln(w, "Timeout", "\t", showBlankIfNil(o.prefClient.Timeout()))
	fmt.Fprintln(w, "BuildTimeout", "\t", showBlankIfNil(o.prefClient.BuildTimeout()))
	fmt.Fprintln(w, "PushTimeout", "\t", showBlankIfNil(o.prefClient.PushTimeout()))
	fmt.Fprintln(w, "Ephemeral", "\t", showBlankIfNil(o.prefClient.EphemeralSourceVolume()))
	fmt.Fprintln(w, "ConsentTelemetry", "\t", showBlankIfNil(o.prefClient.ConsentTelemetry()))

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
	prefClient, err := preference.NewClient()
	if err != nil {
		util.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}
	o := NewViewOptions(prefClient)
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
