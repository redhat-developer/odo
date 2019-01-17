package util

import (
	"fmt"
	"gopkg.in/AlecAivazis/survey.v1"
	"strconv"

	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ValidateName will do validation of application & component names
// Criteria for valid name in kubernetes: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
func ValidateName(name string) error {

	errorList := validation.IsDNS1123Label(name)

	if len(errorList) != 0 {
		return errors.New(fmt.Sprintf("%s is not a valid name:  %s", name, strings.Join(errorList, " ")))
	}

	return nil

}

func validateNameAsValidator(name interface{}) error {
	s := name.(string)
	return ValidateName(s)
}

const (
	// NameValidatorKey is the key used to identify the validator associated with ValidateName function
	NameValidatorKey           = "odo_name_validator"
	defaultIntegerValidatorKey = "odo_default_integer"
)

// Validatable represents a common ancestor for validatable parameters
type Validatable struct {
	Required             bool
	Type                 string
	AdditionalValidators []string
}

// AsValidatable allows avoiding type casts in client code
func (v Validatable) AsValidatable() Validatable {
	return v
}

// Validator is a function that validates that the provided interface is conform to expectations or return an error
type Validator func(interface{}) error

// always validates
var nilValidator = func(ans interface{}) error { return nil }

var validators = make(map[string]Validator)

// GetValidatorFor retrieves a validator for the specified validatable, first validating its required state, then its value
// based on type then any additional validators in the order specified by Validatable.AdditionalValidators
func GetValidatorFor(prop Validatable) Validator {
	v, _ := internalGetValidatorFor(prop)
	return v
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
		validatorChain = append(validatorChain, survey.Validator(validators[defaultIntegerValidatorKey]))
	}

	for i := range prop.AdditionalValidators {
		if v, ok := validators[prop.AdditionalValidators[i]]; ok {
			validatorChain = append(validatorChain, survey.Validator(v))
		}
	}

	if len(validatorChain) > 0 {
		return Validator(survey.ComposeValidators(validatorChain...)), validatorChain
	}

	return nilValidator, validatorChain
}

// init initializes the default validators
func init() {
	validators[defaultIntegerValidatorKey] = func(ans interface{}) error {
		s := ans.(string)
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s': %s", s, err)
		}
		return nil
	}

	validators[NameValidatorKey] = validateNameAsValidator
}
