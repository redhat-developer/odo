package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
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
	// Context
	*genericclioptions.Context

	// Flags
	shortFlag   bool
	contextFlag string

	componentName string
}

// NewGetOptions returns new instance of GetOptions
func NewGetOptions() *GetOptions {
	return &GetOptions{}
}

// Complete completes get args
func (gto *GetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	gto.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
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
func (gto *GetOptions) Run() (err error) {
	klog.V(4).Infof("component get called")

	if gto.shortFlag {
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

	componentGetCmd.Flags().BoolVarP(&o.shortFlag, "short", "q", false, "If true, display only the component name")

	// Hide component get, as we only use this command for autocompletion
	componentGetCmd.Hidden = true

	// add --context flag
	odoutil.AddContextFlag(componentGetCmd, &o.contextFlag)

	//Adding `--project` flag
	project.AddProjectFlag(componentGetCmd)

	return componentGetCmd
}
