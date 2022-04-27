package add

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// BindingRecommendedCommandName is the recommended binding sub-command name
const BindingRecommendedCommandName = "binding"

var addBindingExample = ktemplates.Examples(`
# Add binding between service named 'myservice' and the component present in the working directory in the interactive mode
%[1]s

# Add binding between service named 'myservice' and the component present in the working directory
%[1]s --service myservice --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis', and APIGroup 'redis.redis.opstreelab.in' and the component present in the working directory 
%[1]s --service myservice/Redis.redis.redis.opstreelab.in --name myRedisService
%[1]s --service myservice.Redis.redis.redis.opstreelab.in --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis' and the component present in the working directory
%[1]s --service myservice/Redis --name myRedisService

%[1]s --service myservice.Redis --name myRedisService

`)

type AddBindingOptions struct {
	// name of ServiceBinding instance
	name string
	// service is name of the service to be bound to the component
	service string
	// bindAsFiles decides if the service should be bind as a file
	bindAsFiles bool

	flags map[string]string

	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset
}

// NewAddBindingOptions returns new instance of ComponentOptions
func NewAddBindingOptions() *AddBindingOptions {
	return &AddBindingOptions{}
}

func (o *AddBindingOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *AddBindingOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	if err != nil {
		return err
	}

	// this ensures that the namespace as set in env.yaml is used
	o.clientset.KubernetesClient.SetNamespace(o.GetProject())

	o.flags = o.clientset.BindingClient.GetFlags(cmdline.GetFlags())

	return nil
}

func (o *AddBindingOptions) Validate() (err error) {
	return o.clientset.BindingClient.Validate(o.flags)
}

func (o *AddBindingOptions) Run(_ context.Context) error {
	getServices, serviceMap, err := o.clientset.BindingClient.GetServiceInstances()
	if err != nil {
		return err
	}
	if len(getServices) == 0 {
		return fmt.Errorf("No bindable service instances found")
	}
	service, err := o.clientset.BindingClient.SelectServiceInstance(o.flags, getServices, serviceMap)
	if err != nil {
		return err
	}
	bindingName, err := o.clientset.BindingClient.AskBindingName(o.EnvSpecificInfo.GetDevfileObj().GetMetadataName(), o.flags)
	if err != nil {
		return err
	}
	bindAsFiles, err := o.clientset.BindingClient.AskBindAsFiles(o.flags)
	if err != nil {
		return err
	}

	componentContext, err := os.Getwd()
	if err != nil {
		return err
	}
	return o.clientset.BindingClient.CreateBinding(service, bindingName, bindAsFiles, o.EnvSpecificInfo.GetDevfileObj(), serviceMap, componentContext)
}

// NewCmdBinding implements the component odo sub-command
func NewCmdBinding(name, fullName string) *cobra.Command {
	o := NewAddBindingOptions()

	var bindingCmd = &cobra.Command{
		Use:     name,
		Short:   "Add Binding",
		Long:    "Add a binding between a service and the component in the devfile",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(addBindingExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bindingCmd.Flags().StringVar(&o.name, "name", "", "Name of the Binding to create")
	bindingCmd.Flags().StringVar(&o.service, "service", "", "Name of the service to bind")
	bindingCmd.Flags().BoolVarP(&o.bindAsFiles, "bind-as-files", "", true, "Bind the service as a file")
	clientset.Add(bindingCmd, clientset.BINDING)

	return bindingCmd
}
