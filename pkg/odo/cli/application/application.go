package application

import (
	"fmt"

	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended app command name
const RecommendedCommandName = "app"

// NewCmdApplication implements the odo application command
func NewCmdApplication(name, fullName string) *cobra.Command {
	delete := NewCmdDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	describe := NewCmdDescribe(describeRecommendedCommandName, odoutil.GetFullName(fullName, describeRecommendedCommandName))
	list := NewCmdList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	applicationCmd := &cobra.Command{
		Use:   name,
		Short: "Perform application operations",
		Long:  `Performs application operations related to your project.`,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			delete.Example,
			describe.Example,
			list.Example),
		Aliases: []string{"application"},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	applicationCmd.AddCommand(delete, describe, list)

	// Add a defined annotation in order to appear in the help menu
	applicationCmd.Annotations = map[string]string{"command": "main"}
	applicationCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return applicationCmd
}

// AddApplicationFlag adds a `app` flag to the given cobra command
// Also adds a completion handler to the flag
func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ApplicationFlagName, "", "Application, defaults to active application")
	completion.RegisterCommandFlagHandler(cmd, "app", completion.AppCompletionHandler)
}

// printAppInfo will print things which will be deleted
func printAppInfo(client *occlient.Client, kClient kclient.ClientInterface, appName string, projectName string) error {
	var selector string
	if appName != "" {
		selector = applabels.GetSelector(appName)
	}
	componentList, err := component.List(client, selector)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	if len(componentList.Items) != 0 {
		log.Info("This application has following components that will be deleted")
		for _, currentComponent := range componentList.Items {
			log.Info("component named", currentComponent.Name)

			if len(currentComponent.Spec.URL) != 0 {
				log.Info("This component has following urls that will be deleted with component")
				for _, u := range currentComponent.Spec.URLSpec {
					log.Info("URL named", u.GetName(), "with host", u.Spec.Host, "having protocol", u.Spec.Protocol, "at port", u.Spec.Port)
				}
			}

			if len(currentComponent.Spec.Storage) != 0 {
				log.Info("The component has following storages which will be deleted with the component")
				for _, storage := range currentComponent.Spec.StorageSpec {
					store := storage
					log.Info("Storage named", store.GetName(), "of size", store.Spec.Size)
				}
			}
		}
	}
	return nil
}
