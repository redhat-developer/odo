package validate

import (
	"fmt"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v2 "github.com/devfile/library/v2/pkg/devfile/parser/data/v2"
	parsercommon "github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
// after invoking the generic devfile validation
func ValidateDevfileData(data interface{}) error {

	switch d := data.(type) {
	case *v2.DevfileV2:
		components, err := d.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			return err
		}
		commands, err := d.GetCommands(parsercommon.DevfileOptions{})
		if err != nil {
			return err
		}

		commandsMap := getCommandsMap(commands)

		// Validate all the devfile components before validating commands
		if err := validateComponents(components); err != nil {
			return err
		}

		// Validate all the devfile commands before validating events
		if err := validateCommands(commandsMap); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	// Successful
	klog.V(2).Info("Successfully validated devfile sections")
	return nil

}

// getCommandsMap returns a map of the command Id to the command
func getCommandsMap(commands []devfilev1.Command) map[string]devfilev1.Command {
	commandMap := make(map[string]devfilev1.Command, len(commands))
	for _, command := range commands {
		command.Id = strings.ToLower(command.Id)
		commandMap[command.Id] = command
	}
	return commandMap
}
