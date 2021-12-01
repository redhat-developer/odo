package preference

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the odo preference file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a preference value in the global preference`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	paramName       string
	paramValue      string
	configForceFlag bool
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.paramName = args[0]
	o.paramValue = args[1]
	return
}

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *SetOptions) Run(cmd *cobra.Command) (err error) {

	cfg, err := preference.New()

	if err != nil {
		return errors.Errorf("unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}

	if !o.configForceFlag {
		if isSet := cfg.IsSet(o.paramName); isSet {
			// TODO: could add a logic to check if the new value set by the user is not same as the current value
			if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the config", o.paramName)) {
				log.Info("Aborted by the user")
				return nil
			}
		}
	}

	err = cfg.SetConfiguration(strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}

	log.Info("Global preference was successfully updated")
	return nil
}

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	preferenceSetCmd := &cobra.Command{
		Use:   name,
		Short: "Set a value in odo config file",
		Long:  fmt.Sprintf(setLongDesc, preference.FormatSupportedParameters()),
		Example: func(exampleString, fullName string) string {
			cfg, _ := preference.New()
			properties := preference.NewPreferenceList(*cfg)
			for _, property := range properties.Items {
				value := property.Default
				if value == "" {
					value = "foobar"
				}
				exampleString += fmt.Sprintf("\n  %s %s %v", fullName, property.Name, value)
			}
			return "\n" + exampleString
		}(setExample, fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("please provide a parameter name and value")
			} else if len(args) > 2 {
				return fmt.Errorf("only one value per parameter is allowed")
			} else {
				return nil
			}

		}, Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	preferenceSetCmd.Flags().BoolVarP(&o.configForceFlag, "force", "f", false, "Don't ask for confirmation, set the preference directly")
	return preferenceSetCmd
}
