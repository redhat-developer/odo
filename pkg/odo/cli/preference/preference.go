package preference

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cli/preference/add"
	"github.com/redhat-developer/odo/pkg/odo/cli/preference/remove"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended preference command name
const RecommendedCommandName = "preference"

var preferenceLongDesc = ktemplates.LongDesc(`Modifies odo specific configuration settings within the global preference file.

%[1]s`)

// NewCmdPreference implements the utils config odo command
func NewCmdPreference(ctx context.Context, name, fullName string) *cobra.Command {

	// Main Commands
	preferenceViewCmd := NewCmdView(viewCommandName, util.GetFullName(fullName, viewCommandName))
	preferenceSetCmd := NewCmdSet(ctx, setCommandName, util.GetFullName(fullName, setCommandName))
	preferenceUnsetCmd := NewCmdUnset(unsetCommandName, util.GetFullName(fullName, unsetCommandName))
	preferenceAddCmd := add.NewCmdAdd(add.RecommendedCommandName, util.GetFullName(fullName, add.RecommendedCommandName))
	preferenceRemoveCmd := remove.NewCmdRemove(remove.RecommendedCommandName, util.GetFullName(fullName, remove.RecommendedCommandName))

	// Subcommands

	// Set the examples
	preferenceCmd := &cobra.Command{
		Use:   name,
		Short: "Modifies preference settings",
		Long:  fmt.Sprintf(preferenceLongDesc, preference.FormatSupportedParameters()),
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
			preferenceViewCmd.Example,
			preferenceSetCmd.Example,
			preferenceUnsetCmd.Example,
			preferenceAddCmd.Example,
			preferenceRemoveCmd.Example,
		),
	}

	// Add the commands, help, usage and annotations
	preferenceCmd.AddCommand(preferenceViewCmd, preferenceSetCmd, preferenceUnsetCmd, preferenceAddCmd, preferenceRemoveCmd)
	preferenceCmd.SetUsageTemplate(util.CmdUsageTemplate)
	preferenceCmd.Annotations = map[string]string{"command": "utility"}

	return preferenceCmd
}
