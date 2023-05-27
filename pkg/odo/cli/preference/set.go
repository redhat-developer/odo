package preference

import (
	"context"
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the odo preference file.  
%[1]s`)
	setExample = ktemplates.Examples(`
   # All available preference values you can set`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Flags
	forceFlag bool

	// Parameters
	paramName  string
	paramValue string
}

var _ genericclioptions.Runnable = (*SetOptions)(nil)

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

func (o *SetOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	o.paramName = strings.ToLower(args[0])
	o.paramValue = args[1]
	return
}

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate(ctx context.Context) (err error) {
	return
}

// Run contains the logic for the command
func (o *SetOptions) Run(ctx context.Context) (err error) {

	if !o.forceFlag {
		if isSet := o.clientset.PreferenceClient.IsSet(o.paramName); isSet {
			// TODO: could add a logic to check if the new value set by the user is not same as the current value
			if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the config", o.paramName)) {
				log.Info("Aborted by the user")
				return nil
			}
		}
	}

	err = o.clientset.PreferenceClient.SetConfiguration(o.paramName, o.paramValue)
	if err != nil {
		return err
	}

	log.Successf("Value of '%s' preference was set to '%s'", o.paramName, o.paramValue)

	scontext.SetPreferenceParameter(ctx, o.paramName, &o.paramValue)
	return nil
}

// NewCmdSet implements the config set odo command
func NewCmdSet(ctx context.Context, name, fullName string) *cobra.Command {
	o := NewSetOptions()
	preferenceSetCmd := &cobra.Command{
		Use:   name,
		Short: "Set a value in the odo preference file",
		Long:  fmt.Sprintf(setLongDesc, preference.FormatSupportedParameters()),
		Example: func(exampleString, fullName string) string {
			prefClient, err := preference.NewClient(ctx)
			if err != nil {
				util.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
			}
			properties := prefClient.NewPreferenceList()
			for _, property := range properties.Items {
				value := property.Default
				exampleString += fmt.Sprintf("\n  %s %s %v", fullName, property.Name, value)
			}
			return "\n" + exampleString
		}(setExample, fullName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(preferenceSetCmd, clientset.PREFERENCE)

	preferenceSetCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, set the preference directly")
	return preferenceSetCmd
}
