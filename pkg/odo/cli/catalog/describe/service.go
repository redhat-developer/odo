package describe

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	svc "github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const serviceRecommendedCommandName = "service"

var (
	serviceExample = ktemplates.Examples(`  # Describe a service catalog service
    %[1]s mysql-persistent
	
	# Describe a Operator backed service
	%[1]s 
	`)

	serviceLongDesc = ktemplates.LongDesc(`Describes a service type.
	This command supports both Service Catalog services and Operator backed services.
	A user can describe an Operator backed service by providing the full identifier for an Operand i.e. <operator_type>/<cr_name> which they can find by running "odo catalog list services".

	If the format doesn't match <operator_type>/<cr_name> then service catalog services would be searched.  

`)
)

// DescribeServiceOptions encapsulates the options for the odo catalog describe service command
type DescribeServiceOptions struct {
	// name of the service to describe, from command arguments
	serviceName string
	// resolved service
	service svc.ServiceClass
	plans   []svc.ServicePlan
	// generic context options common to all commands
	*genericclioptions.Context

	// Operator backed services params
	// split the name provided on CLI and populate servicetype & customresource
	isOperator bool
	// the operator name
	OperatorType   string
	CustomResource string
	CSV            olm.ClusterServiceVersion
	CR             *olm.CRDDescription
}

// NewDescribeServiceOptions creates a new DescribeServiceOptions instance
func NewDescribeServiceOptions() *DescribeServiceOptions {
	return &DescribeServiceOptions{}
}

// Complete completes DescribeServiceOptions after they've been created
func (o *DescribeServiceOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	// if the argument contains "/" then we assume the user wants to describe a CRD.
	if strings.Contains(args[0], "/") {
		// we check if the cluster supports ClusterServiceVersion or not.
		isCSVSupported, err := svc.IsCSVSupported()
		if err != nil {
			// if there is an error checking it, we return the error.
			return err
		}
		// if its not supported then we return an error
		if !isCSVSupported {
			return errors.New("it seems the cluster doesn't support Operators. Please install OLM and try again")
		}
		tmpOptrList := strings.Split(args[0], "/")
		o.OperatorType = tmpOptrList[0]
		o.CustomResource = tmpOptrList[1]
		o.isOperator = true
	} else {
		o.serviceName = args[0]
	}

	o.Context, err = genericclioptions.NewContext(cmd, true)

	return
}

// Validate validates the DescribeServiceOptions based on completed values
func (o *DescribeServiceOptions) Validate() (err error) {

	if o.isOperator {

		if o.OperatorType == "" || o.CustomResource == "" {
			return fmt.Errorf("invalid service name, use the format <operator-type>/<crd-name>")
		}
		// make sure that CSV of the specified OperatorType exists
		o.CSV, err = o.KClient.GetClusterServiceVersion(o.OperatorType)
		if err != nil {
			// error only occurs when OperatorHub is not installed.
			// k8s does't have it installed by default but OCP does
			return err
		}

		// Get the specific CR that matches "kind"
		crs := o.KClient.GetCustomResourcesFromCSV(&o.CSV)

		var cr *olm.CRDDescription
		hasCR := false
		for _, custRes := range *crs {
			c := custRes
			if c.Kind == o.CustomResource {
				cr = &c
				hasCR = true
				break
			}
		}
		if !hasCR {
			return fmt.Errorf("the %s resource doesn't exist in specified %s operator", o.CustomResource, o.OperatorType)
		}

		o.CR = cr
		return nil
	}
	o.service, o.plans, err = svc.GetServiceClassAndPlans(o.Client, o.serviceName)
	return err

}

// Run contains the logic for the command associated with DescribeServiceOptions
func (o *DescribeServiceOptions) Run() (err error) {
	if o.isOperator {
		return o.operatorRun()
	}
	return o.serviceCatalogRun()

}
func (o *DescribeServiceOptions) operatorRun() (err error) {
	if log.IsJSON() {
		machineoutput.OutputSuccess(o.CR)
		return
	}
	output, err := yaml.Marshal(svc.ConvertCRDToRepr(o.CR))
	if err != nil {
		return err
	}

	fmt.Print(string(output))
	return nil
}

func (o *DescribeServiceOptions) serviceCatalogRun() (err error) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	serviceData := [][]string{
		{"Name", o.service.Name},
		{"Bindable", fmt.Sprint(o.service.Bindable)},
		{"Operated by the broker", o.service.ServiceBrokerName},
		{"Short Description", o.service.ShortDescription},
		{"Long Description", o.service.LongDescription},
		{"Versions Available", strings.Join(o.service.VersionsAvailable, ",")},
		{"Tags", strings.Join(o.service.Tags, ",")},
	}

	table.AppendBulk(serviceData)

	table.Append([]string{""})

	if len(o.plans) > 0 {
		table.Append([]string{"PLANS"})

		for _, plan := range o.plans {

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
		return fmt.Errorf("no plans found for service %s", o.serviceName)
	}
	return
}

// NewCmdCatalogDescribeService implements the odo catalog describe service command
func NewCmdCatalogDescribeService(name, fullName string) *cobra.Command {
	o := NewDescribeServiceOptions()
	command := &cobra.Command{
		Use:         name,
		Short:       "Describe a service",
		Long:        serviceLongDesc,
		Example:     fmt.Sprintf(serviceExample, fullName),
		Annotations: map[string]string{"machineoutput": "json"},
		Args:        cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return command
}
