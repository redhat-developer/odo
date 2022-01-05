package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const describeRecommendedCommandName = "describe"

var (
	describeExample = ktemplates.Examples(`  # Describe 'webapp' application
  %[1]s webapp`)
)

// DescribeOptions encapsulates the options for the odo command
type DescribeOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	appClient application.Client

	// Parameters
	appName string
}

// NewDescribeOptions creates a new DescribeOptions instance
func NewDescribeOptions(appClient application.Client) *DescribeOptions {
	return &DescribeOptions{
		appClient: appClient,
	}
}

// Complete completes DescribeOptions after they've been created
func (o *DescribeOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	o.appName = o.GetApplication()
	if len(args) == 1 {
		o.appName = args[0]
	}
	return
}

// Validate validates the DescribeOptions based on completed values
func (o *DescribeOptions) Validate() (err error) {
	if o.Context.GetProject() == "" || o.appName == "" {
		return util.ThrowContextError()
	}

	exist, err := o.appClient.Exists(o.appName)
	if !exist {
		return fmt.Errorf("%s app does not exists", o.appName)
	}
	return err
}

// Run contains the logic for the odo command
func (o *DescribeOptions) Run() (err error) {
	if o.IsJSON() {
		appDef := o.appClient.GetMachineReadableFormat(o.appName, o.GetProject())
		machineoutput.OutputSuccess(appDef)
		return nil
	}

	componentList, err := o.appClient.ComponentList(o.appName)
	if err != nil {
		return err
	}

	if len(componentList) == 0 {
		fmt.Printf("Application %s has no components or services deployed.", o.appName)
		return
	}

	fmt.Printf("Application Name: %s has %v component(s):\n--------------------------------------\n",
		o.appName, len(componentList))
	for _, currentComponent := range componentList {
		err := util.PrintComponentInfo(o.KClient, currentComponent.Name, currentComponent, o.appName, o.GetProject())
		if err != nil {
			return err
		}
		fmt.Println("--------------------------------------")
	}

	return nil
}

// NewCmdDescribe implements the odo command.
func NewCmdDescribe(name, fullName string) *cobra.Command {
	// The error is not handled at this point, it will be handled during Context creation
	kubclient, _ := kclient.New()
	o := NewDescribeOptions(application.NewClient(kubclient))
	command := &cobra.Command{
		Use:         fmt.Sprintf("%s [application_name]", name),
		Short:       "Describe the given application",
		Long:        "Describe the given application",
		Example:     fmt.Sprintf(describeExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	completion.RegisterCommandHandler(command, completion.AppCompletionHandler)

	project.AddProjectFlag(command)
	return command
}
