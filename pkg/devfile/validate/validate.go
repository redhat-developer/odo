package validate

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"

	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
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

		commandsMap := common.GetCommandsMap(commands)

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
