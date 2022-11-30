package describe

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// BindingRecommendedCommandName is the recommended binding sub-command name
const BindingRecommendedCommandName = "binding"

var describeBindingExample = ktemplates.Examples(`
# Describe the bindings in the current devfile
%[1]s

# Describe a binding on the cluster
%[1]s --name frontend
`)

type BindingOptions struct {
	// nameFlag of the component to describe, optional
	nameFlag string

	// Clients
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*BindingOptions)(nil)
var _ genericclioptions.JsonOutputter = (*BindingOptions)(nil)

// NewBindingOptions returns new instance of BindingOptions
func NewBindingOptions() *BindingOptions {
	return &BindingOptions{}
}

func (o *BindingOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *BindingOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	if o.nameFlag == "" {
		devfileObj := odocontext.GetDevfileObj(ctx)
		if devfileObj == nil {
			return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
		}
		return nil
	}
	return nil
}

func (o *BindingOptions) Validate(ctx context.Context) (err error) {
	return nil
}

func (o *BindingOptions) Run(ctx context.Context) error {
	if o.nameFlag == "" {
		bindings, err := o.runWithoutName(ctx)
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
	if o.nameFlag == "" {
		return o.runWithoutName(ctx)
	}
	return o.runWithName()
}

func (o *BindingOptions) runWithoutName(ctx context.Context) ([]api.ServiceBinding, error) {
	var (
		workingDir = odocontext.GetWorkingDirectory(ctx)
		devfileObj = odocontext.GetDevfileObj(ctx)
	)

	return o.clientset.BindingClient.GetBindingsFromDevfile(*devfileObj, workingDir)
}

func (o *BindingOptions) runWithName() (api.ServiceBinding, error) {
	return o.clientset.BindingClient.GetBindingFromCluster(o.nameFlag)
}

// NewCmdBinding implements the binding odo sub-command
func NewCmdBinding(name, fullName string) *cobra.Command {
	o := NewBindingOptions()

	var bindingCmd = &cobra.Command{
		Use:     name,
		Short:   "Describe bindings",
		Long:    "Describe bindings",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(describeBindingExample, fullName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	bindingCmd.Flags().StringVar(&o.nameFlag, "name", "", "Name of the binding to describe, optional. By default, the bindings in the local devfile are described")
	clientset.Add(bindingCmd, clientset.KUBERNETES, clientset.BINDING, clientset.FILESYSTEM)
	commonflags.UseOutputFlag(bindingCmd)

	return bindingCmd
}

// printSingleBindingHumanReadableOutput prints information about a binding and returns true if status is unknown
func printSingleBindingHumanReadableOutput(binding api.ServiceBinding) bool {
	log.Describef("Service Binding Name: ", binding.Name)
	log.Info("Services:")
	for _, service := range binding.Spec.Services {
		gvk := schema.FromAPIVersionAndKind(service.APIVersion, service.Kind)
		if service.Namespace != "" {
			log.Printf("%s (%s.%s) (namespace: %s)", service.Name, gvk.Kind, gvk.Group, service.Namespace)
		} else {
			log.Printf("%s (%s.%s)", service.Name, gvk.Kind, gvk.Group)
		}
	}
	log.Describef("Bind as files: ", strconv.FormatBool(binding.Spec.BindAsFiles))
	log.Describef("Detect binding resources: ", strconv.FormatBool(binding.Spec.DetectBindingResources))

	if binding.Spec.NamingStrategy != "" {
		log.Describef("Naming strategy: ", binding.Spec.NamingStrategy)
	}

	if binding.Status == nil {
		log.Describef("Available binding information: ", "unknown")
		return true
	}
	log.Info("Available binding information:")
	for _, info := range binding.Status.BindingFiles {
		log.Printf(info)
	}
	for _, info := range binding.Status.BindingEnvVars {
		log.Printf(info)
	}
	return false
}

func printBindingsHumanReadableOutput(bindings []api.ServiceBinding) {
	if len(bindings) == 0 {
		log.Info("No ServiceBinding used by the current component")
		return
	}

	log.Info("ServiceBinding used by the current component:")
	someStatusUnknown := false
	for _, binding := range bindings {
		fmt.Println()
		statusUnknown := printSingleBindingHumanReadableOutput(binding)
		if statusUnknown {
			someStatusUnknown = true
		}
	}
	if someStatusUnknown {
		fmt.Println()
		log.Info(`Binding information for one or more ServiceBinding is not available because they don't exist on the cluster yet.
Start "odo dev" first to see binding information.`)
	}
}
