package env

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/v2/pkg/odo/util"
	genericUtil "github.com/openshift/odo/v2/pkg/util"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended env command name
const RecommendedCommandName = "env"

const (
	nameParameter                 = "Name"
	nameParameterDescription      = "Use this value to set component name"
	projectParameter              = "Project"
	projectParameterDescription   = "Use this value to set component project"
	debugportParameter            = "DebugPort"
	debugportParameterDescription = "Use this value to set component debug port"
)

var envLongDesc = ktemplates.LongDesc(`Modifies odo specific configuration settings within environment file`)

// NewCmdEnv implements the environment configuration command
func NewCmdEnv(name, fullName string) *cobra.Command {
	envViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	envSetCmd := NewCmdSet(setCommandName, util.GetFullName(fullName, setCommandName))
	envUnsetCmd := NewCmdUnset(unsetCommandName, util.GetFullName(fullName, unsetCommandName))
	envCmd := &cobra.Command{
		Use:   name,
		Short: "Change or view environment configuration",
		Long:  envLongDesc,
		Example: fmt.Sprintf("%s\n\n%s\n\n%s",
			envViewCmd.Example,
			envSetCmd.Example,
			envUnsetCmd.Example,
		),
	}

	envCmd.AddCommand(envViewCmd, envSetCmd, envUnsetCmd)
	envCmd.SetUsageTemplate(util.CmdUsageTemplate)
	envCmd.Annotations = map[string]string{"command": "main"}

	return envCmd
}

func printSupportedParameters(supportedParameters map[string]string) string {
	output := "\n\nAvailable parameters:\n"
	for _, parameter := range genericUtil.GetSortedKeys(supportedParameters) {
		output = fmt.Sprintf("%s  %s: %s\n", output, parameter, supportedParameters[parameter])
	}

	return output
}

func isSupportedParameter(parameter string, supportedParameters map[string]string) bool {
	for supportedParameter := range supportedParameters {
		if strings.EqualFold(supportedParameter, parameter) {
			return true
		}
	}

	return false
}
