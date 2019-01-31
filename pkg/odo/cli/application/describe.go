package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	"os"
)

var applicationDescribeCmd = &cobra.Command{
	Use:   "describe [application_name]",
	Short: "Describe the given application",
	Long:  "Describe the given application",
	Args:  cobra.MaximumNArgs(1),
	Example: `  # Describe webapp application,
  odo app describe webapp
	`,
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		appName := context.Application
		if len(args) == 0 {
			if appName == "" {
				log.Errorf("There's no active application in project: %v", projectName)
				os.Exit(1)
			}
		} else {
			appName = args[0]
			//Check whether application exist or not
			exists, err := application.Exists(client, appName)
			util.LogErrorAndExit(err, "")
			if !exists {
				log.Errorf("Application with the name %s does not exist in %s ", appName, projectName)
				os.Exit(1)
			}
		}

		// List of Component
		componentList, err := component.List(client, appName)
		util.LogErrorAndExit(err, "")

		//we ignore service errors here because it's entirely possible that the service catalog has not been installed
		serviceList, _ := service.ListWithDetailedStatus(client, appName)

		if len(componentList) == 0 && len(serviceList) == 0 {
			log.Errorf("Application %s has no components or services deployed.", appName)
		} else {
			fmt.Printf("Application Name: %s has %v component(s) and %v service(s):\n--------------------------------------\n",
				appName, len(componentList), len(serviceList))
			if len(componentList) > 0 {
				for _, currentComponent := range componentList {
					componentDesc, err := component.GetComponentDesc(client, currentComponent.Name, appName, projectName)
					util.LogErrorAndExit(err, "")
					util.PrintComponentInfo(currentComponent.Name, componentDesc)
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

	},
}
