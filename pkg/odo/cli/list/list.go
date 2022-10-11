package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cli/list/services"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/binding"
	clicomponent "github.com/redhat-developer/odo/pkg/odo/cli/list/component"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/namespace"

	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended list name
const RecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions ...
type ListOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Local variables
	namespaceFilter string

	// Flags
	namespaceFlag string
}

var _ genericclioptions.Runnable = (*ListOptions)(nil)
var _ genericclioptions.JsonOutputter = (*ListOptions)(nil)

// NewListOptions ...
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (o *ListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete ...
func (lo *ListOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {

	// Check to see if KUBECONFIG exists, and if not, error the user that we would not be able to get cluster information
	// Do this before anything else, or else we will just error out with the:
	// invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable
	// instead
	if !dfutil.CheckKubeConfigExist() {
		return errors.New("KUBECONFIG not found. Unable to retrieve cluster information. Please set your Kubernetes configuration via KUBECONFIG env variable or ~/.kube/config")
	}

	// Create the local context and initial Kubernetes client configuration
	lo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	// The command must work without Devfile
	if err != nil && !genericclioptions.IsNoDevfileError(err) {
		return err
	}

	// If the namespace flag has been passed, we will search there.
	// if it hasn't, we will search from the default project / namespace.
	if lo.namespaceFlag != "" {
		lo.namespaceFilter = lo.namespaceFlag
	} else {
		lo.namespaceFilter = odocontext.GetNamespace(ctx)
	}

	return nil
}

// Validate ...
func (lo *ListOptions) Validate(ctx context.Context) (err error) {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run(ctx context.Context) error {
	listSpinner := log.Spinnerf("Listing resources from the namespace %q", lo.namespaceFilter)
	defer listSpinner.End(false)

	list, err := lo.run(ctx)
	if err != nil {
		return err
	}

	listSpinner.End(true)

	fmt.Printf("\nComponents:\n")
	clicomponent.HumanReadableOutput(list)
	fmt.Printf("\nBindings:\n")
	binding.HumanReadableOutput(lo.namespaceFilter, list)
	return nil
}

// Run contains the logic for the odo command
func (lo *ListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return lo.run(ctx)
}

func (lo *ListOptions) run(ctx context.Context) (list api.ResourcesList, err error) {
	devfileComponents, componentInDevfile, err := component.ListAllComponents(
		lo.clientset.KubernetesClient, lo.namespaceFilter, lo.DevfileObj, lo.GetComponentName())
	if err != nil {
		return api.ResourcesList{}, err
	}

	workingDir := odocontext.GetWorkingDirectory(ctx)
	bindings, inDevfile, err := lo.clientset.BindingClient.ListAllBindings(&lo.DevfileObj, workingDir)
	if err != nil {
		return api.ResourcesList{}, err
	}

	return api.ResourcesList{
		ComponentInDevfile: componentInDevfile,
		Components:         devfileComponents,
		BindingsInDevfile:  inDevfile,
		Bindings:           bindings,
	}, nil
}

// NewCmdList implements the list odo command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var listCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components in the current namespace",
		Long:        "List all components in the current namespace.",
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        genericclioptions.NoArgsAndSilenceJSON,
		Annotations: map[string]string{"command": "management"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(listCmd, clientset.KUBERNETES, clientset.BINDING, clientset.FILESYSTEM)

	namespaceCmd := namespace.NewCmdNamespaceList(namespace.RecommendedCommandName, odoutil.GetFullName(fullName, namespace.RecommendedCommandName))
	bindingCmd := binding.NewCmdBindingList(binding.RecommendedCommandName, odoutil.GetFullName(fullName, binding.RecommendedCommandName))
	componentCmd := clicomponent.NewCmdComponentList(clicomponent.RecommendedCommandName, odoutil.GetFullName(fullName, clicomponent.RecommendedCommandName))
	servicesCmd := services.NewCmdServicesList(services.RecommendedCommandName, odoutil.GetFullName(fullName, services.RecommendedCommandName))
	listCmd.AddCommand(namespaceCmd, bindingCmd, componentCmd, servicesCmd)

	listCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	listCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for components")

	completion.RegisterCommandFlagHandler(listCmd, "path", completion.FileCompletionHandler)
	commonflags.UseOutputFlag(listCmd)

	return listCmd
}
