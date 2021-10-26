package preference

import (
	"fmt"

	"github.com/openshift/odo/v2/pkg/odo/util"
	"github.com/openshift/odo/v2/pkg/preference"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended preference command name
const RecommendedCommandName = "preference"

var preferenceLongDesc = ktemplates.LongDesc(`Modifies odo specific configuration settings within the global preference file.

%[1]s`)

// NewCmdPreference implements the utils config odo command
func NewCmdPreference(name, fullName string) *cobra.Command {
	preferenceViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	preferenceSetCmd := NewCmdSet(setCommandName, util.GetFullName(fullName, setCommandName))
	preferenceUnsetCmd := NewCmdUnset(unsetCommandName, util.GetFullName(fullName, unsetCommandName))
	preferenceCmd := &cobra.Command{
		Use:   name,
		Short: "Modifies preference settings",
		Long:  fmt.Sprintf(preferenceLongDesc, preference.FormatSupportedParameters()),
		Example: fmt.Sprintf("%s\n%s\n%s",
			preferenceViewCmd.Example,
			preferenceSetCmd.Example,
			preferenceUnsetCmd.Example,
		),
	}

	preferenceCmd.AddCommand(preferenceViewCmd, preferenceSetCmd)
	preferenceCmd.AddCommand(preferenceUnsetCmd)
	preferenceCmd.SetUsageTemplate(util.CmdUsageTemplate)
	preferenceCmd.Annotations = map[string]string{"command": "main"}

	return preferenceCmd
}
