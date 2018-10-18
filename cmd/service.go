package cmd

import (
	"fmt"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/catalog/ui"
	"github.com/redhat-developer/odo/pkg/occlient"
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

		var uiClass ui.ServiceClass
		var serviceType string
		if len(args) == 0 {
			uiClass, serviceType = selectClassInteractively(client)
		} else {
			serviceType = args[0]

			// make sure the class exists
			class, err := client.GetServiceClass(serviceType)
			checkError(err, "unable to create service because Service Catalog is not enabled in your cluster")
			if class == nil {
				glog.V(4).Infof("Unknown service class %s", serviceType)
				uiClass, serviceType = selectClassInteractively(client)
			}

			uiClass = ui.ConvertToUI(*class)
		}

		plans, _ := client.GetMatchingPlans(uiClass.Class)

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
				plan = selectPlanNameInteractively(plans, "Which service plan should we use ")
				svcPlan = plans[plan]
			}
		} else {
			var ok bool
			svcPlan, ok = plans[plan]
			if !ok {
				plan = selectPlanNameInteractively(plans, fmt.Sprintf("Unknown plan '%s'. Here are the valid options ", plan))
				svcPlan = plans[plan]
			}
		}

		properties, _ := ui.GetProperties(svcPlan)

		var i = 0
		values := make(map[string]string)
		passedValues := util.ConvertKeyValueStringToMap(parameters)
		for i < len(properties) && properties[i].Required {
			prop := properties[i]
			if _, ok := passedValues[prop.Name]; !ok {
				prompt := promptui.Prompt{
					Label:     fmt.Sprintf("Enter a value for %s property %s ", prop.Type, prop.Title),
					AllowEdit: true,
				}

				result, _ := prompt.Run()
				values[prop.Name] = result
			}

			i++
		}
		// if we have non-required properties, ask if user wants to provide values
		if i < len(properties)-1 {
			// todo
		}

		serviceName := serviceType
		if len(args) == 2 {
			serviceName = args[1]
		} else {
			serviceName = selectServiceNameInteractively(serviceType, "How should we name your service ")
		}

		// check if the service we're trying to create doesn't already exist
		exists, err := svc.SvcExists(client, serviceName, applicationName, projectName)
		checkError(err, "")
		if exists {
			fmt.Printf("%s service already exists in the current application.\n", serviceName)
			selectServiceNameInteractively("", "Select a new name for your service ")
		}

		err = svc.CreateService(client, serviceName, serviceType, plan, values, applicationName)
		checkError(err, "")
		fmt.Printf("Service '%s' was created.\n", serviceName)
	},
}

func selectPlanNameInteractively(plans map[string]scv1beta1.ClusterServicePlan, promptLabel string) string {
	prompt := promptui.Select{
		Label: promptLabel,
		Items: ui.GetServicePlanNames(plans),
	}
	_, plan, _ = prompt.Run()
	return plan
}

func selectServiceNameInteractively(defaultValue, promptLabel string) string {
	// if only one arg is given, ask to name the service providing the class name as default
	instancePrompt := promptui.Prompt{
		Label:     promptLabel,
		Default:   defaultValue,
		AllowEdit: true,
		Validate:  validateName,
	}
	serviceName, _ := instancePrompt.Run()
	return serviceName
}

func selectClassInteractively(client *occlient.Client) (uiClass ui.ServiceClass, serviceType string) {
	classesByCategory, _ := client.GetServiceClassesByCategory()
	prompt := promptui.Select{
		Label: "Which kind of service do you wish to create?",
		Items: ui.GetServiceClassesCategories(classesByCategory),
	}
	_, category, _ := prompt.Run()
	templates := &promptui.SelectTemplates{
		Active:   "\U00002620 {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\U00002620 {{ .Name | red | cyan }}",
		Details: `
			--------- Service Class ----------
			{{ "Name:" | faint }}	{{ .Name }}
			{{ "Description:" | faint }}	{{ .Description }}
			{{ "Long:" | faint }}	{{ .LongDescription }}`,
	}
	uiClasses := ui.GetUIServiceClasses(classesByCategory[category])
	prompt = promptui.Select{
		Label:     "Which " + category + " service class should we use?",
		Items:     uiClasses,
		Templates: templates,
	}
	i, _, _ := prompt.Run()
	uiClass = uiClasses[i]
	serviceType = uiClass.Name

	return uiClass, serviceType
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

func getParametersAsMap(params []string) (parameters map[string]string, err error) {
	parameters = make(map[string]string, len(params))
	for _, value := range params {
		equals := strings.IndexRune(value, '=')
		if equals > 0 {
			split := strings.Split(value, "=")
			parameters[split[0]] = split[1]
		} else {
			return parameters, errors.Errorf("Invalid parameter, must follow 'name=value' format")
		}
	}
	return parameters, nil
}
