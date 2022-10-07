package binding

import (
	"context"
	"fmt"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/binding/asker"
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

var _ genericclioptions.Runnable = (*AddBindingOptions)(nil)

// NewAddBindingOptions returns new instance of ComponentOptions
func NewAddBindingOptions() *AddBindingOptions {
	return &AddBindingOptions{}
}

func (o *AddBindingOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *AddBindingOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	// The command must work without Devfile
	if err != nil && !genericclioptions.IsNoDevfileError(err) {
		return err
	}

	o.flags = o.clientset.BindingClient.GetFlags(cmdline.GetFlags())

	return nil
}

func (o *AddBindingOptions) Validate(ctx context.Context) (err error) {
	withDevfile := o.DevfileObj.Data != nil
	return o.clientset.BindingClient.ValidateAddBinding(o.flags, withDevfile)
}

func (o *AddBindingOptions) Run(_ context.Context) error {
	withDevfile := o.DevfileObj.Data != nil

	ns, err := o.clientset.BindingClient.SelectNamespace(o.flags)
	if err != nil {
		return err
	}

	serviceMap, err := o.clientset.BindingClient.GetServiceInstances(ns)
	if err != nil {
		return err
	}

	if len(serviceMap) == 0 {
		msg := "current namespace"
		if ns != "" {
			msg = fmt.Sprintf("namespace %q", ns)
		}
		return fmt.Errorf("No bindable service instances found in %s", msg)
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
		componentName = o.GetComponentName()
	}

	bindingName, err := o.clientset.BindingClient.AskBindingName(serviceName, componentName, o.flags)
	if err != nil {
		return err
	}

	bindAsFiles, err := o.clientset.BindingClient.AskBindAsFiles(o.flags)
	if err != nil {
		return err
	}

	namingStrategy, err := o.clientset.BindingClient.AskNamingStrategy(o.flags)
	if err != nil {
		return err
	}

	if withDevfile {
		var devfileobj parser.DevfileObj
		devfileobj, err = o.clientset.BindingClient.AddBindingToDevfile(
			componentName, bindingName, bindAsFiles, ns, namingStrategy, serviceMap[service], o.DevfileObj)
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
			if ns != "" {
				exitMessage += fmt.Sprintf(" --service-namespace %s", ns)
			}
			if !bindAsFiles {
				exitMessage += " --bind-as-files=false"
			}
			if namingStrategy != "" {
				exitMessage += " --naming-strategy='" + namingStrategy + "'"
			}
		}
		log.Infof(exitMessage)
		return nil
	}

	options, output, filename, err := o.clientset.BindingClient.AddBinding(
		o.flags, bindingName, bindAsFiles, ns, namingStrategy, serviceMap[service], workloadName, workloadGVK)
	if err != nil {
		return err
	}

	if output != "" {
		fmt.Println(output)
	}

	// Display the info after outputting to stdout
	for _, option := range options {
		switch option {
		case asker.OutputToFile:
			log.Infof("The ServiceBinding has been written to the file %q", filename)

		case asker.CreateOnCluster:
			log.Infof("The ServiceBinding has been created in the cluster")
		}
	}

	return nil
}

// NewCmdBinding implements the component odo sub-command
func NewCmdBinding(name, fullName string) *cobra.Command {
	o := NewAddBindingOptions()

	var bindingCmd = &cobra.Command{
		Use:     name,
		Short:   "Add Binding",
		Long:    "Add a binding between a service and the component in the devfile",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(addBindingExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bindingCmd.Flags().String(backend.FLAG_NAME, "", "Name of the Binding to create")
	bindingCmd.Flags().String(backend.FLAG_WORKLOAD, "", "Name of the workload to bind, only when no devfile is present in current directory")
	bindingCmd.Flags().String(backend.FLAG_SERVICE, "", "Name of the service to bind")
	bindingCmd.Flags().String(backend.FLAG_SERVICE_NAMESPACE, "", "Namespace of the service to bind to. Default is the component namespace.")
	bindingCmd.Flags().Bool(backend.FLAG_BIND_AS_FILES, true, "Bind the service as a file")
	bindingCmd.Flags().String(backend.FLAG_NAMING_STRATEGY, "",
		"Naming strategy to use for binding names. "+
			"It can be set to pre-defined strategies: 'none', 'lowercase', or 'uppercase'. "+
			"Otherwise, it is treated as a custom Go template, and it is handled accordingly.")
	clientset.Add(bindingCmd, clientset.BINDING)

	return bindingCmd
}
