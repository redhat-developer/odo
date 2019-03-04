package application

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/storage"

	"github.com/redhat-developer/odo/pkg/component"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended app command name
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

	for _, currentComponent := range componentList.Items {
		componentDesc, err := component.GetComponent(client, currentComponent.Name, appName, projectName)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		log.Info("Component", currentComponent.Name, "will be deleted.")

		if len(componentDesc.Spec.URL) != 0 {
			fmt.Println("  Externally exposed URLs will be removed")
		}
		storages, err := storage.List(client, currentComponent.Name, appName)
		odoutil.LogErrorAndExit(err, "")
		for _, storageName := range componentDesc.Spec.Storage {
			store := storages.Get(storageName)
			fmt.Println("  Storage", store.Name, "of size", store.Spec.Size, "will be removed")
		}

	}
	return nil
}

func ensureAppExists(client *occlient.Client, appName, project string) error {
	exists, err := application.Exists(client, appName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Application %v in project %v does not exist", appName, project)
	}
	return nil
}

// getMachineReadableFormat returns resource information in machine readable format
func getMachineReadableFormat(client *occlient.Client, appName string, projectName string, active bool) application.App {
	componentList, _ := component.List(client, appName)
	var compList []string
	for _, component := range componentList.Items {
		compList = append(compList, component.Name)
	}
	appDef := application.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "app",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: projectName,
		},
		Spec: application.AppSpec{
			Components: compList,
		},
		Status: application.AppStatus{
			Active: active,
		},
	}
	return appDef
}
