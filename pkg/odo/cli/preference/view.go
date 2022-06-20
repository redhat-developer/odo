package preference

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# View all set preference values 
   %[1]s
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	// Clients
	clientset *clientset.Clientset
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {
	return &ViewOptions{}
}

func (o *ViewOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
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
func (o *ViewOptions) Run(ctx context.Context) (err error) {
	preferenceList := o.clientset.PreferenceClient.NewPreferenceList()
	registryList := o.clientset.PreferenceClient.RegistryList()
	if registryList == nil || len(*registryList) == 0 {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return errors.New("No devfile registries added to the configuration. Refer `odo preference registry add -h` to add one")
		//revive:enable:error-strings
	}
	HumanReadableOutput(preferenceList, registryList)
	return
}

func HumanReadableOutput(preferenceList preference.PreferenceList, registryList *[]preference.Registry) {
	tStyle := table.Style{
		Box: table.BoxStyle{PaddingLeft: " ", PaddingRight: " "},
		Color: table.ColorOptions{
			Header: text.Colors{text.FgHiGreen},
		},
	}
	preferenceT := table.NewWriter()
	preferenceT.SetStyle(tStyle)
	preferenceT.SetOutputMirror(log.GetStdout())
	preferenceT.AppendHeader(table.Row{"PARAMETER", "VALUE"})
	preferenceT.SortBy([]table.SortBy{{Name: "PARAMETER", Mode: table.Asc}})
	for _, pref := range preferenceList.Items {
		value := showBlankIfNil(pref.Value)
		if reflect.DeepEqual(value, pref.Default) {
			value = fmt.Sprintf("%v (default)", value)
		}
		preferenceT.AppendRow(table.Row{pref.Name, value})
	}
	registryT := table.NewWriter()
	registryT.SetStyle(tStyle)
	registryT.SetOutputMirror(log.GetStdout())
	registryT.AppendHeader(table.Row{"NAME", "URL", "SECURE"})
	regList := *registryList
	// Loop backwards here to ensure the registry display order is correct (display latest newly added registry firstly)
	for i := len(regList) - 1; i >= 0; i-- {
		registry := regList[i]
		secure := "No"
		if registry.Secure {
			secure = "Yes"
		}
		registryT.AppendRow(table.Row{registry.Name, registry.URL, secure})
	}

	log.Info("Preference parameters:")
	preferenceT.Render()
	log.Info("\nDevfile registries:")
	registryT.Render()
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
		Use:     name,
		Short:   "View current preference values",
		Long:    "View current preference values",
		Example: fmt.Sprintf(fmt.Sprint("\n", viewExample), fullName),

		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(preferenceViewCmd, clientset.PREFERENCE)
	return preferenceViewCmd
}
