package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/glog"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

var (
	serviceForceDeleteFlag bool
	parameters             []string
	plan                   string
)

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
	Use:   "create <service_type> --plan <plan_name> [service_name]",
	Short: "Create a new service",
	Long: `Create a new service from service catalog using the plan defined and deploy it on OpenShift.

If service name is not provided, service type value will be used. The plan to be used must be passed along the service type
using this convention <service_type>/<plan>. The parameters to configure the service are passed as a list of key=value pairs.
The list of the parameters and their type is defined according to the plan selected.

A full list of service types that can be deployed are available using: 'odo catalog list services'`,
	Example: `  # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
  odo service create dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContextCreatingAppIfNeeded(cmd)
		client := context.Client
		applicationName := context.Application

		// make sure the service type exists
		serviceType := args[0]
		matchingService, err := svc.GetSvcByType(client, serviceType)
		util.CheckError(err, "unable to create service because Service Catalog is not enabled in your cluster")
		if matchingService == nil {
			fmt.Printf("Service %v doesn't exist\nRun 'odo service catalog' to see a list of supported services.\n", serviceType)
			os.Exit(1)
		}

		if len(plan) == 0 {
			// when the plan has not been supplied, if there is only one available plan, we select it
			if len(matchingService.PlanList) == 1 {
				plan = matchingService.PlanList[0]
				glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", plan, serviceType)
			} else {
				fmt.Printf("No plan was supplied for service %v.\nPlease select one of: %v\n", serviceType, strings.Join(matchingService.PlanList, ","))
				os.Exit(1)
			}
		} else {
			// when the plan has been supplied, we need to make sure it exists
			planFound := false
			for _, candidatePlan := range matchingService.PlanList {
				if plan == candidatePlan {
					planFound = true
					break
				}
			}
			if !planFound {
				fmt.Printf("Plan %s is invalid for service %v.\nPlease select one of: %v\n", plan, serviceType, strings.Join(matchingService.PlanList, ","))
				os.Exit(1)
			}
		}

		// if only one arg is given, then it is considered as service name and service type both
		serviceName := serviceType
		// if two args are given, first is service type and second one is service name
		if len(args) == 2 {
			serviceName = args[1]
		}
		//validate service name
		err = validateName(serviceName)
		util.CheckError(err, "")
		exists, err := svc.SvcExists(client, serviceName, applicationName)

		util.CheckError(err, "")
		if exists {
			fmt.Printf("%s service already exists in the current application.\n", serviceName)
			os.Exit(1)
		}
		err = svc.CreateService(client, serviceName, serviceType, plan, parameters, applicationName)
		util.CheckError(err, "")
		fmt.Printf(`Service '%s' was created.
Progress of the provisioning will not be reported and might take a long time.
You can see the current status by executing 'odo service list'`, serviceName)
	},
}

var serviceDeleteCmd = &cobra.Command{
	Use:   "delete <service_name>",
	Short: "Delete an existing service",
	Long:  "Delete an existing service",
	Example: `  # Delete the service named 'mysql-persistent'
  odo service delete mysql-persistent
	`,
	Args: cobra.ExactArgs(1),
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

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services in the current application",
	Long:  "List all services in the current application",
	Example: `  # List all services in the application
  odo service list
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		services, err := svc.ListWithDetailedStatus(client, applicationName)
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

func init() {
	serviceDeleteCmd.Flags().BoolVarP(&serviceForceDeleteFlag, "force", "f", false, "Delete service without prompting")
	serviceCreateCmd.Flags().StringVar(&plan, "plan", "", "The name of the plan of the service to be created")
	serviceCreateCmd.Flags().StringSliceVarP(&parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")

	// Add a defined annotation in order to appear in the help menu
	serviceCmd.Annotations = map[string]string{"command": "other"}
	serviceCmd.SetUsageTemplate(cmdUsageTemplate)
	serviceCmd.AddCommand(serviceCreateCmd)
	serviceCmd.AddCommand(serviceDeleteCmd)
	serviceCmd.AddCommand(serviceListCmd)

	//Adding `--project` flag
	addProjectFlag(serviceCreateCmd)
	addProjectFlag(serviceDeleteCmd)
	addProjectFlag(serviceListCmd)

	//Adding `--application` flag
	addApplicationFlag(serviceCreateCmd)
	addApplicationFlag(serviceDeleteCmd)
	addApplicationFlag(serviceListCmd)

	rootCmd.AddCommand(serviceCmd)

	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	completion.RegisterCommandHandler(serviceDeleteCmd, completion.ServiceCompletionHandler)
}
