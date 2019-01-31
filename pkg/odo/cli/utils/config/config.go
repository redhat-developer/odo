package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const RecommendedCommandName = "config"

var configLongDesc = ktemplates.LongDesc(fmt.Sprintf("Modifies Odo specific configuration settings within the config file.\n%s", config.FormatSupportedParameters()))

// NewCmdConfiguration implements the utils config odo command
func NewCmdConfiguration(name, fullName string) *cobra.Command {
	configurationViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	configurationSetCmd := NewCmdSet(setCommandName, util.GetFullName(fullName, setCommandName))
	configurationCmd := &cobra.Command{
		Use:   name,
		Short: "Modifies configuration settings",
		Long:  configLongDesc,
		Example: fmt.Sprintf("%s\n%s",
			configurationViewCmd.Example,
			configurationSetCmd.Example),
		Aliases: []string{"configuration"},
	}

	configurationCmd.AddCommand(configurationViewCmd, configurationSetCmd)

	configurationCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return configurationCmd
}
