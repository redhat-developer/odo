package debug

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/debug"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	k8sgenclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"
)

// PortForwardOptions contains all the options for running the port-forward cli command.
type InfoOptions struct {
	Namespace     string
	PortForwarder *debug.DefaultPortForwarder
	*genericclioptions.Context
	contextDir string
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
func (o *InfoOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	cfg, err := config.NewLocalConfigInfo(o.contextDir)
	o.LocalConfigInfo = cfg

	// Using Discard streams because nothing important is logged
	o.PortForwarder = debug.NewDefaultPortForwarder(cfg.GetName(), cfg.GetApplication(), o.Client, k8sgenclioptions.NewTestIOStreamsDiscard())

	return err
}

// Validate validates all the required options for port-forward cmd.
func (o InfoOptions) Validate() error {
	return nil
}

// Run implements all the necessary functionality for port-forward cmd.
func (o InfoOptions) Run() error {
	if debugFileInfo, debugging := debug.GetDebugInfo(o.PortForwarder); debugging {
		if log.IsJSON() {
			machineoutput.OutputSuccess(debugFileInfo)
		} else {
			log.Infof("Debug is running for the component on the local port : %v", debugFileInfo.Spec.LocalPort)
		}
	} else {
		return fmt.Errorf("debug is not running for the component %v", o.LocalConfigInfo.GetName())
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
	genericclioptions.AddContextFlag(cmd, &opts.contextDir)

	return cmd
}
