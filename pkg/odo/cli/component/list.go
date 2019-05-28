package component

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// ListRecommendedCommandName is the recommended watch command name
const ListRecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions is a dummy container to attach complete, validate and run pattern
type ListOptions struct {
	outputFlag       string
	pathFlag         string
	componentContext string
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
	if lo.Context.Project == "" || lo.Application == "" {
		return odoutil.ThrowContextError()
	}

	return odoutil.CheckOutputFlag(lo.outputFlag)
}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run() (err error) {
	var ap []component.Component
	if lo.pathFlag != "" {
		err := filepath.Walk(lo.pathFlag, func(path string, f os.FileInfo, err error) error {
			if strings.Contains(f.Name(), ".odo") {
				data, err := config.NewLocalConfigInfo(filepath.Dir(path))
				if err != nil {
					return err
				}
				exist, err := component.Exists(lo.Context.Client, data.GetName(), data.GetApplication())
				if err != nil {
					return err
				}
				// context will be filepath.Dir(path)
				con, _ := filepath.Abs(filepath.Dir(path))
				a := component.Component{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Component",
						APIVersion: "odo.openshift.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: data.GetName(),
					},
					Spec: component.ComponentSpec{
						Source: data.GetSourceLocation(),
						Type:   data.GetType(),
					},
					Status: component.ComponentStatus{
						Context: con,
						State:   exist,
					},
				}
				ap = append(ap, a)

			}
			return nil
		})

		if err != nil {
			return err
		}
		activeMark := " "
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE", "\t", "SOURCE", "\t", "STATE", "\t", "CONTEXT")
		currentComponent := lo.Context.ComponentAllowingEmpty(true)
		for _, file := range ap {
			d := "not Deployed"
			if file.Status.State {
				d = "Deployed"
			}
			if file.Name == currentComponent {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", file.Name, "\t", file.Spec.Type, "\t", file.Spec.Source, "\t", d, "\t", file.Status.Context)
			activeMark = " "

		}
		w.Flush()
		return nil
	}

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
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE", "\t", "SOURCE", "\t", "STATE")
		currentComponent := lo.Context.ComponentAllowingEmpty(true)
		for _, comp := range components.Items {
			if comp.Name == currentComponent {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", comp.Name, "\t", comp.Spec.Type, "\t", comp.Spec.Source, "\t", "Deployed")
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
	genericclioptions.AddContextFlag(componentListCmd, &o.componentContext)
	componentListCmd.Flags().StringVarP(&o.outputFlag, "output", "o", "", "output in json format")
	componentListCmd.Flags().StringVar(&o.pathFlag, "path", "", "path")
	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentListCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentListCmd)

	completion.RegisterCommandFlagHandler(componentListCmd, "path", completion.FileCompletionHandler)

	return componentListCmd
}
