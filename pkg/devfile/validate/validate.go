package validate

import (
	"reflect"

	"github.com/golang/glog"
	v100 "github.com/openshift/odo/pkg/devfile/parser/data/1.0.0"
)

// Validate validates whether sections of devfile are odo compatible
func ValidateDevfileData(data interface{}) error {

	typeData := reflect.TypeOf(data)
	if typeData == reflect.TypeOf(&v100.Devfile100{}) {
		d := data.(*v100.Devfile100)

		// Validate Projects
		if err := ValidateProjects(d.Projects); err != nil {
			return err
		}

		// Validate Components
		if err := ValidateComponents(d.Components); err != nil {
			return err
		}

		// Validate Commands
		if err := ValidateCommands(d.Commands); err != nil {
			return err
		}

		// Successful
		glog.V(4).Info("Successfully validated devfile sections")
		return nil
	}
	return nil
}
