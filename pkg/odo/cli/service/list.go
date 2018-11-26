package service

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"text/tabwriter"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`
    # List all services in the application
    %[1]s`)
)

// ServiceListOptions encapsulates the options for the odo service list command
type ServiceListOptions struct {
	*genericclioptions.Context
}

// NewServiceListOptions creates a new ServiceListOptions instance
func NewServiceListOptions() *ServiceListOptions {
	return &ServiceListOptions{}
}

// Complete completes ServiceListOptions after they've been created
func (o *ServiceListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the ServiceListOptions based on completed values
func (o *ServiceListOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo service list command
func (o *ServiceListOptions) Run() (err error) {
	services, err := svc.ListWithDetailedStatus(o.Client, o.Application)
	if err != nil {
		return fmt.Errorf("service catalog is not enabled in your cluster:\n%v", err)
	}

	if len(services) == 0 {
		fmt.Println("There are no services deployed for this application")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "TYPE", "\t", "STATUS")
	for _, comp := range services {
		fmt.Fprintln(w, comp.Name, "\t", comp.Type, "\t", comp.Status)
	}
	w.Flush()
	return
}

// NewCmdServiceList implements the odo service list command.
func NewCmdServiceList(name, fullName string) *cobra.Command {
	o := NewServiceListOptions()
	serviceListCmd := &cobra.Command{
		Use:     name,
		Short:   "List all services in the current application",
		Long:    "List all services in the current application",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckError(o.Complete(name, cmd, args), "")
			util.CheckError(o.Validate(), "")
			util.CheckError(o.Run(), "")
		},
	}
	addProjectFlag(serviceListCmd)
	genericclioptions.AddApplicationFlag(serviceListCmd)
	return serviceListCmd
}
