package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/openshift/odo/pkg/service"
	"k8s.io/klog"

	"github.com/mgutz/ansi"
	terminal2 "golang.org/x/crypto/ssh/terminal"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"gopkg.in/AlecAivazis/survey.v1/terminal"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

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

// GetServicePlanNames returns the service plan names included in the specified map
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

// getServiceClassMap converts the specified array of service classes to a name-service class map
func getServiceClassMap(classes []scv1beta1.ClusterServiceClass) (classMap map[string]scv1beta1.ClusterServiceClass) {
	classMap = make(map[string]scv1beta1.ClusterServiceClass, len(classes))
	for _, v := range classes {
		classMap[v.Spec.ExternalName] = v
	}

	return classMap
}

// getServiceClassNames retrieves the keys (service class names) of the specified name-service class mappings
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

// SelectPlanNameInteractively lets the user to select the plan name from possible options, specifying which text should appear
// in the prompt
func SelectPlanNameInteractively(plans map[string]scv1beta1.ClusterServicePlan, promptText string) (plan string) {
	prompt := &survey.Select{
		Message: promptText,
		Options: GetServicePlanNames(plans),
	}
	err := survey.AskOne(prompt, &plan, nil)
	ui.HandleError(err)
	return plan
}

// EnterServiceNameInteractively lets the user enter the name of the service instance to create, defaulting to the provided
// default value and specifying both the prompt text and validation function for the name
func EnterServiceNameInteractively(defaultValue, promptText string, validateName validation.Validator) (serviceName string) {
	// if only one arg is given, ask to Name the service providing the class Name as default
	instancePrompt := &survey.Input{
		Message: promptText,
		Default: defaultValue,
	}
	err := survey.AskOne(instancePrompt, &serviceName, survey.Validator(validateName))
	ui.HandleError(err)
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
	ui.HandleError(err)

	classes := getServiceClassMap(classesByCategory[category])

	// make a new displayClassInfo function available to survey templates to be able to add class information to the display
	displayClassInfo := "displayClassInfo"
	core.TemplateFuncs[displayClassInfo] = func(index int, pageEntries []string) string {
		if index >= 0 && len(pageEntries) > index {
			selected := pageEntries[index]
			class := classes[selected]
			return ansi.ColorCode("default+bu") + "Service class details" + ansi.ColorCode("reset") + ":\n" +
				classInfoItem("Name", class.GetExternalName()) +
				classInfoItem("Description", class.GetDescription()) +
				classInfoItem("Long", getLongDescription(class))
		}
		return "No matching entry"
	}
	defer delete(core.TemplateFuncs, displayClassInfo)

	// record original template and defer restoring it once done
	original := survey.SelectQuestionTemplate
	defer restoreOriginalTemplate(original)

	// add more information about the currently selected class
	survey.SelectQuestionTemplate = original + `
	{{- if not .ShowAnswer}}
	{{$classInfo:=(displayClassInfo .SelectedIndex .PageEntries)}}
	  {{- if $classInfo}}
{{$classInfo}}
	  {{- end}}
	{{- end}}`

	prompt = &survey.Select{
		Message: "Which " + category + " service class should we use",
		Options: getServiceClassNames(classes),
	}

	err = survey.AskOne(prompt, &serviceType, survey.Required)
	ui.HandleError(err)

	return classes[serviceType], serviceType
}

// classInfoItem computes how a given service class information item should be displayed
func classInfoItem(name, value string) string {
	// wrap value if needed accounting for size of value "header" (its name)
	value = wrapIfNeeded(value, len(name)+3)

	if len(value) > 0 {
		// display the name using the default color, in bold and then reset style right after
		return StyledOutput(name, "default+b") + ": " + value + "\n"
	}
	return ""
}

// StyledOutput returns an ANSI color code to style the specified text accordingly, issuing a reset code when done using the
// https://github.com/mgutz/ansi#style-format format
func StyledOutput(text, style string) string {
	return ansi.ColorCode(style) + text + ansi.ColorCode("reset")
}

const defaultColumnNumberBeforeWrap = 80

// wrapIfNeeded wraps the given string taking the given prefix size into account based on the width of the terminal (or
// defaultColumnNumberBeforeWrap if terminal size cannot be determined).
func wrapIfNeeded(value string, prefixSize int) string {
	// get the width of the terminal
	width, _, err := terminal2.GetSize(0)
	if width == 0 || err != nil {
		// if for some reason we couldn't get the size use default value
		width = defaultColumnNumberBeforeWrap
	}

	// if the value length is greater than the width, wrap it
	// note that we need to account for the size of the name of the value being displayed before the value (i.e. its name)
	valueSize := len(value)
	if valueSize+prefixSize >= width {
		// look at each line of the value
		split := strings.Split(value, "\n")
		for index, line := range split {
			// for each line, trim it and split it in space-separated clusters ("words")
			line = strings.TrimSpace(line)
			words := strings.Split(line, " ")
			newLine := ""
			lineSize := 0

			for _, word := range words {
				if lineSize+len(word)+1+prefixSize < width {
					// concatenate word to the new computed line only if adding it to the line won't make it larger than acceptable
					newLine = newLine + " " + word
					lineSize = lineSize + 1 + len(word) // accumulate the line size
				} else {
					// otherwise, break the line and add the word on a new "line"
					newLine = newLine + "\n" + word
					lineSize = len(word) // reset the line size
				}
			}
			// replace the initial line with the new computed version
			split[index] = strings.TrimSpace(newLine)
		}
		// compute the new value by joining all the modified lines
		value = strings.Join(split, "\n")
	}
	return value
}

// restoreOriginalTemplate restores the original survey template once we're done with the display
func restoreOriginalTemplate(original string) {
	survey.SelectQuestionTemplate = original
}

// Convert the provided ClusterServiceClass to its UI representation
func getLongDescription(class scv1beta1.ClusterServiceClass) (longDescription string) {
	extension := class.Spec.ExternalMetadata
	if extension != nil {
		var meta map[string]interface{}
		err := json.Unmarshal(extension.Raw, &meta)
		if err != nil {
			klog.V(4).Infof("Unable unmarshal Extension metadata for ClusterServiceClass '%v'", class.Spec.ExternalName)
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

// enterServicePropertiesInteractively lets user enter the properties interactively using the specified Stdio instance (useful
// for testing purposes)
func enterServicePropertiesInteractively(svcPlan scv1beta1.ClusterServicePlan, stdio ...terminal.Stdio) (values map[string]string) {
	planDetails, _ := service.NewServicePlan(svcPlan)

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
	if len(properties) > 0 && ui.Proceed("Provide values for non-required properties", stdio...) {
		for _, prop := range properties {
			addValueFor(prop, values, stdio...)
		}
	}

	return values
}

func addValueFor(prop service.ServicePlanParameter, values map[string]string, stdio ...terminal.Stdio) {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter a value for %s property %s:", prop.Type, propDesc(prop)),
	}

	if len(stdio) == 1 {
		prompt.WithStdio(stdio[0])
	}

	if len(prop.Default) > 0 {
		prompt.Default = prop.Default
	}

	err := survey.AskOne(prompt, &result, ui.GetValidatorFor(prop.AsValidatable()))
	ui.HandleError(err)
	values[prop.Name] = result
}

// propDesc computes a human-readable description of the specified property
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
