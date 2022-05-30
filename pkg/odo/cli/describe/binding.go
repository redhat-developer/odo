package describe

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const BindingRecommendedCommandName = "binding"

var describeBindingExample = ktemplates.Examples(`
# Describe the bindings in the current devfile
%[1]s

# Describe a binding in the cluster
%[1]s --name frontend
`)

type BindingOptions struct {
	// nameFlag of the component to describe, optional
	nameFlag string

	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// working directory
	contextDir string
}

// NewComponentOptions returns new instance of ComponentOptions
func NewBindingOptions() *BindingOptions {
	return &BindingOptions{}
}

func (o *BindingOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *BindingOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	if o.nameFlag == "" {
		o.contextDir, err = os.Getwd()
		if err != nil {
			return err
		}

		o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
		if err != nil {
			return err
		}
		// this ensures that the namespace set in env.yaml is used
		o.clientset.KubernetesClient.SetNamespace(o.GetProject())
		return nil
	}
	return nil
}

func (o *BindingOptions) Validate() (err error) {
	return nil
}

func (o *BindingOptions) Run(ctx context.Context) error {
	if len(o.nameFlag) == 0 {
		bindings, err := o.runWithoutName()
		if err != nil {
			return err
		}
		printBindingsHumanReadableOutput(bindings)
		return nil
	}

	binding, err := o.runWithName()
	if err != nil {
		return err
	}
	printSingleBindingHumanReadableOutput(binding)
	return nil
}

// Run contains the logic for the odo command
func (o *BindingOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	if len(o.nameFlag) == 0 {
		return o.runWithoutName()
	}
	return o.runWithName()
}

func (o *BindingOptions) runWithoutName() ([]api.ServiceBinding, error) {
	return o.clientset.BindingClient.GetBindingsFromDevfile(o.EnvSpecificInfo.GetDevfileObj(), o.contextDir)
}

func (o *BindingOptions) runWithName() (api.ServiceBinding, error) {
	return o.clientset.BindingClient.GetBinding(o.nameFlag)
}

// NewCmdComponent implements the component odo sub-command
func NewCmdBinding(name, fullName string) *cobra.Command {
	o := NewBindingOptions()

	var bindingCmd = &cobra.Command{
		Use:     name,
		Short:   "Describe bindings",
		Long:    "Describe bindings",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(describeBindingExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bindingCmd.Flags().StringVar(&o.nameFlag, "name", "", "Name of the binding to describe, optional. By default, the bindings in the local devfile are described")
	clientset.Add(bindingCmd, clientset.KUBERNETES, clientset.BINDING)
	machineoutput.UsedByCommand(bindingCmd)

	return bindingCmd
}

func printSingleBindingHumanReadableOutput(binding api.ServiceBinding) {
	log.Describef("Service Binding Name: ", binding.Name)
	log.Info("Services:")
	for _, service := range binding.Spec.Services {
		gvk := schema.FromAPIVersionAndKind(service.APIVersion, service.Kind)
		log.Printf("%s (%s.%s)", service.Name, gvk.Kind, gvk.Group)
	}
	log.Describef("Bind as files: ", strconv.FormatBool(binding.Spec.BindAsFiles))
	log.Describef("Detect binding resources: ", strconv.FormatBool(binding.Spec.DetectBindingResources))

	if binding.Status == nil {
		log.Describef("Available binding information: ", "unknown")
		return
	}
	log.Info("Available binding information:")
	for _, info := range binding.Status.BindingFiles {
		log.Printf(info)
	}
	for _, info := range binding.Status.BindingEnvVars {
		log.Printf(info)
	}
}

func printBindingsHumanReadableOutput(bindings []api.ServiceBinding) {
	if len(bindings) == 0 {
		log.Info("No ServiceBinding used by the current component")
		return
	}

	log.Info("ServiceBinding used by the current component:")
	for _, binding := range bindings {
		fmt.Println()
		printSingleBindingHumanReadableOutput(binding)
	}
}
