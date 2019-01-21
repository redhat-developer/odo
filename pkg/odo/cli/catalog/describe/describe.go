package describe

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "describe"

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
