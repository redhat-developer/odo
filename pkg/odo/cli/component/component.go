package component

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedComponentCommandName is the recommended component command name
const RecommendedCommandName = "component"

// ComponentOptions encapsulates basic component options
type ComponentOptions struct {
	componentName string
	*genericclioptions.Context
}

// Complete completes component options
func (co *ComponentOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	co.Context = genericclioptions.NewContext(cmd)

	// If no arguments have been passed, get the current component
	// else, use the first argument and check to see if it exists
	if len(args) == 0 {
		co.componentName = co.Context.Component()
	} else {
		co.componentName = co.Context.Component(args[0])
	}
	return
}

// NewCmdComponent implements the component odo command
func NewCmdComponent(name, fullName string) *cobra.Command {

	componentGetCmd := NewCmdGet(GetRecommendedCommandName, odoutil.GetFullName(fullName, GetRecommendedCommandName))
	componentSetCmd := NewCmdSet(SetRecommendedCommandName, odoutil.GetFullName(fullName, SetRecommendedCommandName))
	createCmd := NewCmdCreate(CreateRecommendedCommandName, odoutil.GetFullName(fullName, CreateRecommendedCommandName))
	deleteCmd := NewCmdDelete(DeleteRecommendedCommandName, odoutil.GetFullName(fullName, DeleteRecommendedCommandName))
	describeCmd := NewCmdDescribe(DescribeRecommendedCommandName, odoutil.GetFullName(fullName, DescribeRecommendedCommandName))
	linkCmd := NewCmdLink(LinkRecommendedCommandName, odoutil.GetFullName(fullName, LinkRecommendedCommandName))
	unlinkCmd := NewCmdUnlink(UnlinkRecommendedCommandName, odoutil.GetFullName(fullName, UnlinkRecommendedCommandName))
	listCmd := NewCmdList(ListRecommendedCommandName, odoutil.GetFullName(fullName, ListRecommendedCommandName))
	logCmd := NewCmdLog(LogRecommendedCommandName, odoutil.GetFullName(fullName, LogRecommendedCommandName))
	pushCmd := NewCmdPush(PushRecommendedCommandName, odoutil.GetFullName(fullName, PushRecommendedCommandName))
	updateCmd := NewCmdUpdate(UpdateRecommendedCommandName, odoutil.GetFullName(fullName, UpdateRecommendedCommandName))
	watchCmd := NewCmdWatch(WatchRecommendedCommandName, odoutil.GetFullName(fullName, WatchRecommendedCommandName))

	// componentCmd represents the component command
	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Components of an application",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples, e.g. %s %s -h",
			componentGetCmd.Example,
			componentSetCmd.Example,
			fullName, CreateRecommendedCommandName),
		// 'odo component' is the same as 'odo component get'
		// 'odo component <component_name>' is the same as 'odo component set <component_name>'
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 && args[0] != GetRecommendedCommandName && args[0] != SetRecommendedCommandName {
				componentSetCmd.Run(cmd, args)
			} else {
				componentGetCmd.Run(cmd, args)
			}
		},
	}

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd, componentSetCmd, createCmd, deleteCmd, describeCmd, linkCmd, unlinkCmd, listCmd, logCmd, pushCmd, updateCmd, watchCmd)

	// Add a defined annotation in order to appear in the help menu
	componentCmd.Annotations = map[string]string{"command": "component"}
	componentCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return componentCmd
}

// AddComponentFlag adds a `component` flag to the given cobra command
// Also adds a completion handler to the flag
func AddComponentFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ComponentFlagName, "", "Component, defaults to active component.")
	completion.RegisterCommandFlagHandler(cmd, "component", completion.ComponentNameCompletionHandler)
}

// printDeleteComponentInfo will print things which will be deleted
func printDeleteComponentInfo(client *occlient.Client, componentName string, appName string, projectName string) error {
	componentDesc, err := component.GetComponentDesc(client, componentName, appName, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get component description")
	}

	if len(componentDesc.URLs.Items) != 0 {
		log.Info("This component has following urls that will be deleted with component")
		for _, url := range componentDesc.URLs.Items {
			log.Info(" URL named ", url.GetName(), " with value ", url.Spec.URL)
		}
	}

	if len(componentDesc.LinkedServices) != 0 {
		log.Info("This component has following services linked to it, which will get unlinked")
		for _, linkedService := range componentDesc.LinkedServices {
			log.Info(" Service named ", linkedService)
		}
	}

	if len(componentDesc.Storage) != 0 {
		log.Info("This component has following storages which will be deleted with the component")
		for _, store := range componentDesc.Storage {
			log.Info("  Storage", store.Name, "of size", store.Size)
		}
	}
	return nil
}
