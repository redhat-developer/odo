package ui

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/redhat-developer/odo/pkg/service"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
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

func GetServicePlanNames(stringMap map[string]scv1beta1.ClusterServicePlan) (keys []string) {
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
		Options: GetServicePlanNames(plans),
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
	err := survey.AskOne(prompt, &category, survey.Required)
	handleError(err)

	classes := getServiceClassMap(classesByCategory[category])
	core.TemplateFuncs["foo"] = func(index int, pageEntries []string) string {
		selected := pageEntries[index]
		class := classes[selected]
		return fmt.Sprintf("Name: %s\nDescription: %s\nLong: %s", class.GetExternalName(), class.GetDescription(), getLongDescription(class))
	}

	survey.SelectQuestionTemplate = survey.SelectQuestionTemplate + `
===
{{- if not .ShowAnswer}}
	{{(foo .SelectedIndex .PageEntries)}}
{{- end}}`

	prompt = &survey.Select{
		Message: "Which " + category + " service class should we use",
		Options: getServiceClassNames(classes),
	}

	err = survey.AskOne(prompt, &serviceType, survey.Required)
	handleError(err)

	return classes[serviceType], serviceType
}

// Convert the provided ClusterServiceClass to its UI representation
func getLongDescription(class scv1beta1.ClusterServiceClass) (longDescription string) {
	extension := class.Spec.ExternalMetadata
	if extension != nil {
		var meta map[string]interface{}
		err := json.Unmarshal(extension.Raw, &meta)
		if err != nil {
			glog.V(4).Infof("Unable unmarshal Extension metadata for ClusterServiceClass '%v'", class.Spec.ExternalName)
		}
		if val, ok := meta["longDescription"]; ok {
			longDescription = val.(string)
		}
	}

	return
}

// EnterServicePropertiesInteractively lets the user enter the properties specified by the provided plan if not already
// specified by the passed values
func EnterServicePropertiesInteractively(svcPlan scv1beta1.ClusterServicePlan) (values map[string]string) {
	return enterServicePropertiesInteractively(svcPlan)
}

func enterServicePropertiesInteractively(svcPlan scv1beta1.ClusterServicePlan, stdio ...terminal.Stdio) (values map[string]string) {
	planDetails, _ := service.NewServicePlans(svcPlan)

	properties := make(map[string]service.ServicePlanParameter, len(planDetails.Parameters))
	for _, v := range planDetails.Parameters {
		properties[v.Name] = v
	}

	values = make(map[string]string, len(properties))

	sort.Sort(planDetails.Parameters)

	// first deal with required properties
	for _, prop := range planDetails.Parameters {
		if prop.Required {
			addValueFor(prop, values, stdio...)
			// remove property from list of properties to consider
			delete(properties, prop.Name)
		}
	}

	// finally check if we still have plan properties that have not been considered
	if len(properties) > 0 {
		fillOptionalProps := false
		confirm := &survey.Confirm{
			Message: "Provide values for non-required properties",
		}

		if len(stdio) == 1 {
			confirm.WithStdio(stdio[0])
		}

		err := survey.AskOne(confirm, &fillOptionalProps, survey.Required)
		handleError(err)
		if fillOptionalProps {

			for _, prop := range properties {
				addValueFor(prop, values, stdio...)
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

func getValidatorFor(prop service.ServicePlanParameter) Validator {
	cv := chainedValidator{}
	if prop.Required {
		cv.validators = append(cv.validators, validators[defaultRequiredValidatorKey])
	}

	switch prop.Type {
	case "integer":
		cv.validators = append(cv.validators, validators[defaultIntegerValidatorKey])
	}

	return cv.validate
}

func addValueFor(prop service.ServicePlanParameter, values map[string]string, stdio ...terminal.Stdio) {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter a value for %s property %s:", prop.Type, propDesc(prop)),
	}

	if len(stdio) == 1 {
		prompt.WithStdio(stdio[0])
	}

	if prop.HasDefaultValue {
		prompt.Default = prop.Default
	}

	err := survey.AskOne(prompt, &result, survey.Validator(getValidatorFor(prop)))
	handleError(err)
	values[prop.Name] = result
}

func propDesc(prop service.ServicePlanParameter) string {
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
