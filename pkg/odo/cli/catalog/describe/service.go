package describe

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const serviceRecommendedCommandName = "service"

var (
	serviceExample = ktemplates.Examples(`# Describe a Operator backed service
	%[1]s 
	`)

	serviceLongDesc = ktemplates.LongDesc(`Describes a service type.
	This command supports Operator backed services.
	A user can describe an Operator backed service by providing the full identifier for an Operand i.e. <operator_type>/<cr_name> which they can find by running "odo catalog list services".
`)
)

// DescribeServiceOptions encapsulates the options for the odo catalog describe service command
type DescribeServiceOptions struct {
	// generic context options common to all commands
	*genericclioptions.Context
	backend   CatalogProviderBackend
	isExample bool
}

// NewDescribeServiceOptions creates a new DescribeServiceOptions instance
func NewDescribeServiceOptions() *DescribeServiceOptions {
	return &DescribeServiceOptions{}
}

// Complete completes DescribeServiceOptions after they've been created
func (o *DescribeServiceOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd))
	if err != nil {
		return err
	}

	//we initialize operator backend regardless of if we can split name or not. Decision
	//to describe crs or not will be taken later
	o.backend = NewOperatorBackend()
	return o.backend.CompleteDescribeService(o, args)
}

// Validate validates the DescribeServiceOptions based on completed values
func (o *DescribeServiceOptions) Validate() (err error) {
	return o.backend.ValidateDescribeService(o)

}

// Run contains the logic for the command associated with DescribeServiceOptions
func (o *DescribeServiceOptions) Run(cmd *cobra.Command) (err error) {
	return o.backend.RunDescribeService(o)
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

	command.Flags().BoolVarP(&o.isExample, "example", "e", false, "Show an example of the service")

	return command
}
