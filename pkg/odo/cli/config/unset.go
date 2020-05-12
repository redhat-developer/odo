package config

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const unsetCommandName = "unset"

var (
	unsetLongDesc = ktemplates.LongDesc(`Unset an individual value in the odo configuration file.

%[1]s
%[2]s
`)
	unsetExample = ktemplates.Examples(`
   # Unset a configuration value in the local config
   %[1]s %[2]s
   %[1]s %[3]s
   %[1]s %[4]s
   %[1]s %[5]s
   %[1]s %[6]s
   %[1]s %[7]s
   %[1]s %[8]s
   %[1]s %[9]s
   %[1]s %[10]s
   %[1]s %[11]s

   # Unset a env variable in the local config
    %[1]s --env KAFKA_HOST --env KAFKA_PORT
	`)
)

// UnsetOptions encapsulates the options for the command
type UnsetOptions struct {
	*clicomponent.CommonPushOptions
	paramName       string
	configForceFlag bool
	envArray        []string
	now             bool
}

// NewUnsetOptions creates a new UnsetOptions instance
func NewUnsetOptions() *UnsetOptions {
	return &UnsetOptions{CommonPushOptions: clicomponent.NewCommonPushOptions()}
}

// Complete completes UnsetOptions after they've been created
func (o *UnsetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if o.envArray == nil {
		o.paramName = args[0]
	}

	if o.now {
		o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
		prjName := o.LocalConfigInfo.GetProject()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	} else {
		o.Context = genericclioptions.NewConfigContext(cmd)
	}
	return
}

// Validate validates the UnsetOptions based on completed values
func (o *UnsetOptions) Validate() (err error) {
	if !o.LocalConfigInfo.ConfigFileExists() {
		return errors.New("the directory doesn't contain a component. Use 'odo create' to create a component")
	}
	if o.now {
		err = o.ValidateComponentCreate()
		if err != nil {
			return err
		}
	}
	return
}

// Run contains the logic for the command
func (o *UnsetOptions) Run() (err error) {

	// env variables have been provided
	if o.envArray != nil {

		envList := o.LocalConfigInfo.GetEnvVars()
		newEnvList, err := config.RemoveEnvVarsFromList(envList, o.envArray)
		if err != nil {
			return err
		}

		if err = o.LocalConfigInfo.SetEnvVars(newEnvList); err != nil {
			return err
		}

		log.Success("Environment variables were successfully updated")
		if o.now {
			err = o.Push()
			if err != nil {
				return errors.Wrap(err, "failed to push changes")
			}
		} else {
			log.Italic("\nRun `odo push --config` command to apply changes to the cluster")
		}
		return nil
	}

	if isSet := o.LocalConfigInfo.IsSet(o.paramName); isSet {
		if !o.configForceFlag && !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the config", o.paramName)) {
			fmt.Println("Aborted by the user.")
			return nil
		}
		err = o.LocalConfigInfo.DeleteConfiguration(strings.ToLower(o.paramName))
		if err != nil {
			return err
		}

		fmt.Println("Local config was successfully updated.")
		if o.now {
			err = o.Push()
			if err != nil {
				return errors.Wrap(err, "failed to push changes")
			}
		} else {
			log.Italic("\nRun `odo push --config` command to apply changes to the cluster")
		}
		return nil
	}
	return errors.New("config already unset, cannot unset a configuration which is not set")

}

// NewCmdUnset implements the config unset odo command
func NewCmdUnset(name, fullName string) *cobra.Command {
	o := NewUnsetOptions()
	configurationUnsetCmd := &cobra.Command{
		Use:   name,
		Short: "Unset a value in odo config file",
		Long:  fmt.Sprintf(unsetLongDesc, config.FormatLocallySupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", unsetExample), fullName,
			config.Type, config.Name, config.MinMemory, config.MaxMemory, config.Memory, config.DebugPort, config.Ignore, config.MinCPU, config.MaxCPU, config.CPU),
		Args: func(cmd *cobra.Command, args []string) error {
			if o.envArray != nil {
				// no args are needed
				if len(args) > 0 {
					return fmt.Errorf("expected 0 args")
				}
				return nil
			}

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
	configurationUnsetCmd.Flags().BoolVarP(&o.configForceFlag, "force", "f", false, "Don't ask for confirmation, unsetting the config directly")
	configurationUnsetCmd.Flags().StringSliceVarP(&o.envArray, "env", "e", nil, "Unset the environment variables in config")
	o.AddContextFlag(configurationUnsetCmd)
	genericclioptions.AddNowFlag(configurationUnsetCmd, &o.now)
	return configurationUnsetCmd
}
