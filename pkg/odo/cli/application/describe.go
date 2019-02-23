package application

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const describeRecommendedCommandName = "describe"

var (
	describeExample = ktemplates.Examples(`  # Describe 'webapp' application,
  %[1]s webapp`)
)

// DescribeOptions encapsulates the options for the odo command
type DescribeOptions struct {
	appName      string
	outputFormat string
	*genericclioptions.Context
}

// NewDescribeOptions creates a new DescribeOptions instance
func NewDescribeOptions() *DescribeOptions {
	return &DescribeOptions{}
}

// Complete completes DescribeOptions after they've been created
func (o *DescribeOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.appName = o.Application
	if len(args) == 1 {
		o.appName = args[0]
	}
	return
}

// Validate validates the DescribeOptions based on completed values
func (o *DescribeOptions) Validate() (err error) {
	err = util.CheckOutputFlag(o.outputFormat)
	if err != nil {
		return err
	}
	if o.appName == "" {
		return fmt.Errorf("There's no active application in project: %v", o.Project)
	}

	return nil
}

// Run contains the logic for the odo command
func (o *DescribeOptions) Run() (err error) {
	if o.outputFormat == "json" {
		appDef := application.GetMachineReadableFormat(o.Client, o.appName, o.Project)
		out, err := json.Marshal(appDef)
		if err != nil {
			return err
		}
		fmt.Println(string(out))

	} else {
		// List of Component
		componentList, err := component.List(o.Client, o.appName)
		if err != nil {
			return err
		}

		//we ignore service errors here because it's entirely possible that the service catalog has not been installed
		serviceList, _ := service.ListWithDetailedStatus(o.Client, o.appName)

		if len(componentList.Items) == 0 && len(serviceList) == 0 {
			log.Errorf("Application %s has no components or services deployed.", o.appName)
		} else {
			fmt.Printf("Application Name: %s has %v component(s) and %v service(s):\n--------------------------------------\n",
				o.appName, len(componentList.Items), len(serviceList))
			if len(componentList.Items) > 0 {
				for _, currentComponent := range componentList.Items {
					componentDesc, err := component.GetComponent(o.Client, currentComponent.Name, o.appName, o.Project)
					util.LogErrorAndExit(err, "")
					util.PrintComponentInfo(o.Client, currentComponent.Name, componentDesc, o.Application)
					fmt.Println("--------------------------------------")
				}
			}
			if len(serviceList) > 0 {
				for _, currentService := range serviceList {
					fmt.Printf("Service Name: %s\n", currentService.Name)
					fmt.Printf("Type: %s\n", currentService.Type)
					fmt.Printf("Status: %s\n", currentService.Status)
					fmt.Println("--------------------------------------")
				}
			}
		}
	}

	return
}

// NewCmdDescribe implements the odo command.
func NewCmdDescribe(name, fullName string) *cobra.Command {
	o := NewDescribeOptions()
	command := &cobra.Command{
		Use:     fmt.Sprintf("%s [application_name]", name),
		Short:   "Describe the given application",
		Long:    "Describe the given application",
		Example: fmt.Sprintf(describeExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	command.Flags().StringVarP(&o.outputFormat, "output", "o", "", "output in json format")
	completion.RegisterCommandHandler(command, completion.AppCompletionHandler)
	project.AddProjectFlag(command)
	return command
}
