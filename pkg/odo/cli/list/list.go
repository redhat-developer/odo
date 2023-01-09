package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/binding"
	clicomponent "github.com/redhat-developer/odo/pkg/odo/cli/list/component"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/namespace"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/services"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	dfutil "github.com/devfile/library/v2/pkg/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended list name
const RecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions ...
type ListOptions struct {
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

	// If the namespace flag has been passed, we will search there.
	// if it hasn't, we will search from the default project / namespace.
	if lo.namespaceFlag != "" {
		lo.namespaceFilter = lo.namespaceFlag
	} else if lo.clientset.KubernetesClient != nil {
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
	clicomponent.HumanReadableOutput(ctx, list)
	fmt.Printf("\nBindings:\n")
	binding.HumanReadableOutput(list)
	return nil
}

// Run contains the logic for the odo command
func (lo *ListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return lo.run(ctx)
}

func (lo *ListOptions) run(ctx context.Context) (list api.ResourcesList, err error) {
	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		componentName = odocontext.GetComponentName(ctx)

		kubeClient   = lo.clientset.KubernetesClient
		podmanClient = lo.clientset.PodmanClient
	)

	switch fcontext.GetPlatform(ctx, "") {
	case commonflags.PlatformCluster:
		podmanClient = nil
	case commonflags.PlatformPodman:
		kubeClient = nil
	}

	allComponents, componentInDevfile, err := component.ListAllComponents(
		kubeClient, podmanClient, lo.namespaceFilter, devfileObj, componentName)
	if err != nil {
		return api.ResourcesList{}, err
	}

	var bindings []api.ServiceBinding
	var inDevfile []string

	workingDir := odocontext.GetWorkingDirectory(ctx)
	bindings, inDevfile, err = lo.clientset.BindingClient.ListAllBindings(devfileObj, workingDir)
	if err != nil {
		return api.ResourcesList{}, err
	}

	// RunningOn is displayed only when Platform is active
	if !feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		for i := range allComponents {
			//lint:ignore SA1019 we need to output the deprecated value, before to remove it in a future release
			allComponents[i].RunningOn = ""
			allComponents[i].Platform = ""
		}
	}

	return api.ResourcesList{
		ComponentInDevfile: componentInDevfile,
		Components:         allComponents,
		BindingsInDevfile:  inDevfile,
		Bindings:           bindings,
	}, nil
}

// NewCmdList implements the list odo command
func NewCmdList(ctx context.Context, name, fullName string) *cobra.Command {
	o := NewListOptions()

	var listCmd = &cobra.Command{
		Use:     name,
		Short:   "List all components in the current namespace",
		Long:    "List all components in the current namespace.",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(listCmd, clientset.KUBERNETES_NULLABLE, clientset.BINDING, clientset.FILESYSTEM)
	if feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		clientset.Add(listCmd, clientset.PODMAN_NULLABLE)
	}

	namespaceCmd := namespace.NewCmdNamespaceList(namespace.RecommendedCommandName, odoutil.GetFullName(fullName, namespace.RecommendedCommandName))
	bindingCmd := binding.NewCmdBindingList(binding.RecommendedCommandName, odoutil.GetFullName(fullName, binding.RecommendedCommandName))
	componentCmd := clicomponent.NewCmdComponentList(ctx, clicomponent.RecommendedCommandName, odoutil.GetFullName(fullName, clicomponent.RecommendedCommandName))
	servicesCmd := services.NewCmdServicesList(services.RecommendedCommandName, odoutil.GetFullName(fullName, services.RecommendedCommandName))
	listCmd.AddCommand(namespaceCmd, bindingCmd, componentCmd, servicesCmd)

	util.SetCommandGroup(listCmd, util.ManagementGroup)
	listCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	listCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for components")

	commonflags.UseOutputFlag(listCmd)
	commonflags.UsePlatformFlag(listCmd)

	return listCmd
}
