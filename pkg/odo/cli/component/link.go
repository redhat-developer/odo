package component

import (
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"os"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/secret"

	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

var (
	port string
	wait bool
)

var linkCmd = &cobra.Command{
	Use:   "link <service> --component [component] OR link <component> --component [component]",
	Short: "Link component to a service or component",
	Long: `Link component to a service or component

If the source component is not provided, the current active component is assumed.

In both use cases, link adds the appropriate secret to the environment of the source component. 
The source component can then consume the entries of the secret as environment variables.

For example:

We have created a frontend application called 'frontend':
odo create nodejs frontend

We've also created a backend application called 'backend' with port 8080 exposed:
odo create nodejs backend --port 8080

You can now link the two applications:
odo link backend --component frontend

Now the frontend has 2 ENV variables it can use:
COMPONENT_BACKEND_HOST=backend-app
COMPONENT_BACKEND_PORT=8080

If you wish to use a database, we can use the Service Catalog and link it to our backend:
odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret
odo link dh-postgresql-apb

Now backend has 2 ENV variables it can use:
DB_USER=luke
DB_PASSWORD=secret
`,
	Example: `  # Link the current component to the 'my-postgresql' service
  odo link my-postgresql

  # Link component 'nodejs' to the 'my-postgresql' service
  odo link my-postgresql --component nodejs

  # Link current component to the 'backend' component (backend must have a single exposed port)
  odo link backend

  # Link component 'nodejs' to the 'backend' component
  odo link backend --component nodejs

  # Link current component to port 8080 of the 'backend' component (backend must have port 8080 exposed) 
  odo link backend --port 8080
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project
		applicationName := context.Application
		sourceComponentName := context.Component()

		suppliedName := args[0]

		svcSxists, err := svc.SvcExists(client, suppliedName, applicationName)
		odoutil.CheckError(err, "Unable to determine if service %s exists", suppliedName)

		cmpExists, err := component.Exists(client, suppliedName, applicationName)
		odoutil.CheckError(err, "Unable to determine if component %s exists", suppliedName)

		if svcSxists {
			if cmpExists {
				glog.V(4).Infof("Both a service and component with name %s - assuming a link to the service is required", suppliedName)
			}

			serviceName := suppliedName

			// if there is a ServiceBinding, then that means there is already a secret (or there will be soon)
			// which we can link to
			_, err = client.GetServiceBinding(serviceName, projectName)
			if err != nil {
				log.Errorf(`The service was not created via Odo. Please delete the service and recreate it using 'odo service create %s'`, serviceName)
				os.Exit(1)
			}

			if wait {
				// we wait until the secret has been created on the OpenShift
				// this is done because the secret is only created after the Pod that runs the
				// service is in running state.
				// This can take a long time to occur if the image of the service has yet to be downloaded
				log.Progressf("Waiting for secret of service %s to come up", serviceName)
				_, err = client.WaitAndGetSecret(serviceName, projectName)
				odoutil.CheckError(err, "")
			} else {
				// we also need to check whether there is a secret with the same name as the service
				// the secret should have been created along with the secret
				_, err = client.GetSecret(serviceName, projectName)
				if err != nil {
					log.Errorf(`The service %s created by 'odo service create' is being provisioned. You may have to wait a few seconds until OpenShift fully provisions it.`, serviceName)
					os.Exit(1)
				}
			}

			err = client.LinkSecret(serviceName, sourceComponentName, applicationName, projectName)
			odoutil.CheckError(err, "")
			log.Successf("Service %s has been successfully linked to the component %s", serviceName, sourceComponentName)
		} else if cmpExists {
			targetComponent := args[0]

			secretName, err := secret.DetermineSecretName(client, targetComponent, applicationName, port)
			odoutil.CheckError(err, "")

			err = client.LinkSecret(secretName, sourceComponentName, applicationName, projectName)
			odoutil.CheckError(err, "")
			log.Successf("Component %s has been successfully linked to component %s", targetComponent, sourceComponentName)
		} else {
			log.Errorf(`Neither a service nor a component named %s could be located. Please create one of the two before attempting to use odo link`, suppliedName)
			os.Exit(1)
		}
	},
}

// NewCmdLink implements the link odo command
func NewCmdLink() *cobra.Command {
	linkCmd.PersistentFlags().StringVar(&port, "port", "", "Port of the backend to which to link")
	linkCmd.PersistentFlags().BoolVarP(&wait, "wait", "w", false, "If enabled, the link command will wait for the service to be provisioned")

	// Add a defined annotation in order to appear in the help menu
	linkCmd.Annotations = map[string]string{"command": "component"}
	linkCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	//Adding `--project` flag
	completion.AddProjectFlag(linkCmd)
	//Adding `--application` flag
	completion.AddApplicationFlag(linkCmd)
	//Adding `--component` flag
	completion.AddComponentFlag(linkCmd)

	completion.RegisterCommandHandler(linkCmd, completion.LinkCompletionHandler)

	return linkCmd
}
