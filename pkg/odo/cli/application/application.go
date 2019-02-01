package application

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var (
	applicationShortFlag       bool
	applicationForceDeleteFlag bool
)

// Description holds all information about application
type Description struct {
	Name       string `json:"applicationName,omitempty"`
	Components []component.Description
}

const RecommendedCommandName = "app"

// NewCmdApplication implements the odo application command
func NewCmdApplication(name, fullName string) *cobra.Command {
	create := NewCmdCreate(createRecommendedCommandName, odoutil.GetFullName(fullName, createRecommendedCommandName))
	delete := NewCmdDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	describe := NewCmdDescribe(describeRecommendedCommandName, odoutil.GetFullName(fullName, describeRecommendedCommandName))
	get := NewCmdGet(getRecommendedCommandName, odoutil.GetFullName(fullName, getRecommendedCommandName))
	list := NewCmdList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	set := NewCmdSet(setRecommendedCommandName, odoutil.GetFullName(fullName, setRecommendedCommandName))
	applicationCmd := &cobra.Command{
		Use:   name,
		Short: "Perform application operations",
		Long:  `Performs application operations related to your OpenShift project.`,
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
			create.Example,
			get.Example,
			delete.Example,
			describe.Example,
			list.Example,
			set.Example),
		Aliases: []string{"application"},
		Run: func(cmd *cobra.Command, args []string) {
			// 'odo app' is the same as 'odo app get'
			// 'odo app <application_name>' is the same as 'odo app set <application_name>'
			if len(args) == 1 && args[0] != getRecommendedCommandName && args[0] != setRecommendedCommandName {
				set.Run(cmd, args)
			} else {
				get.Run(cmd, args)
			}
		},
	}

	// add flags from 'get' to application command
	applicationCmd.Flags().AddFlagSet(get.Flags())

	applicationCmd.AddCommand(create, delete, describe, get, list, set)

	// Add a defined annotation in order to appear in the help menu
	applicationCmd.Annotations = map[string]string{"command": "other"}
	applicationCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return applicationCmd
}

// AddApplicationFlag adds a `app` flag to the given cobra command
// Also adds a completion handler to the flag
func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ApplicationFlagName, "", "Application, defaults to active application")
	completion.RegisterCommandFlagHandler(cmd, "app", completion.AppCompletionHandler)
}

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string, projectName string) error {
	componentList, err := component.List(client, appName)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	for _, currentComponent := range componentList {
		componentDesc, err := component.GetComponentDesc(client, currentComponent.ComponentName, appName, projectName)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		log.Info("Component", currentComponent.ComponentName, "will be deleted.")

		if len(componentDesc.URLs) != 0 {
			fmt.Println("  Externally exposed URLs will be removed")
		}

		for _, store := range componentDesc.Storage {
			fmt.Println("  Storage", store.Name, "of size", store.Size, "will be removed")
		}

	}
	return nil
}

func validateApp(client *occlient.Client, appName, project string) error {
	exists, err := application.Exists(client, appName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Application %v in project %v does not exist", appName, project)
	}
	return nil
}
