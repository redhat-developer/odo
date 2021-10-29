package list

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const componentsRecommendedCommandName = "components"

var componentsExample = `  # Get the supported components
  %[1]s`

// ListComponentsOptions encapsulates the options for the odo catalog list components command
type ListComponentsOptions struct {
	// list of known devfiles
	catalogDevfileList catalog.DevfileComponentTypeList
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewListComponentsOptions creates a new ListComponentsOptions instance
func NewListComponentsOptions() *ListComponentsOptions {
	return &ListComponentsOptions{}
}

// Complete completes ListComponentsOptions after they've been created
func (o *ListComponentsOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if err = util.CheckKubeConfigPath(); err == nil {
		o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{Cmd: cmd})
		if err != nil {
			return err
		}
	}

	o.catalogDevfileList, err = catalog.ListDevfileComponents("")
	if err != nil {
		return err
	}
	if o.catalogDevfileList.DevfileRegistries == nil {
		log.Warning("Please run 'odo registry add <registry name> <registry URL>' to add registry for listing devfile components\n")
	}

	return
}

// Validate validates the ListComponentsOptions based on completed values
func (o *ListComponentsOptions) Validate() (err error) {
	if len(o.catalogDevfileList.Items) == 0 {
		return fmt.Errorf("no deployable components found")
	}

	return err
}

type catalogList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Items             []catalog.DevfileComponentType `json:"items,omitempty"`
}

// Run contains the logic for the command associated with ListComponentsOptions
func (o *ListComponentsOptions) Run(cmd *cobra.Command) (err error) {
	if log.IsJSON() {
		combinedList := catalogList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "odo.dev/v1alpha1",
			},
			Items: o.catalogDevfileList.Items,
		}
		machineoutput.OutputSuccess(combinedList)
	} else {
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		if len(o.catalogDevfileList.Items) != 0 {
			fmt.Fprintln(w, "Odo Devfile Components:")
			fmt.Fprintln(w, "NAME", "\t", "DESCRIPTION", "\t", "REGISTRY")

			o.printDevfileCatalogList(w, o.catalogDevfileList.Items, "")
		}
		w.Flush()
	}
	return
}

// NewCmdCatalogListComponents implements the odo catalog list components command
func NewCmdCatalogListComponents(name, fullName string) *cobra.Command {
	o := NewListComponentsOptions()

	var componentListCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components",
		Long:        "List all available component types from OpenShift's Image Builder",
		Example:     fmt.Sprintf(componentsExample, fullName),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return componentListCmd
}

func (o *ListComponentsOptions) printDevfileCatalogList(w io.Writer, catalogDevfileList []catalog.DevfileComponentType, supported string) {
	for _, devfileComponent := range catalogDevfileList {
		if supported != "" {
			fmt.Fprintln(w, devfileComponent.Name, "\t", util.TruncateString(devfileComponent.Description, 60, "..."), "\t", devfileComponent.Registry.Name, "\t", supported)
		} else {
			fmt.Fprintln(w, devfileComponent.Name, "\t", util.TruncateString(devfileComponent.Description, 60, "..."), "\t", devfileComponent.Registry.Name)
		}
	}
}
