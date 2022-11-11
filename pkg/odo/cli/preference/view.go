package preference

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
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

var _ genericclioptions.Runnable = (*ViewOptions)(nil)

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {
	return &ViewOptions{}
}

func (o *ViewOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	return
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate(ctx context.Context) (err error) {
	return
}

// Run contains the logic for the command
func (o *ViewOptions) Run(ctx context.Context) (err error) {
	preferenceList := o.clientset.PreferenceClient.NewPreferenceList()
	registryList := o.clientset.PreferenceClient.RegistryList()
	HumanReadableOutput(preferenceList, registryList)
	return
}

func (o *ViewOptions) RunForJsonOutput(ctx context.Context) (result interface{}, err error) {
	preferenceList := o.clientset.PreferenceClient.NewPreferenceList()
	registryList := o.clientset.PreferenceClient.RegistryList()

	return api.PreferenceView{
		Preferences: preferenceList.Items,
		Registries:  registryList,
	}, nil
}

func HumanReadableOutput(preferenceList preference.PreferenceList, registryList []preference.Registry) {
	preferenceT := ui.NewTable()
	preferenceT.AppendHeader(table.Row{"PARAMETER", "VALUE"})
	preferenceT.SortBy([]table.SortBy{{Name: "PARAMETER", Mode: table.Asc}})
	for _, pref := range preferenceList.Items {
		value := showBlankIfNil(pref.Value)
		if reflect.DeepEqual(value, pref.Default) {
			value = fmt.Sprintf("%v (default)", value)
		}
		preferenceT.AppendRow(table.Row{pref.Name, value})
	}
	registryT := ui.NewTable()
	registryT.AppendHeader(table.Row{"NAME", "URL", "SECURE"})

	// Loop backwards here to ensure the registry display order is correct (display latest newly added registry firstly)
	for i := range registryList {
		registry := (registryList)[i]
		secure := "No"
		if registry.Secure {
			secure = "Yes"
		}
		registryT.AppendRow(table.Row{registry.Name, registry.URL, secure})
	}

	log.Info("Preference parameters:")
	preferenceT.Render()
	log.Info("\nDevfile registries:")
	if len(registryList) == 0 {
		log.Warning("No devfile registries added to the configuration. Refer to `odo preference add registry -h` to add one")
		return
	}
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
	commonflags.UseOutputFlag(preferenceViewCmd)
	return preferenceViewCmd
}
