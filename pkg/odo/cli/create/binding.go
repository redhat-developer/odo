package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// BindingRecommendedCommandName is the recommended binding sub-command name
const BindingRecommendedCommandName = "binding"

var createBindingExample = ktemplates.Examples(`
# Create binding
%[1]s --service myservice/Redis.redis.redis.opstreelab.in --name myRedisService

# Create binding as a file
%[1]s --service myservice/Redis.redis.redis.opstreelab.in --name myRedisService --bind-as-files

# Create binding interactively
%[1]s
`)

type CreateBindingOptions struct {
	// name of the service to bind
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

// NewCreateBindingOptions returns new instance of ComponentOptions
func NewCreateBindingOptions() *CreateBindingOptions {
	return &CreateBindingOptions{}
}

func (o *CreateBindingOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *CreateBindingOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	if err != nil {
		return err
	}

	// this ensures that the namespace as set in env.yaml is used
	o.clientset.KubernetesClient.SetNamespace(o.GetProject())

	o.flags = o.clientset.BindingClient.GetFlags(cmdline.GetFlags())

	return nil
}

func (o *CreateBindingOptions) Validate() (err error) {
	return o.clientset.BindingClient.Validate(o.flags)
}

func (o *CreateBindingOptions) Run(ctx context.Context) error {
	service, err := o.clientset.BindingClient.SelectServiceInstance(o.flags)
	if err != nil {
		return err
	}
	bindingName, err := o.clientset.BindingClient.AskBindingName(o.EnvSpecificInfo.GetName(), o.flags)
	if err != nil {
		return err
	}
	bindAsFiles, err := o.clientset.BindingClient.AskBindAsFiles(o.flags)
	if err != nil {
		return err
	}
	return o.clientset.BindingClient.CreateBinding(service, bindingName, bindAsFiles, o.EnvSpecificInfo.GetDevfileObj())

}

// NewCmdBinding implements the component odo sub-command
func NewCmdBinding(name, fullName string) *cobra.Command {
	o := NewCreateBindingOptions()

	var bindingCmd = &cobra.Command{
		Use:     name,
		Short:   "Create ServiceBinding",
		Long:    "Bind a new service to the component with ServiceBinding",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(createBindingExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bindingCmd.Flags().StringVar(&o.name, "name", "", "Name of the Binding to create")
	bindingCmd.Flags().StringVar(&o.service, "service", "", "Name of the service to bind")
	bindingCmd.Flags().BoolVarP(&o.bindAsFiles, "bind-as-files", "", false, "Create the ServiceBinding as a file")
	clientset.Add(bindingCmd, clientset.KUBERNETES, clientset.BINDING)

	return bindingCmd
}
