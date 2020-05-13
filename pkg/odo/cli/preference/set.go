package preference

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/log"

	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/preference"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the odo preference file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a preference value in the global preference
   %[1]s %[2]s false
   %[1]s %[3]s "app"
   %[1]s %[4]s 20
   %[1]s %[5]s 30
   %[1]s %[6]s true
   %[1]s %[7]s docker
	`)
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
func (o *SetOptions) Run() (err error) {

	cfg, err := preference.New()

	if err != nil {
		return errors.Wrapf(err, "unable to set preference")
	}

	if !o.configForceFlag {
		if isSet := cfg.IsSet(o.paramName); isSet {
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
		Example: fmt.Sprintf(fmt.Sprint("\n", setExample), fullName,
			preference.UpdateNotificationSetting, preference.NamePrefixSetting,
			preference.TimeoutSetting, preference.PushTimeoutSetting,
			preference.ExperimentalSetting, preference.PushTargetSetting),
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
