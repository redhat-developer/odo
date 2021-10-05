package config

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended config command name
const RecommendedCommandName = "config"

var configLongDesc = ktemplates.LongDesc(`Modifies odo specific configuration settings within the devfile or config file.

%[1]s
`)

// NewCmdConfiguration implements the utils config odo command
func NewCmdConfiguration(name, fullName string) *cobra.Command {
	configurationViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	configurationSetCmd := NewCmdSet(setCommandName, util.GetFullName(fullName, setCommandName))
	configurationUnsetCmd := NewCmdUnset(unsetCommandName, util.GetFullName(fullName, unsetCommandName))
	configurationCmd := &cobra.Command{
		Use:   name,
		Short: "Change or view configuration",
		Long:  fmt.Sprintf(configLongDesc, config.FormatDevfileSupportedParameters()),
		Example: fmt.Sprintf("%s\n%s\n%s",
			configurationViewCmd.Example,
			configurationSetCmd.Example,
			configurationUnsetCmd.Example,
		),
		Aliases: []string{"configuration"},
	}

	configurationCmd.AddCommand(configurationViewCmd, configurationSetCmd)
	configurationCmd.AddCommand(configurationUnsetCmd)
	configurationCmd.SetUsageTemplate(util.CmdUsageTemplate)
	configurationCmd.Annotations = map[string]string{"command": "main"}

	return configurationCmd
}
