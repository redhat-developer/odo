package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const setCommandName = "set"

var (
	setLongDesc = ktemplates.LongDesc(`Set an individual value in the Odo configuration file.

%[1]s`)
	setExample = ktemplates.Examples(`
   # Set a configuration value
   %[1]s %[2]s false
   %[1]s %[3]s "app"
   %[1]s %[4]s 20
	`)
)

// SetOptions encapsulates the options for the command
type SetOptions struct {
	paramName  string
	paramValue string
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
	cfg, err := config.New()
	if err != nil {
		return err
	}
	return cfg.SetConfiguration(o.paramName, o.paramValue)
}

// NewCmdSet implements the config set odo command
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	configurationSetCmd := &cobra.Command{
		Use:     name,
		Short:   "Set a value in odo config file",
		Long:    fmt.Sprintf(setLongDesc, config.FormatSupportedParameters()),
		Example: fmt.Sprintf(setExample, fullName, config.UpdateNotificationSetting, config.NamePrefixSetting, config.TimeoutSetting),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("please provide a parameter name and value")
			} else if len(args) > 2 {
				return fmt.Errorf("only one value per parameter is allowed")
			} else {
				return nil
			}

		}, Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}

	return configurationSetCmd
}
