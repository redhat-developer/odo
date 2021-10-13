package debug

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/debug"
	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	k8sgenclioptions "k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"
)

// InfoOptions contains all the options for running the info cli command.
type InfoOptions struct {
	componentName   string
	applicationName string
	Namespace       string
	PortForwarder   *debug.DefaultPortForwarder
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
	devfile := location.DevfileFilenamesProvider(o.contextDir)
	if util.CheckPathExists(filepath.Join(o.contextDir, devfile)) {
		o.Context, err = genericclioptions.NewContext(cmd)
		if err != nil {
			return err
		}

		// a small shortcut
		env := o.Context.EnvSpecificInfo

		o.componentName = env.GetName()
		o.Namespace = env.GetNamespace()
	}

	// Using Discard streams because nothing important is logged
	o.PortForwarder = debug.NewDefaultPortForwarder(o.componentName, o.applicationName, o.Namespace, o.Client, o.KClient, k8sgenclioptions.NewTestIOStreamsDiscard())

	return err
}

// Validate validates all the required options for port-forward cmd.
func (o InfoOptions) Validate() error {
	return nil
}

// Run implements all the necessary functionality for port-forward cmd.
func (o InfoOptions) Run(cmd *cobra.Command) error {
	if debugInfo, debugging := debug.GetInfo(o.PortForwarder); debugging {
		if log.IsJSON() {
			machineoutput.OutputSuccess(debugInfo)
		} else {
			log.Infof("Debug is running for the component on the local port : %v", debugInfo.Spec.LocalPort)
		}
	} else {
		return fmt.Errorf("debug is not running for the component %v", o.componentName)
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
