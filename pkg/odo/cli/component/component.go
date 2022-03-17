package component

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended component command name
const RecommendedCommandName = "component"

// ComponentOptions encapsulates basic component options
type ComponentOptions struct {
	componentName string
	*genericclioptions.Context
}

// Complete completes component options
func (co *ComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	co.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		co.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).IsOffline())
		if err != nil {
			return err
		}
	}

	co.componentName, err = co.Context.Component(args...)
	if err != nil {
		return nil
	}
	return
}

// NewCmdComponent implements the component odo command
func NewCmdComponent(name, fullName string) *cobra.Command {

	componentGetCmd := NewCmdGet(GetRecommendedCommandName, odoutil.GetFullName(fullName, GetRecommendedCommandName))
	listCmd := NewCmdList(ListRecommendedCommandName, odoutil.GetFullName(fullName, ListRecommendedCommandName))

	// componentCmd represents the component command
	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Manage components",
		Example: fmt.Sprintf("%s\n\n  See sub-commands individually for more examples",
			fullName),
		// `odo component set/get` and `odo get/set` are respectively deprecated as per the new workflow
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd, listCmd)

	// Add a defined annotation in order to appear in the help menu
	componentCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return componentCmd
}
