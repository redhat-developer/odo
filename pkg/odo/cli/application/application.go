package application

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/url"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/component"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/service"
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
		Long:  `Performs application operations related to your OpenShift project.`,
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

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string, projectName string) error {
	componentList, err := component.List(client, appName, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	if len(componentList.Items) != 0 {
		log.Info("This application has following components that will be deleted")
		for _, currentComponent := range componentList.Items {
			componentDesc, err := component.GetComponent(client, currentComponent.Name, appName, projectName)
			if err != nil {
				return errors.Wrap(err, "unable to get component description")
			}
			log.Info("component named", currentComponent.Name)

			if len(componentDesc.Spec.URL) != 0 {
				ul, err := url.ListPushed(client, componentDesc.Name, appName)
				if err != nil {
					return errors.Wrap(err, "Could not get url list")
				}
				log.Info("This component has following urls that will be deleted with component")
				for _, u := range ul.Items {
					log.Info("URL named", u.GetName(), "with host", u.Spec.Host, "having protocol", u.Spec.Protocol, "at port", u.Spec.Port)
				}
			}

			storages, err := storage.List(client, currentComponent.Name, appName)
			odoutil.LogErrorAndExit(err, "")
			if len(storages.Items) != 0 {
				log.Info("The component has following storages which will be deleted with the component")
				for _, storageName := range componentDesc.Spec.Storage {
					store := storages.Get(storageName)
					log.Info("Storage named", store.GetName(), "of size", store.Spec.Size)
				}
			}
		}
	}
	// List services that will be removed
	serviceList, err := service.List(client, appName)
	if err != nil {
		log.Info("No services / could not get services")
		klog.V(4).Info(err.Error())
	}
	if len(serviceList.Items) != 0 {
		log.Info("This application has following service(s) that will be deleted")
		for _, ser := range serviceList.Items {
			log.Info("service named", ser.ObjectMeta.Name, "of type", ser.Spec.Type)
		}
	}

	return nil
}
