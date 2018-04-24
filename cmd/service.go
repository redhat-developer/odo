package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/project"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

var serviceForceDeleteFlag bool

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Perform service catalog operations",
	Long:  ` Perform service catalog operations, Limited to template service broker only.`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		serviceCreateCmd.Example,
		serviceDeleteCmd.Example,
		serviceListCmd.Example),
	Args: cobra.RangeArgs(1, 3),
}

var serviceCreateCmd = &cobra.Command{
	Use:   "create <service_type> [service_name]",
	Short: "Create a new service",
	Long: `Create a new service from service catalog to deploy on OpenShift.

If service name is not provided, service type value will be used.

A full list of service types that can be deployed is available using: 'odo service catalog'`,
	Example: `  # Create new mysql-persistent service from service catalog.
  odo service create mysql-persistent
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrentOrGetCreateSetDefault(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		serviceType := args[0]
		exists, err := svc.SvcTypeExists(client, serviceType)
		checkError(err, "unable to create service because Service Catalog is not enabled in your cluster")
		if !exists {
			fmt.Printf("Service %v doesn't exist\nRun 'odo service catalog' to see a list of supported services.\n", serviceType)
			os.Exit(1)
		}
		// if only one arg is given, then it is considered as service name and service type both
		serviceName := args[0]
		// if two args are given, first is service type and second one is service name
		if len(args) == 2 {
			serviceName = args[1]
		}
		//validate service name
		err = validateName(serviceName)
		checkError(err, "")
		exists, err = svc.SvcExists(client, serviceName, applicationName, projectName)
		checkError(err, "")
		if exists {
			fmt.Printf("Service with the name %s already exists in the current application.\n", serviceName)
			os.Exit(1)
		}
		err = svc.CreateService(client, serviceName, serviceType, applicationName)
		checkError(err, "")
		fmt.Printf("Service '%s' was created.\n", serviceName)
	},
}

var serviceDeleteCmd = &cobra.Command{
	Use:   "delete <service_name>",
	Short: "Delete an existing service",
	Long:  "Delete an existing service",
	Example: `  # Delete service named 'mysql-persistent'
  odo service delete mysql-persistent
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		glog.V(4).Infof("service delete called")
		glog.V(4).Infof("args: %#v", strings.Join(args, " "))
		client := getOcClient()

		// Get all necessary names (current application + project)
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		serviceName := args[0]

		// Checks to see if the service actually exists
		exists, err := svc.SvcExists(client, serviceName, applicationName, projectName)
		checkError(err, "unable to delete service because Service Catalog is not enabled in your cluster")
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
			err := svc.DeleteService(client, serviceName, applicationName, projectName)
			checkError(err, "")
			fmt.Printf("Service %s from application %s has been deleted\n", serviceName, applicationName)

		} else {
			fmt.Printf("Aborting deletion of service: %v\n", serviceName)
		}
	},
}

//var serviceCatalogCmd = &cobra.Command{
//	Use:   "catalog",
//	Short: "Lists all the services from service catalog",
//	Long:  "Lists all the services from service catalog",
//	Example: `  # List all services
//  odo service catalog
//	`,
//	Args: cobra.ExactArgs(0),
//	Run: func(cmd *cobra.Command, args []string) {
//		client := getOcClient()
//		catalogList, err := svc.ListCatalog(client)
//		checkError(err, "unable to list services because Service Catalog is not enabled in your cluster")
//		switch len(catalogList) {
//		case 0:
//			fmt.Printf("No deployable services found\n")
//		default:
//			fmt.Println("The following services can be deployed:")
//			for _, service := range catalogList {
//				fmt.Printf("- %v\n", service)
//			}
//		}
//	},
//}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the current application",
	Long:  "List all services in the current application",
	Example: `  # List all services in the application
  odo service list
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		services, err := svc.List(client, applicationName, projectName)
		checkError(err, "Service Catalog is not enabled in your cluster")

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

func init() {
	serviceDeleteCmd.Flags().BoolVarP(&serviceForceDeleteFlag, "force", "f", false, "Delete service without prompting")

	// Add a defined annotation in order to appear in the help menu
	serviceCmd.Annotations = map[string]string{"command": "other"}
	serviceCmd.SetUsageTemplate(cmdUsageTemplate)
	serviceCmd.AddCommand(serviceCreateCmd)
	serviceCmd.AddCommand(serviceDeleteCmd)
	//serviceCmd.AddCommand(serviceCatalogCmd)
	serviceCmd.AddCommand(serviceListCmd)
	rootCmd.AddCommand(serviceCmd)
}
