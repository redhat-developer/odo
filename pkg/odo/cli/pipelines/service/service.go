package service

import (
	"fmt"

	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended environment command name.
const RecommendedCommandName = "service"

// NewCmd creates a new environment command
func NewCmd(name, fullName string) *cobra.Command {

	addCmd := newCmdAdd(addRecommendedCommandName, odoutil.GetFullName(fullName, addRecommendedCommandName))

	var cmd = &cobra.Command{
		Use:   name,
		Short: "Manage services in an environment",
		Long:  "Manage services in a GitOps environment where service source repositories are synchronized",
		Example: fmt.Sprintf("%s\n%s\n\n  See sub-commands individually for more examples",
			fullName, addRecommendedCommandName),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.Flags().AddFlagSet(addCmd.Flags())
	cmd.AddCommand(addCmd)

	cmd.Annotations = map[string]string{"command": "main"}
	cmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return cmd
}
