package ui

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
	"os"
	"sort"
	"strconv"
	"strings"
)

const defaultRequiredValidatorKey = "odo_default_required"
const defaultIntegerValidatorKey = "odo_default_integer"

// Validator is a function that validates that the provided interface is conform to expectations or return an error
type Validator func(interface{}) error

var validators = make(map[string]Validator)

type serviceInstanceCreateParameterSchema struct {
	Required   []string
	Properties map[string]property
}

type property struct {
	Name        string
	Title       string
	Type        string
	Description string
	required    bool
}

type properties map[string]property

// Retrieve the list of existing service class categories
func getServiceClassesCategories(categories map[string][]scv1beta1.ClusterServiceClass) (keys []string) {
	keys = make([]string, len(categories))

	i := 0
	for k := range categories {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

func getServicePlanNames(stringMap map[string]scv1beta1.ClusterServicePlan) (keys []string) {
	keys = make([]string, len(stringMap))

	i := 0
	for k := range stringMap {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

func getServiceClassMap(classes []scv1beta1.ClusterServiceClass) (classMap map[string]scv1beta1.ClusterServiceClass) {
	classMap = make(map[string]scv1beta1.ClusterServiceClass, len(classes))
	for _, v := range classes {
		classMap[v.Spec.ExternalName] = v
	}

	return classMap
}

func getServiceClassNames(stringMap map[string]scv1beta1.ClusterServiceClass) (keys []string) {
	keys = make([]string, len(stringMap))

	i := 0
	for k := range stringMap {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	return keys
}

func getProperties(plan scv1beta1.ClusterServicePlan) (props properties, err error) {
	paramBytes := plan.Spec.CommonServicePlanSpec.ServiceInstanceCreateParameterSchema.Raw
	schema := serviceInstanceCreateParameterSchema{}

	err = json.Unmarshal(paramBytes, &schema)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable unmarshal response: %s", string(paramBytes[:]))
	}

	props = make(properties, len(schema.Properties))
	for k, v := range schema.Properties {
		v.Name = k
		v.required = isRequired(schema.Required, k)
		props[k] = v
	}

	return props, err
}

func isRequired(required []string, name string) bool {
	for _, n := range required {
		if n == name {
			return true
		}
	}
	return false
}

func handleError(err error) {
	if err != nil {
		if err == terminal.InterruptErr {
			os.Exit(-1)
		} else {
			glog.V(4).Infof("Encountered an error processing prompt: %v", err)
		}
	}
}

// SelectPlanNameInteractively lets the user to select the plan name from possible options, specifying which text should appear
// in the prompt
func SelectPlanNameInteractively(plans map[string]scv1beta1.ClusterServicePlan, promptText string) (plan string) {
	prompt := &survey.Select{
		Message: promptText,
		Options: getServicePlanNames(plans),
	}
	err := survey.AskOne(prompt, &plan, nil)
	handleError(err)
	return plan
}

// EnterServiceNameInteractively lets the user enter the name of the service instance to create, defaulting to the provided
// default value and specifying both the prompt text and validation function for the name
func EnterServiceNameInteractively(defaultValue, promptText string, validateName Validator) (serviceName string) {
	// if only one arg is given, ask to Name the service providing the class Name as default
	instancePrompt := &survey.Input{
		Message: promptText,
		Default: defaultValue,
	}
	err := survey.AskOne(instancePrompt, &serviceName, survey.Validator(validateName))
	handleError(err)
	return serviceName
}

// SelectClassInteractively lets the user select target service class from possible options, first filtering by categories then
// by class name
func SelectClassInteractively(classesByCategory map[string][]scv1beta1.ClusterServiceClass) (class scv1beta1.ClusterServiceClass, serviceType string) {
	var category string
	prompt := &survey.Select{
		Message: "Which kind of service do you wish to create",
		Options: getServiceClassesCategories(classesByCategory),
	}
	err := survey.AskOne(prompt, &category, nil)
	handleError(err)

	classes := getServiceClassMap(classesByCategory[category])
	prompt = &survey.Select{
		Message: "Which " + category + " service class should we use",
		Options: getServiceClassNames(classes),
	}
	err = survey.AskOne(prompt, &serviceType, nil)
	handleError(err)

	return classes[serviceType], serviceType
}

// EnterServicePropertiesInteractively lets the user enter the properties specified by the provided plan if not already
// specified by the passed values
func EnterServicePropertiesInteractively(svcPlan scv1beta1.ClusterServicePlan, passedValues map[string]string) (values map[string]string) {
	properties, _ := getProperties(svcPlan)
	values = make(map[string]string, len(properties))

	// first deal with required properties
	for name, prop := range properties {
		if prop.required {
			// if the property is required but we don't have a value for it in the passed values, prompt for it
			if _, ok := passedValues[prop.Name]; !ok {
				addValueFor(prop, values)

				// remove property from list of properties to consider
				delete(properties, name)
				delete(passedValues, name)
			}
		}
	}

	// then check if we still have passed values
	for name, value := range passedValues {
		// ignore property if not specified in the plan
		if _, ok := properties[name]; !ok {
			glog.V(4).Infof("Ignoring unknown property '%v'", name)
		} else {
			values[name] = value
		}

		// remove property from list of properties to consider
		delete(properties, name)
		delete(passedValues, name)
	}

	// finally check if we still have plan properties that have not been considered
	if len(properties) > 0 {
		fillOptionalProps := false
		confirm := &survey.Confirm{
			Message: "Provide values for non-required properties",
		}
		err := survey.AskOne(confirm, &fillOptionalProps, nil)
		handleError(err)
		if fillOptionalProps {

			for _, prop := range properties {
				addValueFor(prop, values)
			}
		}
	}

	return values
}

type chainedValidator struct {
	validators []Validator
}

func (cv chainedValidator) validate(input interface{}) error {
	for _, v := range cv.validators {
		err := v(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func getValidatorFor(prop property) Validator {
	cv := chainedValidator{}
	if prop.required {
		cv.validators = append(cv.validators, validators[defaultRequiredValidatorKey])
	}

	switch prop.Type {
	case "integer":
		cv.validators = append(cv.validators, validators[defaultIntegerValidatorKey])
	}

	return cv.validate
}

func addValueFor(prop property, values map[string]string) {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter a value for %s property %s:", prop.Type, propDesc(prop)),
	}
	err := survey.AskOne(prompt, &result, survey.Validator(getValidatorFor(prop)))
	handleError(err)
	values[prop.Name] = result
}

func propDesc(prop property) string {
	msg := ""
	if len(prop.Title) > 0 {
		msg = prop.Title
	} else if len(prop.Description) > 0 {
		msg = prop.Description
	}

	if len(msg) > 0 {
		msg = " (" + strings.TrimSpace(msg) + ")"
	}

	return prop.Name + msg
}

func init() {
	validators[defaultRequiredValidatorKey] = survey.Required

	validators[defaultIntegerValidatorKey] = func(ans interface{}) error {
		s := ans.(string)
		_, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s': %s", s, err)
		}
		return nil
	}
}
