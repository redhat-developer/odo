package config

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"
)

const viewCommandName = "view"

var viewExample = ktemplates.Examples(`# For viewing the current configuration from devfile or local config file
   %[1]s
   
  `)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	contextDir string
	*genericclioptions.Context
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {
	return &ViewOptions{}
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	params := genericclioptions.NewCreateParameters(cmd).NeedDevfile(o.contextDir)
	o.Context, err = genericclioptions.New(params)
	return
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *ViewOptions) Run(cmd *cobra.Command) (err error) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 2, ' ', tabwriter.TabIndent)
	repr, err := component.ToDevfileRepresentation(o.Context.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return err
	}
	if log.IsJSON() {
		machineoutput.OutputSuccess(component.WrapFromJSONOutput(repr))
		return
	}
	representation, err := yaml.Marshal(repr)
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(representation))
	return err
}

// NewCmdView implements the config view odo command
func NewCmdView(name, fullName string) *cobra.Command {
	o := NewViewOptions()
	configurationViewCmd := &cobra.Command{
		Use:         name,
		Short:       "View current configuration values",
		Long:        "View current configuration values",
		Annotations: map[string]string{"machineoutput": "json"},
		Example:     fmt.Sprintf(fmt.Sprint("\n", viewExample), fullName),
		Args:        cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(configurationViewCmd, &o.contextDir)

	return configurationViewCmd
}
