package component

import (
	"fmt"
	"os"

	"encoding/json"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

// RecommendedListCommandName is the recommended watch command name
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
	if len(components) == 0 {
		log.Errorf("There are no components deployed.")
		return
	}
	if lo.outputFlag == "json" {
		var compList []component.Component
		for _, compo := range components {
			componentDesc, err := component.GetComponentDesc(lo.Client, compo.ComponentName, lo.Application, lo.Project)
			if err != nil {
				return err
			}
			compoDef := getMachineReadableFormat(componentDesc, lo.Application, lo.Project)
			compList = append(compList, compoDef)
		}

		compListDef := component.ComponentList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "List",
				APIVersion: "odo.openshift.io/v1alpha1",
			},
			ListMeta: metav1.ListMeta{},
			Items:    compList,
		}

		out, err := json.Marshal(compListDef)
		if err != nil {
			return err
		}
		fmt.Println(string(out))

	} else {

		activeMark := " "
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE")
		currentComponent := lo.Context.ComponentAllowingEmpty(true)
		for _, comp := range components {
			if comp.ComponentName == currentComponent {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", comp.ComponentName, "\t", comp.ComponentImageType)
			activeMark = " "
		}
		w.Flush()
	}
	return
}

// NewCmdList implements the list odo command
func NewCmdList(name, fullName string) *cobra.Command {
	lo := NewListOptions()

	var componentListCmd = &cobra.Command{
		Use:     name,
		Short:   "List all components in the current application",
		Long:    "List all components in the current application.",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(lo.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(lo.Validate(), "")
			odoutil.LogErrorAndExit(lo.Run(), "")
		},
	}
	// Add a defined annotation in order to appear in the help menu
	componentListCmd.Annotations = map[string]string{"command": "component"}

	componentListCmd.Flags().StringVarP(&lo.outputFlag, "output", "o", "", "output in json format")

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentListCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentListCmd)

	return componentListCmd
}
