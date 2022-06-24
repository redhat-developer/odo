package binding

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// BindingRecommendedCommandName is the recommended binding sub-command name
const BindingRecommendedCommandName = "binding"

var addBindingExample = ktemplates.Examples(`
# Add binding between a service and the component present in the working directory in the interactive mode
%[1]s

# Add binding between service named 'myservice' and the component present in the working directory
%[1]s --service myservice --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis', and APIGroup 'redis.redis.opstreelab.in' and the component present in the working directory 
%[1]s --service myservice/Redis.redis.redis.opstreelab.in --name myRedisService
%[1]s --service myservice.Redis.redis.redis.opstreelab.in --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis' and the component present in the working directory
%[1]s --service myservice/Redis --name myRedisService

%[1]s --service myservice.Redis --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis' and the deployment app (without Devfile)
%[1]s --service myservice/Redis --name myRedisService --workload app/Deployment.apps
%[1]s --service myservice/Redis --name myRedisService --workload app.Deployment.apps
`)

type AddBindingOptions struct {
	// Flags passed to the command
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
	// The command must work without Devfile
	if err != nil && !genericclioptions.IsNoDevfileError(err) {
		return err
	}

	// this ensures that the namespace as set in env.yaml is used
	o.clientset.KubernetesClient.SetNamespace(o.GetProject())

	o.flags = o.clientset.BindingClient.GetFlags(cmdline.GetFlags())

	return nil
}

func (o *AddBindingOptions) Validate() (err error) {
	withDevfile := o.EnvSpecificInfo.GetDevfileObj().Data != nil
	return o.clientset.BindingClient.ValidateAddBinding(o.flags, withDevfile)
}

func (o *AddBindingOptions) Run(_ context.Context) error {
	withDevfile := o.EnvSpecificInfo.GetDevfileObj().Data != nil

	serviceMap, err := o.clientset.BindingClient.GetServiceInstances()
	if err != nil {
		return err
	}

	if len(serviceMap) == 0 {
		return fmt.Errorf("No bindable service instances found")
	}

	service, err := o.clientset.BindingClient.SelectServiceInstance(o.flags, serviceMap)
	if err != nil {
		return err
	}
	splitService := strings.Split(service, " ")
	serviceName := splitService[0]

	var componentName string
	var workloadName string
	var workloadGVK schema.GroupVersionKind

	if !withDevfile {
		workloadName, workloadGVK, err = o.clientset.BindingClient.SelectWorkloadInstance(o.flags)
		if err != nil {
			return err
		}
		componentName = workloadName
	} else {
		componentName = o.EnvSpecificInfo.GetDevfileObj().GetMetadataName()
	}

	bindingName, err := o.clientset.BindingClient.AskBindingName(serviceName, componentName, o.flags)
	if err != nil {
		return err
	}

	bindAsFiles, err := o.clientset.BindingClient.AskBindAsFiles(o.flags)
	if err != nil {
		return err
	}

	if withDevfile {
		devfileobj, err := o.clientset.BindingClient.AddBindingToDevfile(bindingName, bindAsFiles, serviceMap[service], o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
		err = devfileobj.WriteYamlDevfile()
		if err != nil {
			return err
		}
		log.Success("Successfully added the binding to the devfile.")

		exitMessage := "Run `odo dev` to create it on the cluster."
		if len(o.flags) == 0 {
			kindGroup := strings.ReplaceAll(strings.ReplaceAll(splitService[1], "(", ""), ")", "")
			exitMessage += fmt.Sprintf("\nYou can automate this command by executing:\n  odo add binding --service %s.%s --name %s", serviceName, kindGroup, bindingName)
			if !bindAsFiles {
				exitMessage += " --bind-as-files=false"
			}
		}
		log.Infof(exitMessage)
		return nil
	}

	return o.clientset.BindingClient.AddBinding(o.flags, bindingName, bindAsFiles, serviceMap[service], workloadName, workloadGVK)
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
	bindingCmd.Flags().String(backend.FLAG_NAME, "", "Name of the Binding to create")
	bindingCmd.Flags().String(backend.FLAG_WORKLOAD, "", "Name of the workload to bind, only when no devfile is present in current directory")
	bindingCmd.Flags().String(backend.FLAG_SERVICE, "", "Name of the service to bind")
	bindingCmd.Flags().Bool(backend.FLAG_BIND_AS_FILES, true, "Bind the service as a file")
	clientset.Add(bindingCmd, clientset.BINDING)

	return bindingCmd
}
