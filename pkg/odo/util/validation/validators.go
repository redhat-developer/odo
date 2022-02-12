package validation

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"
)

// NameValidator provides a Validator view of the ValidateName function.
func NameValidator(name interface{}) error {
	if s, ok := name.(string); ok {
		return ValidateName(s)
	}

	return fmt.Errorf("can only validate strings, got %v", name)
}

// Validator is a function that validates that the provided interface conforms to expectations or return an error
type Validator func(interface{}) error

// NilValidator always validates
func NilValidator(interface{}) error { return nil }

// IntegerValidator validates that the provided object can be properly converted to an int value
func IntegerValidator(ans interface{}) error {
	if _, ok := ans.(int); ok {
		return nil
	}

	if s, ok := ans.(string); ok {
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s': %s", s, err)
		}
		return nil
	}

	return fmt.Errorf("don't know how to convert %v into an integer", ans)
}

// PathValidator validates whether the given path exists on the file system
func PathValidator(path interface{}) error {
	if s, ok := path.(string); ok {
		exists := util.CheckPathExists(s)
		if exists {
			return nil
		}

		return fmt.Errorf("path '%s' does not exist on the file system", s)
	}

	return fmt.Errorf("can only validate strings, got %v", path)
}

// PortsValidator validates whether all the parts of the input are valid port declarations
// examples of valid input are:
// 8080
// 8080, 9090/udp
// 8080/tcp, 9090/udp
func PortsValidator(portsStr interface{}) error {
	if s, ok := portsStr.(string); ok {
		_, err := dfutil.GetContainerPortsFromStrings(dfutil.GetSplitValuesFromStr(s))
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("can only validate strings, got %v", portsStr)
}

// KeyEqValFormatValidator ensures that all the parts of the input follow the key=value format
// examples of valid input are:
// NAME=JANE
// PORT=8080,PATH=/health
func KeyEqValFormatValidator(portsStr interface{}) error {
	if s, ok := portsStr.(string); ok {
		parts := dfutil.GetSplitValuesFromStr(s)
		for _, part := range parts {
			kvParts := strings.Split(part, "=")
			if len(kvParts) != 2 {
				return fmt.Errorf("Part '%s' does not have the correct format", part)
			}
			if len(strings.TrimSpace(kvParts[0])) != len(kvParts[0]) ||
				len(strings.TrimSpace(kvParts[1])) != len(kvParts[1]) {
				return fmt.Errorf("Spaces are not allowed in '%s'", part)
			}
		}

		return nil
	}

	return fmt.Errorf("can only validate strings, got %v", portsStr)
}

// internalGetValidatorFor exposed for testing purposes
func internalGetValidatorFor(prop Validatable) (validator Validator, chain []survey.Validator) {
	// make sure we don't run into issues when composing validators
	validatorChain := make([]survey.Validator, 0, 5)

	if prop.Required {
		validatorChain = append(validatorChain, survey.Required)
	}

	switch prop.Type {
	case "integer":
		validatorChain = append(validatorChain, survey.Validator(IntegerValidator))
	}

	for i := range prop.AdditionalValidators {
		validatorChain = append(validatorChain, survey.Validator(prop.AdditionalValidators[i]))
	}

	if len(validatorChain) > 0 {
		return Validator(survey.ComposeValidators(validatorChain...)), validatorChain
	}

	return NilValidator, validatorChain
}
