package preference

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/v2/pkg/log"
	"github.com/openshift/odo/v2/pkg/odo/cli/ui"
	"github.com/openshift/odo/v2/pkg/odo/genericclioptions"

	"github.com/openshift/odo/v2/pkg/preference"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const unsetCommandName = "unset"

var (
	unsetLongDesc = ktemplates.LongDesc(`Unset an individual value in the odo preference file.

%[1]s
`)
	unsetExample = ktemplates.Examples(`
   # Unset a preference value in the global preference`)
)

// UnsetOptions encapsulates the options for the command
type UnsetOptions struct {
	paramName           string
	preferenceForceFlag bool
}

// NewUnsetOptions creates a new UnsetOptions instance
func NewUnsetOptions() *UnsetOptions {
	return &UnsetOptions{}
}

// Complete completes UnsetOptions after they've been created
func (o *UnsetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.paramName = args[0]
	return
}

// Validate validates the UnsetOptions based on completed values
func (o *UnsetOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *UnsetOptions) Run(cmd *cobra.Command) (err error) {

	cfg, err := preference.New()

	if err != nil {
		return errors.Errorf("something is wrong with odo, kindly raise an issue at https://github.com/openshift/odo/v2/issues/new?template=Bug.md")
	}

	if !o.preferenceForceFlag {

		if isSet := cfg.IsSet(o.paramName); isSet {
			if !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the preference", o.paramName)) {
				log.Infof("Aborted by the user")
				return nil
			}
		} else {
			return errors.New("preference already unset, cannot unset a preference which is not set")
		}
	}

	err = cfg.DeleteConfiguration(strings.ToLower(o.paramName))
	if err != nil {
		return err
	}

	log.Info("Global preference was successfully updated")
	return nil

}

// NewCmdUnset implements the preference unset odo command
func NewCmdUnset(name, fullName string) *cobra.Command {
	o := NewUnsetOptions()
	preferenceUnsetCmd := &cobra.Command{
		Use:   name,
		Short: "Unset a value in odo preference file",
		Long:  fmt.Sprintf(unsetLongDesc, preference.FormatSupportedParameters()),
		Example: func(exampleString, fullName string) string {
			for _, property := range preference.GetSupportedParameters() {
				exampleString += fmt.Sprintf("\n  %s %s", fullName, property)
			}
			return "\n" + exampleString
		}(unsetExample, fullName),
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
	preferenceUnsetCmd.Flags().BoolVarP(&o.preferenceForceFlag, "force", "f", false, "Don't ask for confirmation, unsetting the preference directly")

	return preferenceUnsetCmd
}
