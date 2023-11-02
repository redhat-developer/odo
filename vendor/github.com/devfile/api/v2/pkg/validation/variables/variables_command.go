package variables

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// ValidateAndReplaceForCommands validates the commands data for global variable references and replaces them with the variable value.
// Returns a map of command ids and invalid variable references if present.
func ValidateAndReplaceForCommands(variables map[string]string, commands []v1alpha2.Command) map[string][]string {

	commandsWarningMap := make(map[string][]string)

	for i := range commands {
		var err error

		// Validate various command types
		switch {
		case commands[i].Exec != nil:
			if err = validateAndReplaceForExecCommand(variables, commands[i].Exec); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					commandsWarningMap[commands[i].Id] = verr.Keys
				}
			}
		case commands[i].Composite != nil:
			if err = validateAndReplaceForCompositeCommand(variables, commands[i].Composite); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					commandsWarningMap[commands[i].Id] = verr.Keys
				}
			}
		case commands[i].Apply != nil:
			if err = validateAndReplaceForApplyCommand(variables, commands[i].Apply); err != nil {
				if verr, ok := err.(*InvalidKeysError); ok {
					commandsWarningMap[commands[i].Id] = verr.Keys
				}
			}
		}
	}

	return commandsWarningMap
}

// validateAndReplaceForExecCommand validates the exec command data for global variable references and replaces them with the variable value
func validateAndReplaceForExecCommand(variables map[string]string, exec *v1alpha2.ExecCommand) error {

	if exec == nil {
		return nil
	}

	var err error
	invalidKeys := make(map[string]bool)

	// Validate exec command line
	if exec.CommandLine, err = validateAndReplaceDataWithVariable(exec.CommandLine, variables); err != nil {
		checkForInvalidError(invalidKeys, err)
	}

	// Validate exec working dir
	if exec.WorkingDir, err = validateAndReplaceDataWithVariable(exec.WorkingDir, variables); err != nil {
		checkForInvalidError(invalidKeys, err)
	}

	// Validate exec label
	if exec.Label, err = validateAndReplaceDataWithVariable(exec.Label, variables); err != nil {
		checkForInvalidError(invalidKeys, err)
	}

	// Validate exec env
	if len(exec.Env) > 0 {
		if err = validateAndReplaceForEnv(variables, exec.Env); err != nil {
			checkForInvalidError(invalidKeys, err)
		}
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForCompositeCommand validates the composite command data for global variable references and replaces them with the variable value
func validateAndReplaceForCompositeCommand(variables map[string]string, composite *v1alpha2.CompositeCommand) error {

	if composite == nil {
		return nil
	}

	var err error
	invalidKeys := make(map[string]bool)

	// Validate composite label
	if composite.Label, err = validateAndReplaceDataWithVariable(composite.Label, variables); err != nil {
		checkForInvalidError(invalidKeys, err)
	}

	return newInvalidKeysError(invalidKeys)
}

// validateAndReplaceForApplyCommand validates the apply command data for global variable references and replaces them with the variable value
func validateAndReplaceForApplyCommand(variables map[string]string, apply *v1alpha2.ApplyCommand) error {

	if apply == nil {
		return nil
	}

	var err error
	invalidKeys := make(map[string]bool)

	// Validate apply label
	if apply.Label, err = validateAndReplaceDataWithVariable(apply.Label, variables); err != nil {
		checkForInvalidError(invalidKeys, err)
	}

	return newInvalidKeysError(invalidKeys)
}
