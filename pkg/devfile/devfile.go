package devfile

import (
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/log"
)

func parseRawDevfile(args parser.ParserArgs) (parser.DevfileObj, error) {
	rawArgs := args
	rawArgs.FlattenedDevfile = pointer.Bool(false)
	rawArgs.ConvertKubernetesContentInUri = pointer.Bool(false)
	rawArgs.ImageNamesAsSelector = nil
	rawArgs.SetBooleanDefaults = pointer.Bool(false)

	devfileObj, varWarnings, err := devfile.ParseDevfileAndValidate(rawArgs)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// display warnings related to variable substitution
	displayVariableWarnings(varWarnings)

	return devfileObj, nil
}

func parseEffectiveDevfile(args parser.ParserArgs) (parser.DevfileObj, error) {
	// Effective Devfile with everything resolved (e.g., parent flattened, K8s URIs inlined, ...)
	effectiveArgs := args
	effectiveArgs.FlattenedDevfile = pointer.Bool(true)
	effectiveArgs.ConvertKubernetesContentInUri = pointer.Bool(true)
	effectiveArgs.SetBooleanDefaults = pointer.Bool(false)

	var varWarnings variables.VariableWarning
	devfileObj, varWarnings, err := devfile.ParseDevfileAndValidate(effectiveArgs)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// odo specific validations
	err = validate.ValidateDevfileData(devfileObj.Data)
	if err != nil {
		return parser.DevfileObj{}, err
	}

	// display warnings related to variable substitution
	displayVariableWarnings(varWarnings)

	return devfileObj, nil
}

// ParseAndValidateFromFile reads, parses and validates  devfile from a file
// if there are warning it logs them on stdout
func ParseAndValidateFromFile(devfilePath string, wantEffective bool) (parser.DevfileObj, error) {
	parserArgs := parser.ParserArgs{
		Path: devfilePath,
	}
	if wantEffective {
		return parseEffectiveDevfile(parserArgs)
	}
	return parseRawDevfile(parserArgs)
}

// ParseAndValidateFromFileWithVariables reads, parses and validates  devfile from a file
// variables are used to override devfile variables.
// If wantEffective is true, it returns a complete view of the Devfile, where everything is resolved.
// For example, parent will be flattened in the child, and Kubernetes manifests referenced by URI will be inlined in the related components.
// If there are warnings, it logs them on stdout.
func ParseAndValidateFromFileWithVariables(devfilePath string, variables map[string]string, wantEffective bool) (parser.DevfileObj, error) {
	parserArgs := parser.ParserArgs{
		Path:               devfilePath,
		ExternalVariables:  variables,
		SetBooleanDefaults: pointer.Bool(false),
	}
	if wantEffective {
		return parseEffectiveDevfile(parserArgs)
	}
	return parseRawDevfile(parserArgs)
}

func displayVariableWarnings(varWarnings variables.VariableWarning) {
	variableWarning := func(section string, variable string, messages []string) string {
		var quotedVars []string
		for _, v := range messages {
			quotedVars = append(quotedVars, fmt.Sprintf("%q", v))
		}
		return fmt.Sprintf("Invalid variable(s) %s in %q section with name %q. ", strings.Join(quotedVars, ","), section, variable)
	}

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

}
