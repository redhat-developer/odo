package component

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ghodss/yaml"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all components in the current application",
	Long:  "List all components in the current application.",
	Example: `  # List all components in the application
  odo list
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		components, err := component.List(client, applicationName)
		odoutil.LogErrorAndExit(err, "")
		if len(components) == 0 {
			log.Errorf("There are no components deployed.")
			return
		}
		if outputFlag == "json" {
			//activeMark := " "
			//w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			//fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE")
			currentComponent := context.ComponentAllowingEmpty(true)
			var q []component.CompoList
			for _, comp := range components {
				if comp.ComponentName == currentComponent {
					//activeMark = "*"
					q = append(q, component.CompoList{Name: comp.ComponentName, Type: comp.ComponentImageType, Active: true})
				}
				//fmt.Fprintln(w, activeMark, "\t", comp.ComponentName, "\t", comp.ComponentImageType)
				//activeMark = " "
			}
			//w.Flush()
			////////////////////
			//var compo []component.Component
			//for _, i := range components {
			//	desc, _ := component.GetComponentDesc(client, i.ComponentName, applicationName)
			//	compo = append(compo, component.Component{TypeMeta: metav1.TypeMeta{Kind: "Component", APIVersion: util.APIVersion}, ObjectMeta: metav1.ObjectMeta{Name: i.ComponentName}, Spec: desc})
			//}
			p := component.ComponentList{
				TypeMeta: metav1.TypeMeta{
					Kind: "List", APIVersion: util.APIVersion,
				},
				ListMeta: metav1.ListMeta{},
				Items:    q,
			}
			out, err := yaml.Marshal(p)
			odoutil.LogErrorAndExit(err, "")
			fmt.Println(string(out))

		} else {

			activeMark := " "
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE")
			currentComponent := context.ComponentAllowingEmpty(true)
			for _, comp := range components {
				if comp.ComponentName == currentComponent {
					activeMark = "*"
				}
				fmt.Fprintln(w, activeMark, "\t", comp.ComponentName, "\t", comp.ComponentImageType)
				activeMark = " "
			}
			w.Flush()
		}

	},
}

// NewCmdList implements the list odo command
func NewCmdList() *cobra.Command {
	// Add a defined annotation in order to appear in the help menu
	componentListCmd.Annotations = map[string]string{"command": "component"}
	componentListCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "gives output in the form of json")
	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentListCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentListCmd)

	return componentListCmd
}
