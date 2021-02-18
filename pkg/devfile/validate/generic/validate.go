package generic

import (
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"k8s.io/klog"

	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
func ValidateDevfileData(data interface{}) error {
	var events devfilev1.Events

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
		events = d.GetEvents()

		// Validate all the devfile components before validating commands
		if err := validateComponents(components); err != nil {
			return err
		}

		// Validate all the devfile commands before validating events
		if err := validateCommands(d.Commands, commandsMap, components); err != nil {
			return err
		}

		// Validate all the events after validating the commands
		if err := validateEvents(events, commandsMap); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	// Successful
	klog.V(2).Info("Successfully validated devfile sections")
	return nil

}
