package config

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const deleteCommandName = "delete"

var (
	deleteLongDesc = ktemplates.LongDesc(`Delete an individual value in the Odo configuration file.

%[1]s
%[2]s
`)
	deleteExample = ktemplates.Examples(`
   # Delete a configuration value in the global config
   %[1]s --global %[2]s 
   %[1]s --global %[3]s 
   %[1]s --global %[4]s 

   # Delete a configuration value in the local config
   %[1]s %[5]s
	`)
)

// DeleteOptions encapsulates the options for the command
type DeleteOptions struct {
	paramName        string
	configGlobalFlag bool
	configForceFlag  bool
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.paramName = args[0]
	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	return
}

// Run contains the logic for the command
func (o *DeleteOptions) Run() (err error) {
	var cfg config.Info

	if o.configGlobalFlag {
		cfg, err = config.NewGlobalConfig()
	} else {
		cfg, err = config.NewLocalConfig()
	}

	if err != nil {
		return errors.Wrapf(err, "")
	}

	if _, ok := cfg.GetConfiguration(o.paramName); ok {
		err = cfg.DeleteConfiguration(strings.ToLower(o.paramName))
		if err != nil {
			return err
		}
	} else {
		return errors.New("config already unset, cannot delete an unset configuration")
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

// NewCmdDelete implements the config delete odo command
func NewCmdDelete(name, fullName string) *cobra.Command {
	o := NewDeleteOptions()
	configurationDeleteCmd := &cobra.Command{
		Use:     name,
		Short:   "Delete a value in odo config file",
		Long:    fmt.Sprintf(deleteLongDesc, config.FormatSupportedParameters(), config.FormatLocallySupportedParameters()),
		Example: fmt.Sprintf(fmt.Sprint("\n", deleteExample), fullName, config.UpdateNotificationSetting, config.NamePrefixSetting, config.TimeoutSetting, config.ComponentType),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("please provide a parameter name")
			} else if len(args) > 1 {
				return fmt.Errorf("only one parameter is allowed")
			} else {
				return nil
			}

		}, Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}
	configurationDeleteCmd.Flags().BoolVarP(&o.configGlobalFlag, "global", "g", false, "Use the global config file")
	return configurationDeleteCmd
}
