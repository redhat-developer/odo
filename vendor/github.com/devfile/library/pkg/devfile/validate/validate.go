package validate

import (
	"fmt"
	v2Validation "github.com/devfile/api/v2/pkg/validation"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/hashicorp/go-multierror"
)

// ValidateDevfileData validates whether sections of devfile are compatible
func ValidateDevfileData(data devfileData.DevfileData) error {

	commands, err := data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return err
	}
	components, err := data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	projects, err := data.GetProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}
	starterProjects, err := data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}

	var returnedErr error
	switch d := data.(type) {
	case *v2.DevfileV2:
		// validate components
		err = v2Validation.ValidateComponents(components)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}

		// validate commands
		err = v2Validation.ValidateCommands(commands, components)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}

		err = v2Validation.ValidateEvents(data.GetEvents(), commands)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}

		err = v2Validation.ValidateProjects(projects)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}

		err = v2Validation.ValidateStarterProjects(starterProjects)
		if err != nil {
			returnedErr = multierror.Append(returnedErr, err)
		}

		return returnedErr

	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}
}
