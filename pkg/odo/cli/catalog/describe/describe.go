package describe

import (
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "describe"

// NewCmdCatalogDescribe implements the odo catalog describe command
func NewCmdCatalogDescribe(name, fullName string) *cobra.Command {
	catalogDescribeServiceCmd := NewCmdCatalogDescribeService(serviceRecommendedCommandName, util.GetFullName(fullName, serviceRecommendedCommandName))
	command := &cobra.Command{
		Use:     name,
		Short:   "Describe catalog item",
		Long:    "Describe the given catalog item from OpenShift",
		Args:    cobra.ExactArgs(1),
		Example: catalogDescribeServiceCmd.Example,
	}
	command.AddCommand(catalogDescribeServiceCmd)
	return command

}
