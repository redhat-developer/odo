package env

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended env command name
const RecommendedCommandName = "env"

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
