package config

import (
	"fmt"
	"github.com/openshift/odo/pkg/util"
	"strings"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	clicomponent "github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const unsetCommandName = "unset"

var (
	unsetLongDesc = ktemplates.LongDesc(`Unset an individual value in the devfile or odo configuration file.
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

	devfileUnsetExample = ktemplates.Examples(`
	# Unset a configuration value in the devfile
	%[1]s %[2]s 
	%[1]s %[3]s 
	%[1]s %[4]s

	# Unset a env variable in the devfiles
	%[1]s --env KAFKA_HOST --env KAFKA_PORT
	`)
)

// UnsetOptions encapsulates the options for the command
type UnsetOptions struct {
	*clicomponent.PushOptions
	paramName       string
	configForceFlag bool
	envArray        []string
	now             bool
	IsDevfile       bool
}

// NewUnsetOptions creates a new UnsetOptions instance
func NewUnsetOptions() *UnsetOptions {
	return &UnsetOptions{PushOptions: clicomponent.NewPushOptions()}
}

// Complete completes UnsetOptions after they've been created
func (o *UnsetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	checkRouteAvailability := false
	if o.now {
		checkRouteAvailability = true
	}

	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:                    cmd,
		DevfilePath:            "",
		ComponentContext:       o.GetComponentContext(),
		IsNow:                  o.now,
		CheckRouteAvailability: checkRouteAvailability,
	})
	if err != nil {
		if err1 := util.IsInvalidKubeConfigError(err); err1 != nil {
			return err1
		}
		return err
	}
	if o.Context.EnvSpecificInfo != nil {
		o.IsDevfile = true
		o.DevfilePath = o.Context.EnvSpecificInfo.GetDevfilePath()
		o.EnvSpecificInfo = o.Context.EnvSpecificInfo
	} else {
		o.IsDevfile = false
	}

	if o.envArray == nil {
		o.paramName = args[0]
	}

	if o.now {
		prjName := o.Context.LocalConfigProvider.GetNamespace()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	}

	return
}

// Validate validates the UnsetOptions based on completed values
func (o *UnsetOptions) Validate() (err error) {
	if !o.Context.LocalConfigProvider.Exists() {
		return fmt.Errorf("the directory doesn't contain a component. Use 'odo create' to create a component")
	}
	if !o.IsDevfile && o.now {
		err = o.ValidateComponentCreate()
		if err != nil {
			return err
		}
	}
	return
}

// DevfileRun is ran when the context detects a devfile locally
func (o *UnsetOptions) DevfileRun() (err error) {
	if o.envArray != nil {

		if err := o.EnvSpecificInfo.GetDevfileObj().RemoveEnvVars(o.envArray); err != nil {
			return err
		}
		log.Success("Environment variables were successfully updated")
		if o.now {
			return o.DevfilePush()
		}
		log.Italic("\nRun `odo push` command to apply changes to the cluster")
		return err
	}
	if isSet := config.IsSetInDevfile(o.EnvSpecificInfo.GetDevfileObj(), o.paramName); isSet {
		if !o.configForceFlag && !ui.Proceed(fmt.Sprintf("Do you want to unset %s in the devfile", o.paramName)) {
			fmt.Println("Aborted by the user.")
			return nil
		}
		err = config.DeleteDevfileConfiguration(o.EnvSpecificInfo.GetDevfileObj(), strings.ToLower(o.paramName))
		log.Success("Devfile was successfully updated.")
		if o.now {
			return o.DevfilePush()
		}
		return err
	}
	return fmt.Errorf("config already unset, cannot unset a configuration which is not set")
}

// Run contains the logic for the command
func (o *UnsetOptions) Run(cmd *cobra.Command) (err error) {

	if o.IsDevfile {
		return o.DevfileRun()
	}

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
				return fmt.Errorf("failed to push changes %w", err)
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

		log.Success("Local config was successfully updated.")
		if o.now {
			err = o.Push()
			if err != nil {
				return fmt.Errorf("failed to push changes %w", err)
			}
		} else {
			log.Italic("\nRun `odo push --config` command to apply changes to the cluster")
		}
		return nil
	}
	return fmt.Errorf("config already unset, cannot unset a configuration which is not set")

}

func getUnSetExampleString(fullName string) string {
	s2iExample := fmt.Sprintf(fmt.Sprint("\n", unsetExample), fullName, config.Type,
		config.Name, config.MinMemory, config.MaxMemory, config.Memory, config.DebugPort, config.Ignore, config.MinCPU, config.MaxCPU, config.CPU, config.Ports)
	devfileExample := fmt.Sprintf("\n"+devfileUnsetExample, fullName, config.Name, config.Ports, config.Memory)
	return devfileExample + "\n" + s2iExample
}

// NewCmdUnset implements the config unset odo command
func NewCmdUnset(name, fullName string) *cobra.Command {
	o := NewUnsetOptions()
	configurationUnsetCmd := &cobra.Command{
		Use:     name,
		Short:   "Unset a value in odo config file",
		Long:    fmt.Sprintf(unsetLongDesc, config.FormatDevfileSupportedParameters(), config.FormatLocallySupportedParameters()),
		Example: getUnSetExampleString(fullName),
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
