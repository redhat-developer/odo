package application

import (
	"fmt"
	"os"
	"strings"

	"encoding/json"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/odo/util/validation"
	"github.com/redhat-developer/odo/pkg/service"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	applicationShortFlag       bool
	applicationForceDeleteFlag bool
	outputFlag                 string
)

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:   "app",
	Short: "Perform application operations",
	Long:  `Performs application operations related to your OpenShift project.`,
	Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		applicationCreateCmd.Example,
		applicationGetCmd.Example,
		applicationDeleteCmd.Example,
		applicationDescribeCmd.Example,
		applicationListCmd.Example,
		applicationSetCmd.Example),
	Aliases: []string{"application"},
	// 'odo app' is the same as 'odo app get'
	// 'odo app <application_name>' is the same as 'odo app set <application_name>'
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] != "get" && args[0] != "set" {
			applicationSetCmd.Run(cmd, args)
		} else {
			applicationGetCmd.Run(cmd, args)
		}
	},
}

var applicationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application",
	Long: `Create an application.
If no app name is passed, a default app name will be auto-generated.
	`,
	Example: `  # Create an application
  odo app create myapp
  odo app create
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		var appName string
		if len(args) == 1 {
			// The only arg passed is the app name
			appName = args[0]
		} else {
			// Desired app name is not passed so, generate a new app name
			// Fetch existing list of apps
			apps, err := application.List(client)
			odoutil.LogErrorAndExit(err, "")

			// Generate a random name that's not already in use for the existing apps
			appName, err = application.GetDefaultAppName(apps)
			odoutil.LogErrorAndExit(err, "")
		}
		// validate application name
		err := validation.ValidateName(appName)
		odoutil.LogErrorAndExit(err, "")
		log.Progressf("Creating application: %v in project: %v", appName, projectName)
		err = application.Create(client, appName)
		odoutil.LogErrorAndExit(err, "")
		err = application.SetCurrent(client, appName)

		// TODO: updating the app name should be done via SetCurrent and passing the Context
		// not strictly needed here but Context should stay in sync
		context.Application = appName

		odoutil.LogErrorAndExit(err, "")
		log.Infof("Switched to application: %v in project: %v", appName, projectName)
	},
}

var applicationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the active application",
	Long:  "Get the active application",
	Example: `  # Get the currently active application
  odo app get
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		projectName := context.Project
		app := context.Application
		if applicationShortFlag {
			fmt.Print(app)
			return
		}
		if app == "" {
			log.Infof("There's no active application.\nYou can create one by running 'odo application create <name>'.")
			return
		}
		log.Infof("The current application is: %v in project: %v", app, projectName)
	},
}

var applicationDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the given application",
	Long:  "Delete the given application",
	Example: `  # Delete the application
  odo app delete myapp
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project
		appName := context.Application
		if len(args) == 1 {
			// If app name passed, consider it for deletion
			appName = args[0]
		}

		var confirmDeletion string

		// Print App Information which will be deleted
		err := printDeleteAppInfo(client, appName, projectName)
		odoutil.LogErrorAndExit(err, "")
		exists, err := application.Exists(client, appName)
		odoutil.LogErrorAndExit(err, "")
		if !exists {
			log.Errorf("Application %v in project %v does not exist", appName, projectName)
			os.Exit(1)
		}

		if applicationForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			log.Askf("Are you sure you want to delete the application: %v from project: %v? [y/N]: ", appName, projectName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := application.Delete(client, appName)
			odoutil.LogErrorAndExit(err, "")
			log.Infof("Deleted application: %s from project: %v", appName, projectName)
		} else {
			log.Infof("Aborting deletion of application: %v", appName)
		}
	},
}

var applicationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications in the current project",
	Long:  "List all applications in the current project.",
	Example: `  # List all applications in the current project
  odo app list

  # List all applications in the specified project
  odo app list --project myproject
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		apps, err := application.ListInProject(client)
		odoutil.LogErrorAndExit(err, "unable to get list of applications")
		if len(apps) > 0 {

			if outputFlag == "json" {
				var appList []application.App
				for _, app := range apps {
					appDef := getMachineReadableFormat(client, app.Name, projectName, app.Active)
					appList = append(appList, appDef)
				}

				appListDef := application.AppList{
					TypeMeta: metav1.TypeMeta{
						Kind:       "List",
						APIVersion: "odo.openshift.io/v1alpha1",
					},
					ListMeta: metav1.ListMeta{},
					Items:    appList,
				}
				out, err := json.Marshal(appListDef)
				odoutil.LogErrorAndExit(err, "")
				fmt.Println(string(out))

			} else {

				log.Infof("The project '%v' has the following applications:", projectName)
				tabWriter := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
				fmt.Fprintln(tabWriter, "ACTIVE", "\t", "NAME")
				for _, app := range apps {
					activeMark := " "
					if app.Active {
						activeMark = "*"
					}
					fmt.Fprintln(tabWriter, activeMark, "\t", app.Name)
				}
				tabWriter.Flush()
			}
		} else {
			log.Infof("There are no applications deployed in the project '%v'.", projectName)
		}
	},
}

var applicationSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set application as active",
	Long:  "Set application as active",
	Example: `  # Set an application as active
  odo app set myapp
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Error("Please provide application name")
			os.Exit(1)
		}
		if len(args) > 1 {
			log.Error("Only one argument (application name) is allowed")
			os.Exit(1)
		}
		return nil
	}, Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		// error if application does not exist
		appName := args[0]
		exists, err := application.Exists(client, appName)
		odoutil.LogErrorAndExit(err, "unable to check if application exists")
		if !exists {
			log.Errorf("Application %v does not exist", appName)
			os.Exit(1)
		}

		err = application.SetCurrent(client, appName)
		odoutil.LogErrorAndExit(err, "")
		log.Infof("Switched to application: %v in project: %v", args[0], projectName)

		// TODO: updating the app name should be done via SetCurrent and passing the Context
		// not strictly needed here but Context should stay in sync
		context.Application = appName
	},
}

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
			odoutil.LogErrorAndExit(err, "")
			if !exists {
				log.Errorf("Application with the name %s does not exist in %s ", appName, projectName)
				os.Exit(1)
			}
		}

		if outputFlag == "json" {
			app, _ := application.GetCurrent(projectName)
			appDef := getMachineReadableFormat(client, appName, projectName, app == appName)
			out, err := json.Marshal(appDef)
			odoutil.LogErrorAndExit(err, "")
			fmt.Println(string(out))

		} else {

			// List of all the components
			componentList, err := component.List(client, appName)
			odoutil.LogErrorAndExit(err, "")

			//we ignore service errors here because it's entirely possible that the service catalog has not been installed
			serviceList, _ := service.ListWithDetailedStatus(client, appName)

			if len(componentList) == 0 && len(serviceList) == 0 {
				log.Errorf("Application %s has no components or services deployed.", appName)
			} else {
				fmt.Printf("Application Name: %s has %v component(s) and %v service(s):\n--------------------------------------\n",
					appName, len(componentList), len(serviceList))
				if len(componentList) > 0 {
					for _, currentComponent := range componentList {
						componentDesc, err := component.GetComponentDesc(client, currentComponent.ComponentName, appName, projectName)
						odoutil.LogErrorAndExit(err, "")
						odoutil.PrintComponentInfo(currentComponent.ComponentName, componentDesc)
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

	},
}

// getMachineReadableFormat returns resource information in machine readable format
func getMachineReadableFormat(client *occlient.Client, appName string, projectName string, active bool) application.App {
	componentList, _ := component.List(client, appName)
	var compList []string
	for _, comp := range componentList {
		compList = append(compList, comp.ComponentName)
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

// NewCmdApplication implements the odo application command
func NewCmdApplication() *cobra.Command {
	applicationDeleteCmd.Flags().BoolVarP(&applicationForceDeleteFlag, "force", "f", false, "Delete application without prompting")

	applicationGetCmd.Flags().BoolVarP(&applicationShortFlag, "short", "q", false, "If true, display only the application name")

	applicationDescribeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output in json format")
	applicationListCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output in json format")

	// add flags from 'get' to application command
	applicationCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	applicationCmd.AddCommand(applicationListCmd)
	applicationCmd.AddCommand(applicationDeleteCmd)
	applicationCmd.AddCommand(applicationGetCmd)
	applicationCmd.AddCommand(applicationCreateCmd)
	applicationCmd.AddCommand(applicationSetCmd)
	applicationCmd.AddCommand(applicationDescribeCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(applicationListCmd)
	projectCmd.AddProjectFlag(applicationCreateCmd)
	projectCmd.AddProjectFlag(applicationDeleteCmd)
	projectCmd.AddProjectFlag(applicationDescribeCmd)
	projectCmd.AddProjectFlag(applicationSetCmd)
	projectCmd.AddProjectFlag(applicationGetCmd)

	// Add a defined annotation in order to appear in the help menu
	applicationCmd.Annotations = map[string]string{"command": "other"}
	applicationCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(applicationDescribeCmd, completion.AppCompletionHandler)
	completion.RegisterCommandHandler(applicationDeleteCmd, completion.AppCompletionHandler)
	completion.RegisterCommandHandler(applicationSetCmd, completion.AppCompletionHandler)

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
