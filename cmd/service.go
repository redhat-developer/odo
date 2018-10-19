package cmd

import (
	"fmt"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/redhat-developer/odo/pkg/catalog/ui"
	"github.com/redhat-developer/odo/pkg/util"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/project"
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
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrentOrGetCreateSetDefault(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		var class scv1beta1.ClusterServiceClass
		var serviceType string
		if len(args) == 0 {
			class, serviceType = ui.SelectClassInteractively(client)
		} else {
			serviceType = args[0]

			// make sure the class exists
			class, err := client.GetServiceClass(serviceType)
			checkError(err, "unable to create service because Service Catalog is not enabled in your cluster")
			if class == nil {
				glog.V(4).Infof("Unknown service class %s", serviceType)
				*class, serviceType = ui.SelectClassInteractively(client)
			}
		}

		plans, _ := client.GetMatchingPlans(class)

		var svcPlan scv1beta1.ClusterServicePlan
		if len(plan) == 0 {
			// when the plan has not been supplied, if there is only one available plan, we select it
			if len(plans) == 1 {
				for k, v := range plans {
					plan = k
					svcPlan = v
				}
				glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", plan, serviceType)
			} else {
				plan = ui.SelectPlanNameInteractively(plans, "Which service plan should we use ")
				svcPlan = plans[plan]
			}
		} else {
			var ok bool
			svcPlan, ok = plans[plan]
			if !ok {
				plan = ui.SelectPlanNameInteractively(plans, fmt.Sprintf("Unknown plan '%s'. Here are the valid options ", plan))
				svcPlan = plans[plan]
			}
		}

		passedValues := util.ConvertKeyValueStringToMap(parameters)
		values := ui.EnterServicePropertiesInteractively(svcPlan, passedValues)

		serviceName := serviceType
		if len(args) == 2 {
			serviceName = args[1]
		} else {
			serviceName = ui.SelectServiceNameInteractively(serviceType, "How should we name your service ", validateName)
		}

		// check if the service we're trying to create doesn't already exist
		exists, err := svc.SvcExists(client, serviceName, applicationName, projectName)
		checkError(err, "")
		if exists {
			fmt.Printf("%s service already exists in the current application.\n", serviceName)
			ui.SelectServiceNameInteractively("", "Select a new name for your service ", validateName)
		}

		err = svc.CreateService(client, serviceName, serviceType, plan, values, applicationName)
		checkError(err, "")
		fmt.Printf("Service '%s' was created.\n", serviceName)
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
			err := svc.DeleteService(client, serviceName, applicationName)
			checkError(err, "")
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
	serviceCreateCmd.Flags().StringVar(&plan, "plan", "", "The name of the plan of the service to be created")
	serviceCreateCmd.Flags().StringSliceVarP(&parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")

	// Add a defined annotation in order to appear in the help menu
	serviceCmd.Annotations = map[string]string{"command": "other"}
	serviceCmd.SetUsageTemplate(cmdUsageTemplate)
	serviceCmd.AddCommand(serviceCreateCmd)
	serviceCmd.AddCommand(serviceDeleteCmd)
	serviceCmd.AddCommand(serviceListCmd)
	rootCmd.AddCommand(serviceCmd)
}
