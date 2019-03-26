package component

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// ListRecommendedCommandName is the recommended watch command name
const ListRecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions is a dummy container to attach complete, validate and run pattern
type ListOptions struct {
	outputFlag string
	*genericclioptions.Context
}

// NewListOptions returns new instance of ListOptions
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes log args
func (lo *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	lo.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the list parameters
func (lo *ListOptions) Validate() (err error) {
	return odoutil.CheckOutputFlag(lo.outputFlag)
}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run() (err error) {
	components, err := component.List(lo.Client, lo.Application)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch components list")
	}
	glog.V(4).Infof("the components are %+v", components)

	if lo.outputFlag == "json" {

		out, err := json.Marshal(components)
		if err != nil {
			return err
		}
		fmt.Println(string(out))

	} else {
		if len(components.Items) == 0 {
			log.Errorf("There are no components deployed.")
			return
		}
		activeMark := " "
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE")
		currentComponent := lo.Context.ComponentAllowingEmpty(true)
		for _, comp := range components.Items {
			if comp.Name == currentComponent {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", comp.Name, "\t", comp.Spec.Type)
			activeMark = " "
		}
		w.Flush()
	}
	return
}

// NewCmdList implements the list odo command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var componentListCmd = &cobra.Command{
		Use:     name,
		Short:   "List all components in the current application",
		Long:    "List all components in the current application.",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	// Add a defined annotation in order to appear in the help menu
	componentListCmd.Annotations = map[string]string{"command": "component"}

	componentListCmd.Flags().StringVarP(&o.outputFlag, "output", "o", "", "output in json format")

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentListCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentListCmd)

	return componentListCmd
}
