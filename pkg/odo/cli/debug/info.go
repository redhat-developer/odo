package debug

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/debug"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	k8sgenclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"
)

// InfoOptions contains all the options for running the info cli command.
type InfoOptions struct {
	// Context
	*genericclioptions.Context

	// Flags
	contextFlag string

	// Port forwarder backend
	PortForwarder *debug.DefaultPortForwarder
}

var (
	infoLong = templates.LongDesc(`
			Gets information regarding any debug session of the component.
	`)

	infoExample = templates.Examples(`
		# Get information regarding any debug session of the component
		odo debug info
		
		`)
)

const (
	infoCommandName = "info"
)

func NewInfoOptions() *InfoOptions {
	return &InfoOptions{}
}

// Complete completes all the required options for port-forward cmd.
func (o *InfoOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}

	// Using Discard streams because nothing important is logged
	o.PortForwarder = debug.NewDefaultPortForwarder(o.Context.EnvSpecificInfo.GetName(), o.Context.GetApplication(), o.Context.EnvSpecificInfo.GetNamespace(), o.KClient, k8sgenclioptions.NewTestIOStreamsDiscard())

	return err
}

// Validate validates all the required options for port-forward cmd.
func (o InfoOptions) Validate() error {
	return nil
}

// Run implements all the necessary functionality for port-forward cmd.
func (o InfoOptions) Run() error {
	if debugInfo, debugging := debug.GetInfo(o.PortForwarder); debugging {
		if log.IsJSON() {
			machineoutput.OutputSuccess(debugInfo)
		} else {
			log.Infof("Debug is running for the component on the local port : %v", debugInfo.Spec.LocalPort)
		}
	} else {
		return fmt.Errorf("debug is not running for the component %v", o.Context.EnvSpecificInfo.GetName())
	}
	return nil
}

// NewCmdInfo implements the debug info odo command
func NewCmdInfo(name, fullName string) *cobra.Command {

	opts := NewInfoOptions()
	cmd := &cobra.Command{
		Use:         name,
		Short:       "Displays debug info of a component",
		Long:        infoLong,
		Example:     infoExample,
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(opts, cmd, args)
		},
	}
	odoutil.AddContextFlag(cmd, &opts.contextFlag)

	return cmd
}
