package env

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const unsetCommandName = "unset"

var (
	unsetLongDesc = ktemplates.LongDesc(`
	Unset an individual value in the odo environment file
	`)

	unsetExample = ktemplates.Examples(`
   	# Unset an individual value in the environment file
   	%[1]s %[2]s
	`)
)

var (
	supportedUnsetParameters = map[string]string{
		debugportParameter: debugportParameterDescription,
	}
)

// UnsetOptions encapsulates the options for the command
type UnsetOptions struct {
	// Env context
	cfg *envinfo.EnvSpecificInfo

	// Parameters
	paramName string

	// Flags
	contextFlag string
	forceFlag   bool
}

// NewUnsetOptions creates a new UnsetOptions instance
func NewUnsetOptions() *UnsetOptions {
	return &UnsetOptions{}
}

// Complete completes UnsetOptions after they've been created
func (o *UnsetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.cfg, err = envinfo.NewEnvSpecificInfo(o.contextFlag)
	if err != nil {
		return errors.Wrap(err, "failed to load environment file")
	}

	o.paramName = args[0]

	return nil
}

// Validate validates the UnsetOptions based on completed values
func (o *UnsetOptions) Validate() (err error) {
	if !o.cfg.Exists() {
		return errors.Errorf("the context directory doesn't contain a component, please refer `odo create --help` to create a component")
	}

	if !isSupportedParameter(o.paramName, supportedUnsetParameters) {
		return errors.Errorf("%q is not a valid parameter to unset, please refer `odo env unset --help` to unset a valid parameter", o.paramName)
	}

	return nil
}

// Run contains the logic for the command
func (o *UnsetOptions) Run() (err error) {
	if !o.forceFlag {
		if isSet := o.cfg.IsSet(o.paramName); isSet {
			if !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the environment", o.paramName)) {
				log.Infof("Aborted by the user")
				return nil
			}
		} else {
			return errors.New("environment already unset, cannot unset a environment which is not set")
		}
	}

	err = o.cfg.DeleteConfiguration(strings.ToLower(o.paramName))
	if err != nil {
		return err
	}

	log.Info("Environment was successfully updated")
	return nil

}

// NewCmdUnset implements the environment unset odo command
func NewCmdUnset(name, fullName string) *cobra.Command {
	o := NewUnsetOptions()
	envUnsetCmd := &cobra.Command{
		Use:     name,
		Short:   "Unset a value in odo environment file",
		Long:    unsetLongDesc + printSupportedParameters(supportedUnsetParameters),
		Example: fmt.Sprintf(fmt.Sprint(unsetExample), fullName, envinfo.DebugPort),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("please provide a parameter name")
			} else if len(args) > 1 {
				return fmt.Errorf("only one parameter is allowed")
			} else {
				return nil
			}

		}, Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	envUnsetCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, unsetting the environment directly")
	envUnsetCmd.Flags().StringVar(&o.contextFlag, "context", "", "Use given context directory as a source for component settings")

	return envUnsetCmd
}
