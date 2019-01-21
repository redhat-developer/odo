package describe

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"strings"
)

const serviceRecommendedCommandName = "service"

var (
	serviceExample = ktemplates.Examples(`  # Describe a service
    %[1]s mysql-persistent`)

	serviceLongDesc = ktemplates.LongDesc(`Describe a service type.

This describes the service and the associated plans.
`)
)

func NewCmdCatalogDescribeService(name, fullName string) *cobra.Command {
	command := &cobra.Command{
		Use:     name,
		Short:   "Describe a service",
		Long:    serviceLongDesc,
		Example: fmt.Sprintf(serviceExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := genericclioptions.Client(cmd)
			serviceName := args[0]
			service, plans, err := svc.GetServiceClassAndPlans(client, serviceName)
			odoutil.LogErrorAndExit(err, "")

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorder(false)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			serviceData := [][]string{
				{"Name", service.Name},
				{"Bindable", fmt.Sprint(service.Bindable)},
				{"Operated by the broker", service.ServiceBrokerName},
				{"Short Description", service.ShortDescription},
				{"Long Description", service.LongDescription},
				{"Versions Available", strings.Join(service.VersionsAvailable, ",")},
				{"Tags", strings.Join(service.Tags, ",")},
			}

			table.AppendBulk(serviceData)

			table.Append([]string{""})

			if len(plans) > 0 {
				table.Append([]string{"PLANS"})

				for _, plan := range plans {

					// create the display values for required  and optional parameters
					requiredWithMandatoryUserInputParameterNames := []string{}
					requiredWithOptionalUserInputParameterNames := []string{}
					optionalParameterDisplay := []string{}
					for _, parameter := range plan.Parameters {
						if parameter.Required {
							// until we have a better solution for displaying the plan data (like a separate table perhaps)
							// this is simplest thing to do
							if parameter.HasDefaultValue {
								requiredWithOptionalUserInputParameterNames = append(
									requiredWithOptionalUserInputParameterNames,
									fmt.Sprintf("%s (default: '%s')", parameter.Name, parameter.Default))
							} else {
								requiredWithMandatoryUserInputParameterNames = append(requiredWithMandatoryUserInputParameterNames, parameter.Name)
							}

						} else {
							optionalParameterDisplay = append(optionalParameterDisplay, parameter.Name)
						}
					}

					table.Append([]string{"***********************", "*****************************************************"})
					planLineSeparator := []string{"-----------------", "-----------------"}

					planData := [][]string{
						{"Name", plan.Name},
						planLineSeparator,
						{"Display Name", plan.DisplayName},
						planLineSeparator,
						{"Short Description", plan.Description},
						planLineSeparator,
						{"Required Params without a default value", strings.Join(requiredWithMandatoryUserInputParameterNames, ", ")},
						planLineSeparator,
						{"Required Params with a default value", strings.Join(requiredWithOptionalUserInputParameterNames, ", ")},
						planLineSeparator,
						{"Optional Params", strings.Join(optionalParameterDisplay, ", ")},
						{"", ""},
					}
					table.AppendBulk(planData)
				}
				table.Render()
			} else {
				log.Errorf("No plans found for service %s", serviceName)
				os.Exit(1)
			}
		},
	}
	return command
}
