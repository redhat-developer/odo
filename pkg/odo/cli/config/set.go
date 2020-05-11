package config

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the odo configuration file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a configuration value in the local config
   %[1]s %[2]s java
   %[1]s %[3]s test
   %[1]s %[4]s 50M
   %[1]s %[5]s 500M
   %[1]s %[6]s 250M
   %[1]s %[7]s 4040
   %[1]s %[8]s false
   %[1]s %[9]s 0.5
   %[1]s %[10]s 2
   %[1]s %[11]s 1
   %[1]s %[12]s 8080/TCP,8443/TCP

   # Set a env variable in the local config
   %[1]s --env KAFKA_HOST=kafka --env KAFKA_PORT=6639
	`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	*clicomponent.CommonPushOptions
	paramName       string
	paramValue      string
	configForceFlag bool
	envArray        []string
	now             bool
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{CommonPushOptions: clicomponent.NewCommonPushOptions()}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	if o.envArray == nil {
		o.paramName = args[0]
		o.paramValue = args[1]
	}
	// we initialize the context irrespective of --now flag being provided
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

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() (err error) {
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
func (o *SetOptions) Run() (err error) {

	// env variables have been provided
	if o.envArray != nil {
		newEnvVarList, err := config.NewEnvVarListFromSlice(o.envArray)
		if err != nil {
			return err
		}
		// keeping the old env vars as well
		presentEnvVarList := o.LocalConfigInfo.GetEnvVars()
		newEnvVarList = presentEnvVarList.Merge(newEnvVarList)
		if err := o.LocalConfigInfo.SetEnvVars(newEnvVarList); err != nil {
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

	if !o.configForceFlag {
		if isSet := o.LocalConfigInfo.IsSet(o.paramName); isSet {
			if strings.ToLower(o.paramName) == "name" || strings.ToLower(o.paramName) == "project" || strings.ToLower(o.paramName) == "application" {
				if !ui.Proceed(fmt.Sprintf("Are you sure you want to change the component's %s?\nThis action might result in the creation of a duplicate component.\nIf your component is already pushed, please delete the component %q after you apply the changes (odo component delete %s --app %s --project %s)", o.paramName, o.LocalConfigInfo.GetName(), o.LocalConfigInfo.GetName(), o.LocalConfigInfo.GetApplication(), o.LocalConfigInfo.GetProject())) {
					fmt.Println("Aborted by the user.")
					return nil
				}
			} else {
				if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the config", o.paramName)) {
					fmt.Println("Aborted by the user.")
					return nil
				}
			}
		}
	}

	err = o.LocalConfigInfo.SetConfiguration(strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}

	log.Success("Local config successfully updated")
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

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	configurationSetCmd := &cobra.Command{
		Use:   name,
		Short: "Set a value in odo config file",
		Long:  fmt.Sprintf(setLongDesc, config.FormatLocallySupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", setExample), fullName, config.Type,
			config.Name, config.MinMemory, config.MaxMemory, config.Memory, config.DebugPort, config.Ignore, config.MinCPU, config.MaxCPU, config.CPU, config.Ports),
		Args: func(cmd *cobra.Command, args []string) error {
			if o.envArray != nil {
				// no args are needed
				if len(args) > 0 {
					return fmt.Errorf("expected 0 args")
				}
				return nil
			}

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
	configurationSetCmd.Flags().BoolVarP(&o.configForceFlag, "force", "f", false, "Don't ask for confirmation, set the config directly")
	configurationSetCmd.Flags().StringSliceVarP(&o.envArray, "env", "e", nil, "Set the environment variables in config")
	o.AddContextFlag(configurationSetCmd)
	genericclioptions.AddNowFlag(configurationSetCmd, &o.now)
	return configurationSetCmd
}
