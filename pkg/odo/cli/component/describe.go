package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/component"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/spf13/cobra"
)

// RecommendedDescribeCommandName is the recommended describe command name
const RecommendedDescribeCommandName = "describe"

var describeExample = ktemplates.Examples(`  # Describe nodejs component,
%[1]s nodejs
`)

// DescribeOptions is a dummy container to attach complete, validate and run pattern
type DescribeOptions struct {
	componentName string
	*ComponentOptions
}

// NewDescribeOptions returns new instance of ListOptions
func NewDescribeOptions() *DescribeOptions {
	return &DescribeOptions{"", &ComponentOptions{}}
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
	return
}

// Run has the logic to perform the required actions as part of command
func (do *DescribeOptions) Run() (err error) {
	componentDesc, err := component.GetComponentDesc(do.Context.Client, do.componentName, do.Context.Application, do.Context.Project)
	if err != nil {
		return err
	}

	odoutil.PrintComponentInfo(do.componentName, componentDesc)

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
			odoutil.LogErrorAndExit(do.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(do.Validate(), "")
			odoutil.LogErrorAndExit(do.Run(), "")
		},
	}

	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(describeCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(describeCmd)

	return describeCmd
}
