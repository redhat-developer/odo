package validate

import (
	"fmt"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/openshift/odo/pkg/devfile/validate/generic"

	"k8s.io/klog"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
// after invoking the generic devfile validation
func ValidateDevfileData(data interface{}) error {
	var components []devfilev1.Component
	var commandsMap map[string]devfilev1.Command
	var events devfilev1.Events

	// Validate the generic devfile data before validating odo specific logic
	if err := generic.ValidateDevfileData(data); err != nil {
		return err
	}

	switch d := data.(type) {
	case *v2.DevfileV2:
		components = d.GetComponents()
		commandsMap = d.GetCommands()
		events = d.GetEvents()

		// Validate all the devfile components before validating commands
		if err := validateComponents(components); err != nil {
			return err
		}

		// Validate all the devfile commands before validating events
		if err := validateCommands(commandsMap); err != nil {
			return err
		}

		if err := validateEvents(events); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	// Successful
	klog.V(2).Info("Successfully validated devfile sections")
	return nil

}
