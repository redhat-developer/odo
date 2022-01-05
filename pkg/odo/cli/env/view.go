package env

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const viewCommandName = "view"

var (
	viewLongDesc = ktemplates.LongDesc(`
	View current values in odo environment file
	`)

	viewExample = ktemplates.Examples(`
	# For viewing the current environment configuration settings
	%[1]s
	`)
)

// ViewOptions encapsulates the options for the command
type ViewOptions struct {
	// Env context
	cfg *envinfo.EnvSpecificInfo

	// Flags
	contextFlag string
}

// NewViewOptions creates a new ViewOptions instance
func NewViewOptions() *ViewOptions {
	return &ViewOptions{}
}

// Complete completes ViewOptions after they've been created
func (o *ViewOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.cfg, err = envinfo.NewEnvSpecificInfo(o.contextFlag)
	if err != nil {
		return errors.Wrap(err, "failed to load environment file")
	}

	return nil
}

// Validate validates the ViewOptions based on completed values
func (o *ViewOptions) Validate() (err error) {
	if !o.cfg.Exists() {
		return errors.Errorf("the context directory doesn't contain a component, please refer `odo create --help` on how to create a component")
	}

	return nil
}

// Run contains the logic for the command
func (o *ViewOptions) Run() (err error) {
	info := envinfo.NewInfo(o.cfg.GetComponentSettings())
	if log.IsJSON() {
		machineoutput.OutputSuccess(info)
		return
	}
	info.Output(os.Stdout)
	return nil
}

// NewCmdView implements the env view odo command
func NewCmdView(name, fullName string) *cobra.Command {
	o := NewViewOptions()
	envViewCmd := &cobra.Command{
		Use:         name,
		Short:       "View current values in odo environment file",
		Long:        viewLongDesc,
		Example:     fmt.Sprintf(fmt.Sprint(viewExample), fullName),
		Annotations: map[string]string{"machineoutput": "json"},

		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	envViewCmd.Flags().StringVar(&o.contextFlag, "context", "", "Use given context directory as a source for component settings")

	return envViewCmd
}
