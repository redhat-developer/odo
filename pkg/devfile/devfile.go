package devfile

import (
	"fmt"
	"strings"

	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/log"
)

func parseDevfile(args parser.ParserArgs) (parser.DevfileObj, error) {
	devObj, varWarnings, err := devfile.ParseDevfileAndValidate(args)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// odo specific validations
	err = validate.ValidateDevfileData(devObj.Data)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// display warnings related to variable substitution
	for variable, messages := range varWarnings.Commands {
		log.Warningf(variableWarning("commands", variable, messages))
	}
	for variable, messages := range varWarnings.Components {
		log.Warningf(variableWarning("components", variable, messages))
	}
	for variable, messages := range varWarnings.Projects {
		log.Warningf(variableWarning("projects", variable, messages))
	}
	for variable, messages := range varWarnings.StarterProjects {
		log.Warningf(variableWarning("starterProjects", variable, messages))
	}

	return devObj, nil
}

// ParseAndValidateFromFile reads, parses and validates  devfile from a file
// if there are warning it logs them on stdout
func ParseAndValidateFromFile(devfilePath string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{Path: devfilePath})
}

// ParseAndValidateFromFileWithVariables reads, parses and validates  devfile from a file
// variables are used to override devfile variables
// if there are warning it logs them on stdout
func ParseAndValidateFromFileWithVariables(devfilePath string, variables map[string]string) (parser.DevfileObj, error) {
	return parseDevfile(parser.ParserArgs{
		Path:              devfilePath,
		ExternalVariables: variables,
	})
}

func variableWarning(section string, variable string, messages []string) string {
	quotedVars := []string{}
	for _, v := range messages {
		quotedVars = append(quotedVars, fmt.Sprintf("%q", v))
	}
	return fmt.Sprintf("Invalid variable(s) %s in %q section with name %q. ", strings.Join(quotedVars, ","), section, variable)

}
