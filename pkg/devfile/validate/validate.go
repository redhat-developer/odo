package validate

import (
	"fmt"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/util"
	"k8s.io/klog"

	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
func ValidateDevfileData(data interface{}) error {
	var components []common.DevfileComponent

	switch d := data.(type) {
	case *v200.Devfile200:
		components = d.GetComponents()

		// Validate Events
		if err := validateEvents(d); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	// Validate Components
	if err := validateComponents(components); err != nil {
		return err
	}

	// Successful
	klog.V(2).Info("Successfully validated devfile sections")
	return nil

}

// ValidateContatinerName validates whether the container name is valid for K8
func ValidateContatinerName(devfileData data.DevfileData) error {
	containerComponents := adaptersCommon.GetDevfileContainerComponents(devfileData)
	for _, comp := range containerComponents {
		err := util.ValidateK8sResourceName("container name", comp.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
