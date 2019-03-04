package preference

import (
	"fmt"
	"log"
	"strings"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the Odo preference file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a preference value in the global preference
   %[1]s %[2]s false
   %[1]s %[3]s "app"
   %[1]s %[4]s 20
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
		if value, ok := cfg.GetConfiguration(o.paramName); ok && (value != nil) {
			log.Printf("%v is already set. Current value is %v.\n", o.paramName, value)
			if !ui.Proceed("Do you want to override it in the config") {
				log.Println("Aborted by the user.")
				return nil
			}
		} else if !ok {
			util.LogErrorAndExit(fmt.Errorf("'%s' is not a parameter in the odo config", o.paramName), "")
		}
	}

	err = cfg.SetConfiguration(strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}

	log.Println("Preference was successfully updated.")
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
			preference.UpdateNotificationSetting, preference.NamePrefixSetting, preference.TimeoutSetting),
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
