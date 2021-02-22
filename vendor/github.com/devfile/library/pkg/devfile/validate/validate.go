package validate

import (
	"fmt"
	v2Validation "github.com/devfile/api/v2/pkg/validation"
	devfileData "github.com/devfile/library/pkg/devfile/parser/data"
	v2 "github.com/devfile/library/pkg/devfile/parser/data/v2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"strings"
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

	var errstrings []string
	switch d := data.(type) {
	case *v2.DevfileV2:
		// validate components
		err = v2Validation.ValidateComponents(components)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}

		// validate commands
		err = v2Validation.ValidateCommands(commands, components)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}

		err = v2Validation.ValidateEvents(data.GetEvents(), commands)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}

		err = v2Validation.ValidateProjects(projects)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}

		err = v2Validation.ValidateStarterProjects(starterProjects)
		if err != nil {
			errstrings = append(errstrings, err.Error())
		}

		if len(errstrings) > 0 {
			return fmt.Errorf(strings.Join(errstrings, "\n"))
		} else {
			return nil
		}
	default:
		return fmt.Errorf("unknown devfile type %T", d)
	}

	if len(errstrings) > 0 {
		return fmt.Errorf(strings.Join(errstrings, "\n"))
	}

	return nil
}
