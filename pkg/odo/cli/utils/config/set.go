package config

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the Odo configuration file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a configuration value in the global config
   %[1]s --global %[2]s false
   %[1]s --global %[3]s "app"
   %[1]s --global %[4]s 20

   # Set a configuration value in the local config
   %[1]s %[5]s java
	`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	paramName        string
	paramValue       string
	configGlobalFlag bool
	configForceFlag  bool
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
	var cfg config.ConfigInfo

	if o.configGlobalFlag {
		cfg, err = config.NewGlobalConfig()
	} else {
		cfg, err = config.NewLocalConfig()
	}

	if err != nil {
		return errors.Wrapf(err, "unable to set configuration")
	}

	if !o.configForceFlag {
		var confirmOveride string
		if value, ok := cfg.GetConfiguration(o.paramName); ok {
			fmt.Printf("%v is already set. Current value is %v.\n", o.paramName, value)
			log.Askf("Do you want to override it in the config? y/N ")
			fmt.Scanln(&confirmOveride)
			if confirmOveride != "y" {
				fmt.Println("Aborted by the user.")
				return nil
			}
		}
	}

	err = cfg.SetConfiguration(strings.ToLower(o.paramName), o.paramValue)
	if err != nil {
		return err
	}

	// cannot use the type switch on non-interface variables so a hack
	var intfcfg interface{} = cfg
	switch intfcfg.(type) {
	case *config.GlobalConfigInfo:
		fmt.Println("Global config was successfully updated.")
	case *config.LocalConfigInfo:
		fmt.Println("Local config was successfully updated.")

	}
	return nil
}

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	configurationSetCmd := &cobra.Command{
		Use:     name,
		Short:   "Set a value in odo config file",
		Long:    fmt.Sprintf(setLongDesc, config.FormatSupportedParameters(), config.FormatLocallySupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", setExample), fullName, config.UpdateNotificationSetting, config.NamePrefixSetting, config.TimeoutSetting, config.ComponentType),
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
	configurationSetCmd.Flags().BoolVarP(&o.configGlobalFlag, "global", "g", false, "Use the global config file")
	configurationSetCmd.Flags().BoolVarP(&o.configForceFlag, "force", "f", false, "Dont ask for confirmation, directly move forward")
	return configurationSetCmd
}
