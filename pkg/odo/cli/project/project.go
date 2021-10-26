package project

import (
	"fmt"

	"github.com/openshift/odo/v2/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/v2/pkg/odo/util"
	"github.com/openshift/odo/v2/pkg/odo/util/completion"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended project command name
const RecommendedCommandName = "project"

// NewCmdProject implements the project odo command
func NewCmdProject(name, fullName string) *cobra.Command {

	projectCreateCmd := NewCmdProjectCreate(createRecommendedCommandName, odoutil.GetFullName(fullName, createRecommendedCommandName))
	projectSetCmd := NewCmdProjectSet(setRecommendedCommandName, odoutil.GetFullName(fullName, setRecommendedCommandName))
	projectListCmd := NewCmdProjectList(listRecommendedCommandName, odoutil.GetFullName(fullName, listRecommendedCommandName))
	projectDeleteCmd := NewCmdProjectDelete(deleteRecommendedCommandName, odoutil.GetFullName(fullName, deleteRecommendedCommandName))
	projectGetCmd := NewCmdProjectGet(getRecommendedCommandName, odoutil.GetFullName(fullName, getRecommendedCommandName))

	projectCmd := &cobra.Command{
		Use:   name + " [options]",
		Short: "Perform project operations",
		Long:  "Perform project operations",
		Example: fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			projectSetCmd.Example,
			projectCreateCmd.Example,
			projectListCmd.Example,
			projectDeleteCmd.Example,
			projectGetCmd.Example),
		// 'odo project' is the same as 'odo project get'
		// 'odo project <project_name>' is the same as 'odo project set <project_name>'
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 && args[0] != getRecommendedCommandName && args[0] != setRecommendedCommandName {
				projectSetCmd.Run(cmd, args)
			} else {
				projectGetCmd.Run(cmd, args)
			}
		},
	}

	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectSetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add a defined annotation in order to appear in the help menu
	projectCmd.Annotations = map[string]string{"command": "main"}
	projectCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(projectSetCmd, completion.ProjectNameCompletionHandler)
	completion.RegisterCommandHandler(projectDeleteCmd, completion.ProjectNameCompletionHandler)

	return projectCmd
}

// AddProjectFlag adds a `project` flag to the given cobra command
// Also adds a completion handler to the flag
func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ProjectFlagName, "", "Project, defaults to active project")
	completion.RegisterCommandFlagHandler(cmd, "project", completion.ProjectNameCompletionHandler)
}

// TODO: The following function should work with devfile components, and use good abstraction that will make it possible to share
// code with 'odo describe' and 'odo list'
// printDeleteProjectInfo prints objects affected by project deletion
// func printDeleteProjectInfo(client *occlient.Client, projectName string) error {
// 	localConfig, err := config.New()
// 	if err != nil {
// 		return errors.Wrapf(err, "unable to get the local config")
// 	}
// 	// Fetch and List the applications
// 	applicationList, err := application.ListInProject(client)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get application list")
// 	}
// 	if len(applicationList) != 0 {
// 		log.Info("This project contains the following applications, which will be deleted")
// 		for _, app := range applicationList {
// 			log.Info("Application", app)

// 			// List the components
//			var selector string
//			if app != "" {
//				selector = applabels.GetSelector(app)
//			}
// 			componentList, err := component.List(client, selector, nil)
// 			if err != nil {
// 				return errors.Wrap(err, "failed to get Component list")
// 			}
// 			if len(componentList.Items) != 0 {
// 				log.Info("This application has following components that will be deleted")

// 				for _, currentComponent := range componentList.Items {
// 					componentDesc, err := component.GetComponent(client, currentComponent.Name, app, projectName)
// 					if err != nil {
// 						return errors.Wrap(err, "unable to get component description")
// 					}
// 					log.Info("component named", componentDesc.Name)

// 					if len(componentDesc.Spec.URL) != 0 {
// 						ul, err := url.ListPushed(client, componentDesc.Name, app)
// 						if err != nil {
// 							return errors.Wrap(err, "Could not get url list")
// 						}
// 						log.Info("This component has following urls that will be deleted with component")
// 						for _, u := range ul.Items {
// 							log.Info("URL named", u.GetName(), "with host", u.Spec.Host, "having protocol", u.Spec.Protocol, "at port", u.Spec.Port)
// 						}
// 					}

// 					storages, err := localConfig.StorageList()
// 					odoutil.LogErrorAndExit(err, "")
// 					if len(storages) != 0 {
// 						log.Info("This component has following storages which will be deleted with the component")
// 						for _, store := range storages {
// 							log.Info("Storage named", store.Name, "of size", store.Size)
// 						}
// 					}
// 				}
// 			}

// 			// List services that will be removed
// 			serviceList, err := service.List(client, app)
// 			if err != nil {
// 				log.Info("No services / could not get services")
// 				klog.V(4).Info(err.Error())
// 			}

// 			if len(serviceList.Items) != 0 {
// 				log.Info("This application has following service that will be deleted")
// 				for _, ser := range serviceList.Items {
// 					log.Info("service named", ser.ObjectMeta.Name, "of type", ser.Spec.Type)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
