package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/storage"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
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
	createCmd := NewCmdCreate(CreateRecommendedCommandName, odoutil.GetFullName(fullName, CreateRecommendedCommandName))
	deleteCmd := NewCmdDelete(DeleteRecommendedCommandName, odoutil.GetFullName(fullName, DeleteRecommendedCommandName))
	describeCmd := NewCmdDescribe(DescribeRecommendedCommandName, odoutil.GetFullName(fullName, DescribeRecommendedCommandName))
	deployCmd := NewCmdDeploy(DeployRecommendedCommandName, odoutil.GetFullName(fullName, DeployRecommendedCommandName))
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
		Short: "Manage components",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, CreateRecommendedCommandName),
		// `odo component set/get` and `odo get/set` are respectively deprecated as per the new workflow
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd, createCmd, deleteCmd, describeCmd, deployCmd, linkCmd, unlinkCmd, listCmd, logCmd, pushCmd, updateCmd, watchCmd)

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

// printDeleteComponentInfo will print things which will be deleted
func printDeleteComponentInfo(client *occlient.Client, componentName string, appName string, projectName string) error {
	componentDesc, err := component.GetComponent(client, componentName, appName, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get component description")
	}

	if len(componentDesc.Spec.URL) != 0 {
		log.Info("This component has following urls that will be deleted with component")
		ul, err := url.ListPushed(client, componentDesc.Name, appName)
		if err != nil {
			return errors.Wrap(err, "Could not get url list")
		}
		for _, u := range ul.Items {
			log.Info("URL named", u.GetName(), "with host", u.Spec.Host, "having protocol", u.Spec.Protocol, "at port", u.Spec.Port)
		}
	}

	storages, err := storage.List(client, componentDesc.Name, appName)
	odoutil.LogErrorAndExit(err, "")
	if len(storages.Items) != 0 {
		log.Info("This component has following storages which will be deleted with the component")
		for _, storageName := range componentDesc.Spec.Storage {
			store := storages.Get(storageName)
			log.Info("Storage", store.GetName(), "of size", store.Spec.Size)
		}
	}
	return nil
}
