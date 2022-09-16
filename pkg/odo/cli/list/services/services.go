package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	"os"
	"strings"
)

const RecommendedCommandName = "services"

var (
	listExample = ktemplates.Examples(`
	# List the bindable Operator backed services from current namespace
    %[1]s

	# List all the bindable Operator backed services from all the namespaces
	%[1]s --all-namespaces
	%[1]s -A

	# List the bindable Operator backed services in JSON format
	%[1]s -o json
	%[1]s --all-namespaces -o json
	%[1]s -A -o json`)

	listLongDesc = ktemplates.LongDesc(`
	List the bindable Operator backed services that could be bound to the odo component 
`)
)

type ServiceListOptions struct {
	// clientset
	clientset *clientset.Clientset
	// context
	*genericclioptions.Context
	// working directory
	contextDir string
	// flags
	namespaceFlag     string
	allNamespacesFlag bool
}

var _ genericclioptions.Runnable = (*ServiceListOptions)(nil)
var _ genericclioptions.JsonOutputter = (*ServiceListOptions)(nil)

func (o *ServiceListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ServiceListOptions) Complete(cmdline cmdline.Cmdline, _ []string) error {
	var err error
	o.contextDir, err = os.Getwd()
	if err != nil {
		return err
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}

	if o.namespaceFlag == "" && !o.allNamespacesFlag {
		o.namespaceFlag = o.GetProject()
	}

	return err
}

func (o *ServiceListOptions) Validate() error {
	if o.allNamespacesFlag && o.namespaceFlag != "" {
		return errors.New("cannot use --all-namespaces and --namespace flags together")
	}
	return nil
}

func (o *ServiceListOptions) Run(_ context.Context) error {
	var s *log.Status
	if o.allNamespacesFlag {
		s = log.Spinner("Listing bindable services from all namespaces")
		defer s.End(false)
	} else {
		s = log.Spinner(fmt.Sprintf("Listing bindable services from namespace %q", o.namespaceFlag))
		defer s.End(false)
	}
	services, err := o.run()
	if err != nil {
		return err
	}
	s.End(true)
	HumanReadable(services)
	return nil
}

func (o *ServiceListOptions) run() (api.ResourcesList, error) {
	services := []unstructured.Unstructured{}

	if o.allNamespacesFlag {
		projects, err := o.clientset.ProjectClient.List()
		if err != nil {
			return api.ResourcesList{}, err
		}
		for _, project := range projects.Items {
			svcs, err := o.clientset.BindingClient.GetServiceInstances(project.Name)
			if err != nil {
				return api.ResourcesList{}, err
			}
			for k := range svcs {
				services = append(services, svcs[k])
			}
		}
	} else {
		svcs, err := o.clientset.BindingClient.GetServiceInstances(o.namespaceFlag)
		if err != nil {
			return api.ResourcesList{}, err
		}
		for k := range svcs {
			services = append(services, svcs[k])
		}
	}

	var servicesList []api.BindableService
	for _, svc := range services {
		s := api.BindableService{
			Name:      svc.GetName(),
			Namespace: svc.GetNamespace(),
			Kind:      svc.GetKind(),
			Group:     strings.Split(svc.GetAPIVersion(), "/")[0],
		}
		s.Service = fmt.Sprintf("%s/%s.%s", s.Name, s.Kind, s.Group)
		servicesList = append(servicesList, s)

	}
	return api.ResourcesList{BindableServices: servicesList}, nil
}

func (o *ServiceListOptions) RunForJsonOutput(_ context.Context) (out interface{}, err error) {
	return o.run()
}

func HumanReadable(services api.ResourcesList) {
	if len(services.BindableServices) == 0 {
		log.Error("no bindable Operator backed services found")
		return
	}
	fmt.Println()
	t := ui.NewTable()
	t.AppendHeader(table.Row{"NAME", "NAMESPACE"})
	for _, svc := range services.BindableServices {
		t.AppendRow(table.Row{svc.Service, svc.Namespace})
	}
	t.Render()
}

func NewServicesListOptions() *ServiceListOptions {
	return &ServiceListOptions{}
}

func NewCmdServicesList(name, fullName string) *cobra.Command {
	o := NewServicesListOptions()
	servicesListCmd := &cobra.Command{
		Use:     name,
		Short:   listLongDesc,
		Long:    listLongDesc,
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Aliases: []string{"service"},
	}
	clientset.Add(servicesListCmd, clientset.PROJECT, clientset.BINDING)
	servicesListCmd.Flags().BoolVarP(&o.allNamespacesFlag, "all-namespaces", "A", false, "Show bindable services from all namespaces")
	servicesListCmd.Flags().StringVarP(&o.namespaceFlag, "namespace", "n", "", "Show bindable services from a specific namespace (uses current namespace in kubeconfig by default)")
	machineoutput.UsedByCommand(servicesListCmd)
	return servicesListCmd
}
