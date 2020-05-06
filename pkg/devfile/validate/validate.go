package validate

import (
	"reflect"

	v100 "github.com/openshift/odo/pkg/devfile/parser/data/1.0.0"
	"k8s.io/klog"
)

// ValidateDevfileData validates whether sections of devfile are odo compatible
func ValidateDevfileData(data interface{}) error {

	typeData := reflect.TypeOf(data)
	if typeData == reflect.TypeOf(&v100.Devfile100{}) {
		d := data.(*v100.Devfile100)

		// Validate Components
		if err := ValidateComponents(d.Components); err != nil {
			return err
		}

		// Successful
		klog.V(4).Info("Successfully validated devfile sections")
		return nil
	}
	return nil
}
