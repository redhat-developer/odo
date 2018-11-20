package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"os"
	"path/filepath"

	"github.com/fatih/color"
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
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

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
			componentName = context.Component()
		} else {
			componentName = context.Component(args[0])
		}

		if len(applicationName) == 0 {
			fmt.Println("Cannot update as no application is set as active")
			os.Exit(1)
		}

		if len(componentGit) != 0 {
			err := component.Update(client, componentName, applicationName, "git", componentGit, stdout)
			util.CheckError(err, "")
			fmt.Printf("The component %s was updated successfully\n", componentName)
		} else if len(componentLocal) != 0 {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs(componentLocal)
			util.CheckError(err, "")
			fileInfo, err := os.Stat(dir)
			util.CheckError(err, "")
			if !fileInfo.IsDir() {
				fmt.Println("Please provide a path to the directory")
				os.Exit(1)
			}
			err = component.Update(client, componentName, applicationName, "local", dir, stdout)
			util.CheckError(err, "")
			fmt.Printf("The component %s was updated successfully, please use 'odo push' to push your local changes\n", componentName)
		} else if len(componentBinary) != 0 {
			path, err := filepath.Abs(componentBinary)
			util.CheckError(err, "")
			err = component.Update(client, componentName, applicationName, "binary", path, stdout)
			util.CheckError(err, "")
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
	updateCmd.SetUsageTemplate(cli.CmdUsageTemplate)

	//Adding `--application` flag
	genericclioptions.AddApplicationFlag(updateCmd)

	//Adding `--project` flag
	genericclioptions.AddProjectFlag(updateCmd)

	completion.RegisterCommandFlagHandler(updateCmd, "local", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(updateCmd, "binary", completion.FileCompletionHandler)

	cli.RootCmd().AddCommand(updateCmd)
}
