package cmd

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var (
	linkComponent string
)

var linkCmd = &cobra.Command{
	Use:   "link <target component> --component [source component]",
	Short: "Link target component to source component",
	Long: `Link target component to source component

If source component is not provided, the link is created to the current active
component.

In the linking process, the environment variables containing the connection
information from target component are injected into the source component and
printed to STDOUT.
`,
	Example: `  # Link current component to a component 'mariadb'
  odo link mariadb

  # Link 'mariadb' component to 'nodejs' component
  odo link mariadb --component nodejs
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		sourceComponent := getComponent(client, linkComponent, applicationName, projectName)
		targetComponent := args[0]

		exists, err := component.Exists(client, sourceComponent, applicationName, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Component %v does not exist\n", sourceComponent)
		}
		exists, err = component.Exists(client, targetComponent, applicationName, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Component %v does not exist\n", targetComponent)
		}

		linkInfo, err := component.Link(client, sourceComponent, targetComponent, applicationName)
		checkError(err, fmt.Sprintf("Failed to link %v to %v", targetComponent, sourceComponent))

		fmt.Printf("Successfully linked %v to %v\n", targetComponent, sourceComponent)
		fmt.Printf("The following environment variables have been injected in %v to connect to %v\n", linkInfo.SourceComponent, linkInfo.TargetComponent)
		for _, env := range linkInfo.Envs {
			fmt.Printf("- %v\n", env)
		}
	},
}

func init() {
	linkCmd.PersistentFlags().StringVarP(&linkComponent, "component", "c", "", "Component to add link to, defaults to active component")

	// Add a defined annotation in order to appear in the help menu
	linkCmd.Annotations = map[string]string{"command": "component"}
	linkCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(linkCmd)
}
