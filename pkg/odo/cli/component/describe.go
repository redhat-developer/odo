package component

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/component"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	"github.com/spf13/cobra"
)

// DescribeRecommendedCommandName is the recommended describe command name
const DescribeRecommendedCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe nodejs component,
%[1]s nodejs
`)

// DescribeOptions is a dummy container to attach complete, validate and run pattern
type DescribeOptions struct {
	outputFlag       string
	componentContext string
	*ComponentOptions
}

// NewDescribeOptions returns new instance of ListOptions
func NewDescribeOptions() *DescribeOptions {
	return &DescribeOptions{"", "", &ComponentOptions{}}
}

// Complete completes describe args
func (do *DescribeOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = do.ComponentOptions.Complete(name, cmd, args)
	return
}

// Validate validates the describe parameters
func (do *DescribeOptions) Validate() (err error) {
	isExists, err := component.Exists(do.Context.Client, do.componentName, do.Context.Application)
	if err != nil {
		return err
	}
	if !isExists {
		return fmt.Errorf("component %s does not exist", do.componentName)
	}

	return odoutil.CheckOutputFlag(do.outputFlag)
}

// Run has the logic to perform the required actions as part of command
func (do *DescribeOptions) Run() (err error) {
	componentDesc, err := component.GetComponent(do.Context.Client, do.componentName, do.Context.Application, do.Context.Project)
	if err != nil {
		return err
	}
	if do.outputFlag == "json" {
		out, err := json.Marshal(componentDesc)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {

		odoutil.PrintComponentInfo(do.Context.Client, do.componentName, componentDesc, do.Context.Application)
	}

	return
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(name, fullName string) *cobra.Command {
	do := NewDescribeOptions()

	var describeCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component_name]", name),
		Short:   "Describe the given component",
		Long:    `Describe the given component.`,
		Example: fmt.Sprintf(describeExample, fullName),
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(do, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)
	describeCmd.Flags().StringVarP(&do.outputFlag, "output", "o", "", "output in json format")
	// Adding --context flag
	genericclioptions.AddContextFlag(describeCmd, &do.componentContext)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	return describeCmd
}
