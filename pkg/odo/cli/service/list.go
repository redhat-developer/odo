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

// NewCmdServiceList implements the odo service list command.
func NewCmdServiceList(name, fullName string) *cobra.Command {
	serviceListCmd := &cobra.Command{
		Use:     name,
		Short:   "List all services in the current application",
		Long:    "List all services in the current application",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			context := genericclioptions.NewContext(cmd)
			client := context.Client
			applicationName := context.Application
			services, err := svc.List(client, applicationName)
			util.CheckError(err, "Service Catalog is not enabled in your cluster")
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
		},
	}
	addProjectFlag(serviceListCmd)
	genericclioptions.AddApplicationFlag(serviceListCmd)
	return serviceListCmd
}
