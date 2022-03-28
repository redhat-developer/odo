package preference

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"

	"github.com/redhat-developer/odo/pkg/preference"
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
	// Clients
	clientset *clientset.Clientset

	//Parameters
	paramName string

	// Flags
	forceFlag bool
}

// NewUnsetOptions creates a new UnsetOptions instance
func NewUnsetOptions() *UnsetOptions {
	return &UnsetOptions{}
}

func (o *UnsetOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes UnsetOptions after they've been created
func (o *UnsetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.paramName = strings.ToLower(args[0])
	return
}

// Validate validates the UnsetOptions based on completed values
func (o *UnsetOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *UnsetOptions) Run(ctx context.Context) (err error) {

	if !o.forceFlag {

		if isSet := o.clientset.PreferenceClient.IsSet(o.paramName); isSet {
			if !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the preference", o.paramName)) {
				log.Infof("Aborted by the user")
				return nil
			}
		} else {
			return errors.New("preference already unset, cannot unset a preference which is not set")
		}
	}

	err = o.clientset.PreferenceClient.DeleteConfiguration(o.paramName)
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
			// Just show one example of how to unset a value.
			exampleString += fmt.Sprintf("\n  %s %s", fullName, preference.GetSupportedParameters()[0])
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
	clientset.Add(preferenceUnsetCmd, clientset.PREFERENCE)

	preferenceUnsetCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, unsetting the preference directly")

	return preferenceUnsetCmd
}
