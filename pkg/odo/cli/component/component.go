package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended component command name
const RecommendedCommandName = "component"

// ComponentOptions encapsulates basic component options
type ComponentOptions struct {
	componentName string
	*genericclioptions.Context
}

// Complete completes component options
func (co *ComponentOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	co.Context, err = genericclioptions.New(genericclioptions.CreateParameters{Cmd: cmd})
	if err != nil {
		co.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
			Cmd:     cmd,
			Offline: true,
		})
		if err != nil {
			return err
		}
	}

	co.componentName, err = co.Context.Component(args...)
	if err != nil {
		return nil
	}
	return
}

// NewCmdComponent implements the component odo command
func NewCmdComponent(name, fullName string) *cobra.Command {

	componentGetCmd := NewCmdGet(GetRecommendedCommandName, odoutil.GetFullName(fullName, GetRecommendedCommandName))
	createCmd := NewCmdCreate(CreateRecommendedCommandName, odoutil.GetFullName(fullName, CreateRecommendedCommandName))
	deleteCmd := NewCmdDelete(DeleteRecommendedCommandName, odoutil.GetFullName(fullName, DeleteRecommendedCommandName))
	describeCmd := NewCmdDescribe(DescribeRecommendedCommandName, odoutil.GetFullName(fullName, DescribeRecommendedCommandName))
	linkCmd := NewCmdLink(LinkRecommendedCommandName, odoutil.GetFullName(fullName, LinkRecommendedCommandName))
	unlinkCmd := NewCmdUnlink(UnlinkRecommendedCommandName, odoutil.GetFullName(fullName, UnlinkRecommendedCommandName))
	listCmd := NewCmdList(ListRecommendedCommandName, odoutil.GetFullName(fullName, ListRecommendedCommandName))
	logCmd := NewCmdLog(LogRecommendedCommandName, odoutil.GetFullName(fullName, LogRecommendedCommandName))
	pushCmd := NewCmdPush(PushRecommendedCommandName, odoutil.GetFullName(fullName, PushRecommendedCommandName))
	watchCmd := NewCmdWatch(WatchRecommendedCommandName, odoutil.GetFullName(fullName, WatchRecommendedCommandName))
	testCmd := NewCmdTest(TestRecommendedCommandName, odoutil.GetFullName(fullName, TestRecommendedCommandName))
	execCmd := NewCmdExec(ExecRecommendedCommandName, odoutil.GetFullName(fullName, ExecRecommendedCommandName))
	statusCmd := NewCmdStatus(StatusRecommendedCommandName, odoutil.GetFullName(fullName, StatusRecommendedCommandName))

	// componentCmd represents the component command
	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Manage components",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, CreateRecommendedCommandName),
		// `odo component set/get` and `odo get/set` are respectively deprecated as per the new workflow
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd, createCmd, deleteCmd, describeCmd, linkCmd, unlinkCmd, listCmd, logCmd, pushCmd, watchCmd, execCmd)
	componentCmd.AddCommand(testCmd, statusCmd)

	// Add a defined annotation in order to appear in the help menu
	componentCmd.Annotations = map[string]string{"command": "main"}
	componentCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return componentCmd
}

// AddComponentFlag adds a `component` flag to the given cobra command
// Also adds a completion handler to the flag
func AddComponentFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ComponentFlagName, "", "Component, defaults to active component.")
	completion.RegisterCommandFlagHandler(cmd, "component", completion.ComponentNameCompletionHandler)
}
