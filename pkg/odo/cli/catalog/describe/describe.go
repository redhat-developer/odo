package describe

import (
	"fmt"

	"github.com/openshift/odo/v2/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "describe"

// NewCmdCatalogDescribe implements the odo catalog describe command
func NewCmdCatalogDescribe(name, fullName string) *cobra.Command {
	component := NewCmdCatalogDescribeComponent(componentRecommendedCommandName, util.GetFullName(fullName, componentRecommendedCommandName))
	service := NewCmdCatalogDescribeService(serviceRecommendedCommandName, util.GetFullName(fullName, serviceRecommendedCommandName))
	catalogDescribeCmd := &cobra.Command{
		Use:     name,
		Short:   "Describe catalog item",
		Long:    "Describe the given catalog item from OpenShift",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf("%s\n\n%s\n", component.Example, service.Example),
	}
	catalogDescribeCmd.AddCommand(
		component,
		service,
	)

	return catalogDescribeCmd

}
