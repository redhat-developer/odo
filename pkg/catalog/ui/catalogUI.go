package ui

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Validator func(string) error

var validators map[string]Validator

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

type serviceClass struct {
	Name            string
	Description     string
	LongDescription string
	Class           scv1beta1.ClusterServiceClass
}

type serviceClasses []serviceClass

func (classes serviceClasses) Len() int {
	return len(classes)
}

func (classes serviceClasses) Less(i, j int) bool {
	return classes[i].Name < classes[j].Name
}

func (classes serviceClasses) Swap(i, j int) {
	classes[i], classes[j] = classes[j], classes[i]
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

// Convert the provided ClusterServiceClasses to serviceClasses
func getUIServiceClasses(classes []scv1beta1.ClusterServiceClass) (uiClasses serviceClasses) {
	uiClasses = make(serviceClasses, 0, len(classes))
	for _, v := range classes {
		uiClasses = append(uiClasses, convertToUI(v))
	}

	sort.Sort(uiClasses)
	return uiClasses
}

// Convert the provided ClusterServiceClass to its UI representation
func convertToUI(class scv1beta1.ClusterServiceClass) serviceClass {
	var meta map[string]interface{}
	err := json.Unmarshal(class.Spec.ExternalMetadata.Raw, &meta)
	if err != nil {
		glog.V(4).Infof("Unable unmarshal Extension metadata for ClusterServiceClass '%v'", class.Spec.ExternalName)
	}
	longDescription := ""
	if val, ok := meta["longDescription"]; ok {
		longDescription = val.(string)
	}
	return serviceClass{
		Name:            class.Spec.ExternalName,
		Description:     class.Spec.Description,
		LongDescription: longDescription,
		Class:           class,
	}
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
		if err == promptui.ErrInterrupt {
			os.Exit(-1)
		} else {
			glog.V(4).Infof("Encountered an error processing prompt: %v", err)
		}
	}
}

// SelectPlanNameInteractively lets the user to select the plan name from possible options, specifying which text should appear
// in the prompt
func SelectPlanNameInteractively(plans map[string]scv1beta1.ClusterServicePlan, promptText string) string {
	prompt := promptui.Select{
		Label: promptText,
		Items: getServicePlanNames(plans),
	}
	_, plan, err := prompt.Run()
	handleError(err)
	return plan
}

// EnterServiceNameInteractively lets the user enter the name of the service instance to create, defaulting to the provided
// default value and specifying both the prompt text and validation function for the name
func EnterServiceNameInteractively(defaultValue, promptText string, validateName Validator) string {
	// if only one arg is given, ask to Name the service providing the class Name as default
	instancePrompt := promptui.Prompt{
		Label:     promptText,
		Default:   defaultValue,
		AllowEdit: true,
		Validate:  promptui.ValidateFunc(validateName),
	}
	serviceName, err := instancePrompt.Run()
	handleError(err)
	return serviceName
}

// SelectClassInteractively lets the user select target service class from possible options, first filtering by categories then
// by class name
func SelectClassInteractively(classesByCategory map[string][]scv1beta1.ClusterServiceClass) (class scv1beta1.ClusterServiceClass, serviceType string) {
	templates := &promptui.SelectTemplates{
		Active:   promptui.IconSelect + " {{ . | cyan }}",
		Inactive: "  {{ . | cyan }}",
		Selected: promptui.IconGood + " Selected category: {{ . | yellow }}",
	}
	prompt := promptui.Select{
		Label:     "Which kind of service do you wish to create",
		Items:     getServiceClassesCategories(classesByCategory),
		Templates: templates,
	}
	_, category, _ := prompt.Run()

	templates = &promptui.SelectTemplates{
		Active:   promptui.IconSelect + " {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: promptui.IconGood + " Selected service class: {{ .Name | yellow | cyan }}",
		Details: `
--------- Service Class ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ "Long:" | faint }}	{{ .LongDescription }}`,
	}
	uiClasses := getUIServiceClasses(classesByCategory[category])
	prompt = promptui.Select{
		Label:     "Which " + category + " service class should we use",
		Items:     uiClasses,
		Templates: templates,
	}
	i, _, err := prompt.Run()
	handleError(err)
	uiClass := uiClasses[i]

	return uiClass.Class, uiClass.Name
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
		trueOrFalse := []bool{true, false}
		templates := &promptui.SelectTemplates{
			Selected: promptui.IconGood + " Provide values for non-required properties: {{ . | yellow }}",
		}
		prompt := promptui.Select{
			Label:     "Provide values for non-required properties",
			Items:     trueOrFalse,
			Templates: templates,
		}

		i, _, err := prompt.Run()
		handleError(err)
		if trueOrFalse[i] {

			for _, prop := range properties {
				addValueFor(prop, values)
			}
		}
	}

	return values
}

var (
	propTemplates = &promptui.PromptTemplates{
		Invalid: promptui.IconBad + "Enter a value for {{ .Type }} property {{ . | propDesc }}: ",
		Valid:   promptui.IconGood + "Enter a value for {{ .Type }} property {{ . | propDesc }}: ",
		Success: promptui.IconGood + " Property {{ .Name | yellow }} set to: ",
	}
	funcMapInit = false
)

type chainedValidator struct {
	validators []Validator
}

func (cv chainedValidator) validate(input string) error {
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
		cv.validators = append(cv.validators, func(s string) error {
			if len(s) == 0 {
				return errors.New("A value is required")
			} else {
				return nil
			}
		})
	}

	switch prop.Type {
	case "integer":
		cv.validators = append(cv.validators, func(s string) error {
			_, err := strconv.Atoi(s)
			if err != nil {
				return errors.New(fmt.Sprintf("Invalid integer value '%s': %s", s, err))
			} else {
				return nil
			}
		})
	}

	return cv.validate
}

func addValueFor(prop property, values map[string]string) {
	if !funcMapInit {
		funcMap := promptui.FuncMap
		funcMap["propDesc"] = func(prop property) string {
			msg := ""
			if len(prop.Title) > 0 {
				msg = prop.Title
			} else if len(prop.Description) > 0 {
				msg = prop.Description
			}

			if len(msg) > 0 {
				msg = " (" + strings.TrimSpace(msg) + ")"
			}

			return funcMap["yellow"].(func(interface{}) string)(prop.Name) + msg
		}
		funcMapInit = true
	}

	prompt := promptui.Prompt{
		Label:     prop,
		AllowEdit: true,
		Templates: propTemplates,
		Validate:  promptui.ValidateFunc(getValidatorFor(prop)),
	}
	result, err := prompt.Run()
	handleError(err)
	values[prop.Name] = result
}
