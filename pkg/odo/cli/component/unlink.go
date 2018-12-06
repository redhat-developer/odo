package component

import (
	"fmt"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/secret"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

// RecommendedUnlinkCommandName is the recommended unlink command name
const RecommendedUnlinkCommandName = "unlink"

var (
	unlinkExample = ktemplates.Examples(`# Unlink the 'my-postgresql' service from the current component 
%[1]s my-postgresql

# Unlink the 'my-postgresql' service  from the 'nodejs' component
%[1]s my-postgresql --component nodejs

# Unlink the 'backend' component from the current component (backend must have a single exposed port)
%[1]s backend

# Unlink the 'backend' service  from the 'nodejs' component
%[1]s backend --component nodejs

# Unlink the backend's 8080 port from the current component 
%[1]s backend --port 8080`)

	unlinkLongDesc = `Unlink component or service from a component. 
For this command to be successful, the service or component needs to have been linked prior to the invocation using 'odo link'`
)

// UnlinkOptions encapsulates the options for the odo link command
type UnlinkOptions struct {
	port         string
	suppliedName string
	*genericclioptions.Context
}

// NewUnlinkOptions creates a new LinkOptions instance
func NewUnlinkOptions() *UnlinkOptions {
	return &UnlinkOptions{}
}

// Complete completes UnlinkOptions after they've been created
func (o *UnlinkOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.suppliedName = args[0]
	o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	return err
}

// Run contains the logic for the odo link command
func (o *UnlinkOptions) Run() (err error) {
	client := o.Client
	svcExists, err := svc.SvcExists(client, o.suppliedName, o.Application)
	if err != nil {
		return fmt.Errorf("Unable to determine if service exists:\n%v", err)
	}

	cmpExists, err := component.Exists(client, o.suppliedName, o.Application)
	if err != nil {
		return fmt.Errorf("Unable to determine if component exists:\n%v", err)
	}

	if svcExists {
		if cmpExists {
			glog.V(4).Infof("Both a service and component with name %s - assuming a link to the service is required", o.suppliedName)
		}

		// we check whether there is a secret with the same name as the service
		// the secret should have been created along with the secret
		_, err = client.GetSecret(o.suppliedName, o.Project)
		if err != nil {
			return fmt.Errorf("The service %s created by 'odo service create' is being provisioned. It doesn't make sense to call unlink unless the service has been provisioned", o.suppliedName)
		}

		err = client.UnlinkSecret(o.suppliedName, o.Component(), o.Application)
		if err != nil {
			return err
		}

		log.Successf("Service %s has been successfully unlinked from the component %s", o.suppliedName, o.Component())
		return nil
	} else if cmpExists {
		secretName, err := secret.DetermineSecretName(client, o.suppliedName, o.Application, o.port)
		if err != nil {
			return err
		}

		err = client.UnlinkSecret(secretName, o.Component(), o.Application)
		if err != nil {
			return err
		}

		log.Successf("Component %s has been successfully unlinked from component %s", o.suppliedName, o.Component())
		return nil
	} else {
		return fmt.Errorf("Neither a service nor a component named %s could be located. Unlink should not be called unless the target service or component exists", o.suppliedName)
	}
}

// NewCmdUnlink implements the link odo command
func NewCmdUnlink(name, fullName string) *cobra.Command {
	o := NewUnlinkOptions()

	unlinkCmd := &cobra.Command{
		Use:     "unlink <service> --component [component] OR unlink <component> --component [component]",
		Short:   "Unlink component to a service or component",
		Long:    unlinkLongDesc,
		Example: fmt.Sprintf(unlinkExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckError(o.Complete(name, cmd, args), "")
			util.CheckError(o.Run(), "")
		},
	}

	unlinkCmd.PersistentFlags().StringVar(&o.port, "port", "", "Port of the backend to which to unlink")

	// Add a defined annotation in order to appear in the help menu
	unlinkCmd.Annotations = map[string]string{"command": "component"}
	unlinkCmd.SetUsageTemplate(util.CmdUsageTemplate)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(unlinkCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(unlinkCmd)
	//Adding `--component` flag
	AddComponentFlag(unlinkCmd)

	completion.RegisterCommandHandler(unlinkCmd, completion.UnlinkCompletionHandler)

	return unlinkCmd
}
