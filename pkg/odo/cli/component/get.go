package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// GetRecommendedCommandName is the recommended get command name
const GetRecommendedCommandName = "get"

var getExample = ktemplates.Examples(`  # Get the currently active component
%[1]s
  `)

// GetOptions encapsulates component get options
type GetOptions struct {
	componentShortFlag bool
	componentName      string
	componentContext   string
	*genericclioptions.Context
}

// NewGetOptions returns new instance of GetOptions
func NewGetOptions() *GetOptions {
	return &GetOptions{}
}

// Complete completes get args
func (gto *GetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	gto.Context, err = genericclioptions.New(genericclioptions.CreateParameters{Cmd: cmd})
	if err != nil {
		return err
	}
	gto.componentName, err = gto.Context.ComponentAllowingEmpty(true)
	if err != nil {
		return err
	}
	return
}

// Validate validates the get parameters
func (gto *GetOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (gto *GetOptions) Run(cmd *cobra.Command) (err error) {
	klog.V(4).Infof("component get called")

	if gto.componentShortFlag {
		fmt.Print(gto.componentName)
	} else {
		if gto.componentName == "" {
			log.Error("No component is set as current")
			return
		}
		log.Infof("The current component is: %v", gto.componentName)
	}
	return
}

// NewCmdGet implements odo component get command
func NewCmdGet(name, fullName string) *cobra.Command {
	o := NewGetOptions()

	var componentGetCmd = &cobra.Command{
		Use:     name,
		Short:   "Get currently active component",
		Long:    "Get currently active component.",
		Example: fmt.Sprintf(getExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	componentGetCmd.Flags().BoolVarP(&o.componentShortFlag, "short", "q", false, "If true, display only the component name")

	// Hide component get, as we only use this command for autocompletion
	componentGetCmd.Hidden = true

	// add --context flag
	genericclioptions.AddContextFlag(componentGetCmd, &o.componentContext)

	//Adding `--project` flag
	project.AddProjectFlag(componentGetCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentGetCmd)

	return componentGetCmd
}
