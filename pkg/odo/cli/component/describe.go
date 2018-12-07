package component

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/spf13/cobra"
)

var outputFlag string

var describeCmd = &cobra.Command{
	Use:   "describe [component_name]",
	Short: "Describe the given component",
	Long:  `Describe the given component.`,
	Example: `  # Describe nodejs component,
  odo describe nodejs
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		// If no arguments have been passed, get the current component
		// else, use the first argument and check to see if it exists
		var componentName string
		if len(args) == 0 {
			componentName = context.Component()
		} else {
			componentName = context.Component(args[0])
		}
		componentDesc, err := component.GetComponentDesc(client, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "")

		if outputFlag == "json" {
			spec := component.Component{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Component",
					APIVersion: util.APIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: componentName,
				},
				Spec: component.ComponentSpec{
					ComponentImageType: componentDesc.ComponentImageType,
					Path:               componentDesc.Path,
					URLs:               componentDesc.URLs,
					Env:                componentDesc.Env,
					Storage:            componentDesc.Storage,
				},
			}
			out, err := yaml.Marshal(spec)
			odoutil.LogErrorAndExit(err, "")
			fmt.Println(string(out))
		} else {

			odoutil.PrintComponentInfo(componentName, componentDesc)
		}
	},
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe() *cobra.Command {
	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	// -o flag to get machine readable output
	describeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "gives output in the form of json")

	return describeCmd
}
