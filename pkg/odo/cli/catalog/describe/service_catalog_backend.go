package describe

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	svc "github.com/openshift/odo/pkg/service"
)

type serviceCatalogBackend struct {
}

func NewServiceCatalogBackend() *serviceCatalogBackend {
	return &serviceCatalogBackend{}
}

func (ohb *serviceCatalogBackend) CompleteDescribeService(dso *DescribeServiceOptions, args []string) error {
	dso.serviceName = args[0]
	return nil
}

func (ohb *serviceCatalogBackend) ValidateDescribeService(dso *DescribeServiceOptions) error {
	var err error
	dso.service, dso.plans, err = svc.GetServiceClassAndPlans(dso.Client, dso.serviceName)
	return err
}

func (ohb *serviceCatalogBackend) RunDescribeService(dso *DescribeServiceOptions) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	serviceData := [][]string{
		{"Name", dso.service.Name},
		{"Bindable", fmt.Sprint(dso.service.Bindable)},
		{"Operated by the broker", dso.service.ServiceBrokerName},
		{"Short Description", dso.service.ShortDescription},
		{"Long Description", dso.service.LongDescription},
		{"Versions Available", strings.Join(dso.service.VersionsAvailable, ",")},
		{"Tags", strings.Join(dso.service.Tags, ",")},
	}

	table.AppendBulk(serviceData)

	table.Append([]string{""})

	if len(dso.plans) > 0 {
		table.Append([]string{"PLANS"})

		for _, plan := range dso.plans {

			// create the display values for required  and optional parameters
			requiredWithMandatoryUserInputParameterNames := []string{}
			requiredWithOptionalUserInputParameterNames := []string{}
			optionalParameterDisplay := []string{}
			for _, parameter := range plan.Parameters {
				if parameter.Required {
					// until we have a better solution for displaying the plan data (like a separate table perhaps)
					// this is simplest thing to do
					if len(parameter.Default) > 0 {
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
		return fmt.Errorf("no plans found for service %s", dso.serviceName)
	}
	return nil
}
