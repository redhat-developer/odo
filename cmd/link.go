package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/redhat-developer/odo/pkg/component"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
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

  # Link component 'nodejs' to the 'my-postgresql' service
  odo link my-postgresql --component nodejs
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetOcClient()
		projectName := util.GetAndSetNamespace(client)
		applicationName := util.GetAppName(client)
		componentName := util.GetComponent(client, util.ComponentFlag, applicationName)
		serviceName := args[0]

		exists, err := component.Exists(client, componentName, applicationName)
		util.CheckError(err, "")
		if !exists {
			fmt.Printf("Component %v does not exist\n", componentName)
			os.Exit(1)
		}

		exists, err = svc.SvcExists(client, serviceName, applicationName)
		util.CheckError(err, "Unable to determine if service %s exists within the current namespace", serviceName)
		if !exists {
			fmt.Printf(`Service %s does not exist within the current namespace.
Please perform 'odo service create %s ...' before attempting to link the service.`, serviceName, serviceName)
			os.Exit(1)
		}

		// we also need to check whether there is a secret with the same name as the service
		// the secret should have been created along with the secret
		_, err = client.GetSecret(serviceName, projectName)
		if err != nil {
			fmt.Printf(`Secret %s should have been created along with the service
If you previously created the service with 'odo service create', then you may have to wait a few seconds until OpenShift provisions it.
If not, then please delete the service and recreate it using 'odo service create %s`, serviceName, serviceName)
			os.Exit(1)
		}

		err = client.LinkSecret(serviceName, componentName, applicationName, projectName)
		util.CheckError(err, "")

		fmt.Printf("Service %s has been successfully linked to the component %s.\n", serviceName, applicationName)
	},
}

func init() {

	// Add a defined annotation in order to appear in the help menu
	linkCmd.Annotations = map[string]string{"command": "component"}
	linkCmd.SetUsageTemplate(cmdUsageTemplate)
	//Adding `--project` flag
	addProjectFlag(linkCmd)
	//Adding `--application` flag
	addApplicationFlag(linkCmd)
	// Adding `--component` flag
	addComponentFlag(linkCmd)

	rootCmd.AddCommand(linkCmd)
}
