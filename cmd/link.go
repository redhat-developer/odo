package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

var (
	linkComponent string
)

var linkCmd = &cobra.Command{
	Use:   "link <service> --component [component]",
	Short: "Link component to a service",
	Long: `Link component to a service

If source component is not provided, the link is created to the current active
component.

During the linking process, the secret that is created during the service creation (odo service create),
is injected into the component.
`,
	Example: `  # Link the current component to the 'my-postgresql' service
  odo link my-postgresql

  # Link  component 'nodejs' to the 'my-postgresql' service
  odo link my-postgresql --component nodejs
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		componentName := getComponent(client, linkComponent, applicationName, projectName)
		serviceName := args[0]

		exists, err := component.Exists(client, componentName, applicationName, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Component %v does not exist\n", componentName)
			os.Exit(1)
		}

		exists, err = svc.SvcExists(client, serviceName, applicationName, projectName)
		checkError(err, "Service %s doesn't exist within the current namespace", serviceName)
		if !exists {
			fmt.Printf("Service with the name %s does not exist in the current project\n", serviceName)
			os.Exit(1)
		}

		// we also need to check whether there is a secret with the same name as the service
		// the secret should have been created along with the secret
		secret, err := client.GetSecret(serviceName, projectName)
		checkError(err, "Secret %s should have been created along with the service! Please delete and recreate it", serviceName)
		if secret == nil {
			fmt.Printf("Secret %v does not exist\n", serviceName)
			os.Exit(1)
		}

		err = client.LinkSecret(serviceName, componentName, applicationName, projectName)
		checkError(err, "")

		fmt.Printf("Service %s has been successfully linked to the component %s.\n", serviceName, applicationName)
	},
}

func init() {
	linkCmd.PersistentFlags().StringVar(&linkComponent, "component", "", "Component to add link to, defaults to active component")

	// Add a defined annotation in order to appear in the help menu
	linkCmd.Annotations = map[string]string{"command": "component"}
	linkCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(linkCmd)
}
