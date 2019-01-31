package config

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var supportedParameters = map[string]string{
	"UpdateNotification": "Controls if an update notification is shown or not (true or false)",
	"NamePrefix":         "Default prefix is the current directory name. Use this value to set a default name prefix",
	"Timeout":            "Timeout (in seconds) for OpenShift server connection check",
}

const RecommendedCommandName = "config"

var configLongDesc = ktemplates.LongDesc(fmt.Sprintf("Modifies Odo specific configuration settings within the config file.\n%s", formatSupportedParameters()))

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

func formatSupportedParameters() (result string) {
	for k, v := range supportedParameters {
		result = result + "\n" + k + " - " + v
	}
	return
}

func getSupportedParameters() []string {
	keys := make([]string, len(supportedParameters))

	i := 0
	for k := range supportedParameters {
		keys[i] = k
		i++
	}

	return keys
}
