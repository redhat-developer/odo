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
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the devfile or odo configuration file.
%[1]s
%[2]s
`)
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

	devfileSetExample = ktemplates.Examples(`
	# Set a configuration value in the devfile
	%[1]s %[2]s testapp
	%[1]s %[3]s 8080/TCP,8443/TCP
	%[1]s %[4]s 500M

	# Set a env variable in the devfiles
	%[1]s --env KAFKA_HOST=kafka --env KAFKA_PORT=6639
	`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	*clicomponent.PushOptions
	paramName       string
	paramValue      string
	configForceFlag bool
	envArray        []string
	now             bool
	IsDevfile       bool
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{PushOptions: clicomponent.NewPushOptions()}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	checkRouteAvailability := false
	if o.now {
		checkRouteAvailability = true
	}
	var err2 error
	o.Context, err2 = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:                    cmd,
		DevfilePath:            "",
		ComponentContext:       o.GetComponentContext(),
		IsNow:                  o.now,
		CheckRouteAvailability: checkRouteAvailability,
	})
	if err2 != nil {
		if util.IsInvalidKubeConfigError(err2) {
			return fmt.Errorf("invalid KUBECONFIG provided. Please point to a valid KUBECONFIG. You do not have to be logged in %w", err2)
		}
		return err2
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
		o.paramValue = args[1]
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

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() (err error) {
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
func (o *SetOptions) DevfileRun() (err error) {
	if o.envArray != nil {
		newEnvVarList, err := config.NewEnvVarListFromSlice(o.envArray)
		if err != nil {
			return err
		}
		err = o.EnvSpecificInfo.GetDevfileObj().AddEnvVars(newEnvVarList.ToDevfileEnv())
		if err != nil {
			return err
		}
		log.Success("Environment variables were successfully updated")
		if o.now {
			return o.DevfilePush()
		}
		log.Italic("\nRun `odo push` command to apply changes to the cluster")
		return err
	}
	if !o.configForceFlag {

		if config.IsSetInDevfile(o.EnvSpecificInfo.GetDevfileObj(), o.paramName) {
			if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the devfile", o.paramName)) {
				fmt.Println("Aborted by the user.")
				return nil
			}
		}
	}

	err = config.SetDevfileConfiguration(o.EnvSpecificInfo.GetDevfileObj(), strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}
	log.Success("Devfile successfully updated")
	if o.now {
		return o.DevfilePush()
	}
	log.Italic("\nRun `odo push` command to apply changes to the cluster")
	return err
}

// Run contains the logic for the command
func (o *SetOptions) Run(cmd *cobra.Command) (err error) {

	if o.IsDevfile {
		return o.DevfileRun()
	}

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

	if !o.configForceFlag {

		if o.LocalConfigInfo.IsSet(o.paramName) {
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
			return fmt.Errorf("failed to push changes %w", err)
		}
	} else {
		log.Italic("\nRun `odo push --config` command to apply changes to the cluster")
	}
	return nil
}

func isValidArgumentList(args []string) error {

	if len(args) < 2 {
		return fmt.Errorf("please provide a parameter name and value")
	} else if len(args) > 2 {
		return fmt.Errorf("only one value per parameter is allowed")
	}

	var err error
	param, ok := config.AsLocallySupportedParameter(args[0])

	if !ok {
		err = errors.Errorf("the provided parameter is not supported, %v", args[0])
	}

	switch param {
	case "memory", "minmemory", "maxmemory", "cpu", "mincpu", "maxcpu":
		err = validation.NonNegativeValidator(args[1])
		if err != nil {
			err = errors.Errorf("%s is invalid %v", param, err)
		}
	case "ports", "debugport":
		err = validation.PortsValidator(args[1])
	}

	if err != nil {
		err = errors.Errorf("validation failed for the provided arguments, %v", err)
	}

	return err
}

func getSetExampleString(fullName string) string {
	s2iExample := fmt.Sprintf(fmt.Sprint("\n", setExample), fullName, config.Type,
		config.Name, config.MinMemory, config.MaxMemory, config.Memory, config.DebugPort, config.Ignore, config.MinCPU, config.MaxCPU, config.CPU, config.Ports)
	devfileExample := fmt.Sprintf("\n"+devfileSetExample, fullName, config.Name, config.Ports, config.Memory)
	return devfileExample + "\n" + s2iExample
}

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	configurationSetCmd := &cobra.Command{
		Use:     name,
		Short:   "Set a value in odo config file",
		Long:    fmt.Sprintf(setLongDesc, config.FormatDevfileSupportedParameters(), config.FormatLocallySupportedParameters()),
		Example: getSetExampleString(fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if o.envArray != nil {
				// no args are needed
				if len(args) > 0 {
					return fmt.Errorf("expected 0 args")
				}
				return nil
			}
			return isValidArgumentList(args)
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
