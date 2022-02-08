package config

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/envvar"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/log"
	clicomponent "github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/validation"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the devfile or odo configuration file.
%[1]s
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
	// Push context
	*clicomponent.PushOptions

	// Parameters
	paramName  string
	paramValue string

	// Flags
	forceFlag    bool
	envArrayFlag []string
	nowFlag      bool
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions(prjClient project.Client, prefClient preference.Client) *SetOptions {
	return &SetOptions{
		PushOptions: clicomponent.NewPushOptions(prjClient, prefClient),
	}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	params := genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.GetComponentContext())
	if o.nowFlag {
		params.CreateAppIfNeeded().RequireRouteAvailability()
	}
	o.Context, err = genericclioptions.New(params)
	if err != nil {
		if err1 := util.IsInvalidKubeConfigError(err); err1 != nil {
			return err1
		}
		return err
	}

	o.DevfilePath = o.Context.EnvSpecificInfo.GetDevfilePath()
	o.EnvSpecificInfo = o.Context.EnvSpecificInfo

	if o.envArrayFlag == nil {
		o.paramName = args[0]
		o.paramValue = args[1]
	}

	if o.nowFlag {
		prjName := o.Context.LocalConfigProvider.GetNamespace()
		o.ResolveSrcAndConfigFlags()
		err = o.ResolveProject(prjName)
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() error {
	if !o.Context.LocalConfigProvider.Exists() {
		return fmt.Errorf("the directory doesn't contain a component. Use 'odo create' to create a component")
	}
	return nil
}

// Run contains the logic for the command
func (o *SetOptions) Run() error {
	if o.envArrayFlag != nil {
		newEnvVarList, err := envvar.NewListFromSlice(o.envArrayFlag)
		if err != nil {
			return err
		}
		err = o.EnvSpecificInfo.GetDevfileObj().AddEnvVars(newEnvVarList.ToDevfileEnvVar())
		if err != nil {
			return err
		}

		log.Success("Environment variables were successfully updated")
		if o.nowFlag {
			return o.DevfilePush()
		}
		log.Italic("\nRun `odo push` command to apply changes to the cluster")
		return err
	}
	if !o.forceFlag {
		if config.IsSetInDevfile(o.EnvSpecificInfo.GetDevfileObj(), o.paramName) {
			if !ui.Proceed(fmt.Sprintf("%v is already set. Do you want to override it in the devfile", o.paramName)) {
				fmt.Println("Aborted by the user.")
				return nil
			}
		}
	}

	err := config.SetDevfileConfiguration(o.EnvSpecificInfo.GetDevfileObj(), strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}
	log.Success("Devfile successfully updated")
	if o.nowFlag {
		return o.DevfilePush()
	}
	log.Italic("\nRun `odo push` command to apply changes to the cluster")
	return err
}

func isValidArgumentList(args []string) error {

	if len(args) < 2 {
		return fmt.Errorf("please provide a parameter name and value")
	} else if len(args) > 2 {
		return fmt.Errorf("only one value per parameter is allowed")
	}

	var err error
	param, ok := config.AsDevfileSupportedParameter(args[0])

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

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	// The error is not handled at this point, it will be handled during Context creation
	kubclient, _ := kclient.New()
	prefClient, err := preference.NewClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}
	o := NewSetOptions(project.NewClient(kubclient), prefClient)
	configurationSetCmd := &cobra.Command{
		Use:     name,
		Short:   "Set a value in odo config file",
		Long:    fmt.Sprintf(setLongDesc, config.FormatDevfileSupportedParameters()),
		Example: fmt.Sprintf("\n"+devfileSetExample, fullName, config.Name, config.Ports, config.Memory),
		Args: func(cmd *cobra.Command, args []string) error {
			if o.envArrayFlag != nil {
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
	configurationSetCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, set the config directly")
	configurationSetCmd.Flags().StringArrayVarP(&o.envArrayFlag, "env", "e", nil, "Set the environment variables in config")
	o.AddContextFlag(configurationSetCmd)
	odoutil.AddNowFlag(configurationSetCmd, &o.nowFlag)

	return configurationSetCmd
}
