package preference

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const unsetCommandName = "unset"

var (
	unsetLongDesc = ktemplates.LongDesc(`Unset an individual value in the Odo preference file.

%[1]s
%[2]s
`)
	unsetExample = ktemplates.Examples(`
   # Unset a preference value in the global preference
   %[1]s  %[2]s 
   %[1]s  %[3]s 
   %[1]s  %[4]s 
	`)
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
func (o *UnsetOptions) Run() (err error) {

	cfg, err := preference.New()

	if err != nil {
		return errors.Wrapf(err, "")
	}

	if value, ok := cfg.GetConfiguration(o.paramName); ok && (value != nil) {
		if !o.preferenceForceFlag && !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the preference", o.paramName)) {
			fmt.Println("Aborted by the user.")
			return nil
		}
		err = cfg.DeleteConfiguration(strings.ToLower(o.paramName))
		if err != nil {
			return err
		}
		fmt.Println("Global preference was successfully updated.")
		return nil
		// if its found but nil then show the error
	} else if ok && (value == nil) {
		return errors.New("preference already unset, cannot unset a preference which is not set")
	}
	return errors.New(o.paramName + " is not a valid preference variable")

}

// NewCmdUnset implements the preference unset odo command
func NewCmdUnset(name, fullName string) *cobra.Command {
	o := NewUnsetOptions()
	preferenceUnsetCmd := &cobra.Command{
		Use:     name,
		Short:   "Unset a value in odo preference file",
		Long:    fmt.Sprintf(unsetLongDesc, preference.FormatSupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", unsetExample), fullName, preference.UpdateNotificationSetting, preference.NamePrefixSetting, preference.TimeoutSetting),
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
