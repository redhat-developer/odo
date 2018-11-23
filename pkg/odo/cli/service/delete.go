package service

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"strings"
)

var serviceForceDeleteFlag bool

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`
    # Delete the service named 'mysql-persistent'
    %[1]s mysql-persistent`)
)

// NewCmdServiceDelete implements the odo service delete command.
func NewCmdServiceDelete(name, fullName string) *cobra.Command {
	serviceDeleteCmd := &cobra.Command{
		Use:     name + " <service_name>",
		Short:   "Delete an existing service",
		Long:    "Delete an existing service",
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			glog.V(4).Infof("service delete called\n args: %#v", strings.Join(args, " "))
			context := genericclioptions.NewContext(cmd)
			client := context.Client
			applicationName := context.Application
			serviceName := args[0]
			// Checks to see if the service actually exists
			exists, err := svc.SvcExists(client, serviceName, applicationName)
			util.CheckError(err, "unable to delete service because Service Catalog is not enabled in your cluster")
			if !exists {
				fmt.Printf("Service with the name %s does not exist in the current application\n", serviceName)
				os.Exit(1)
			}
			var confirmDeletion string
			if serviceForceDeleteFlag {
				confirmDeletion = "y"
			} else {
				fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", serviceName, applicationName)
				fmt.Scanln(&confirmDeletion)
			}
			if strings.ToLower(confirmDeletion) == "y" {
				err := svc.DeleteService(client, serviceName, applicationName)
				util.CheckError(err, "")
				fmt.Printf("Service %s from application %s has been deleted\n", serviceName, applicationName)
			} else {
				fmt.Printf("Aborting deletion of service: %v\n", serviceName)
			}
		},
	}
	serviceDeleteCmd.Flags().BoolVarP(&serviceForceDeleteFlag, "force", "f", false, "Delete service without prompting")
	addProjectFlag(serviceDeleteCmd)
	genericclioptions.AddApplicationFlag(serviceDeleteCmd)
	completion.RegisterCommandHandler(serviceDeleteCmd, completion.ServiceCompletionHandler)
	return serviceDeleteCmd
}
