package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.MaximumNArgs(1),
	Short: "Update the source code path of a component",
	Long:  "Update the source code path of a component",
	Example: `  # Change the source code path of a currently active component to local (use the current directory as a source)
  odo update --local

  # Change the source code path of the frontend component to local with source in ./frontend directory
  odo update frontend --local ./frontend

  # Change the source code path of a currently active component to git 
  odo update --git https://github.com/openshift/nodejs-ex.git

  # Change the source code path of the component named node-ex to git
  odo update node-ex --git https://github.com/openshift/nodejs-ex.git

  # Change the source code path of the component named wildfly to a binary named sample.war in ./downloads directory
  odo update wildfly --binary ./downloads/sample.war
	`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()

		projectName := getAndSetNamespace(client)
		applicationName := getAppName(client)

		stdout := color.Output

		checkFlag := 0

		if len(componentBinary) != 0 {
			checkFlag++
		}
		if len(componentGit) != 0 {
			checkFlag++
		}
		if len(componentLocal) != 0 {
			checkFlag++
		}

		if checkFlag != 1 {
			fmt.Println("The source can be either --binary or --local or --git")
			os.Exit(1)
		}

		var componentName string

		if len(args) == 0 {
			componentName, err := component.GetCurrent(applicationName, projectName)
			checkError(err, "unable to get current component")
			if len(componentName) == 0 {
				appList, err := application.ListInProject(client)
				checkError(err, "")
				if len(appList) == 0 {
					fmt.Println("Cannot update as no application exists in the current projectName")
					os.Exit(1)
				}
			}
		} else {
			componentName = args[0]
			exists, err := component.Exists(client, componentName, applicationName)
			checkError(err, "")
			if !exists {
				fmt.Printf("Component with name %s does not exist in the current application\n", componentName)
				os.Exit(1)
			}
		}

		if len(applicationName) == 0 {
			fmt.Println("Cannot update as no application is set as active")
			os.Exit(1)
		}

		if len(componentGit) != 0 {
			err := component.Update(client, componentName, applicationName, "git", componentGit, stdout)
			checkError(err, "")
			fmt.Printf("The component %s was updated successfully\n", componentName)
		} else if len(componentLocal) != 0 {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs(componentLocal)
			checkError(err, "")
			fileInfo, err := os.Stat(dir)
			checkError(err, "")
			if !fileInfo.IsDir() {
				fmt.Println("Please provide a path to the directory")
				os.Exit(1)
			}
			err = component.Update(client, componentName, applicationName, "local", dir, stdout)
			checkError(err, "")
			fmt.Printf("The component %s was updated successfully, please use 'odo push' to push your local changes\n", componentName)
		} else if len(componentBinary) != 0 {
			path, err := filepath.Abs(componentBinary)
			checkError(err, "")
			err = component.Update(client, componentName, applicationName, "binary", path, stdout)
			checkError(err, "")
			fmt.Printf("The component %s was updated successfully, please use 'odo push' to push your local changes\n", componentName)
		}
	},
}

func init() {
	updateCmd.Flags().StringVarP(&componentBinary, "binary", "b", "", "binary artifact")
	updateCmd.Flags().StringVarP(&componentGit, "git", "g", "", "git source")
	updateCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use local directory as a source for component.")

	// Add a defined annotation in order to appear in the help menu
	updateCmd.Annotations = map[string]string{"command": "component"}
	updateCmd.SetUsageTemplate(cmdUsageTemplate)

	//Adding `--application` flag
	addApplicationFlag(updateCmd)

	//Adding `--project` flag
	addProjectFlag(updateCmd)

	rootCmd.AddCommand(updateCmd)
}
