package ui

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/validation"
	"github.com/redhat-developer/odo/pkg/util"
	"gopkg.in/AlecAivazis/survey.v1"
	"sort"
	"strings"
)

// SelectComponentType lets the user to select the builder image (name only) in the prompt
func SelectComponentType(options []catalog.CatalogImage) string {
	var componentType string
	prompt := &survey.Select{
		Message: "Which component type would wish to create",
		Options: getComponentTypeNameCandidates(options),
	}
	err := survey.AskOne(prompt, &componentType, survey.Required)
	ui.HandleError(err)
	return componentType
}

func getComponentTypeNameCandidates(options []catalog.CatalogImage) []string {
	result := make([]string, len(options))
	for i, option := range options {
		result[i] = option.Name
	}
	sort.Strings(result)
	return result
}

// SelectImageTagInteractively lets the user to select a specific tag for the previously selected builder image in a prompt
func SelectImageTagInteractively(options []catalog.CatalogImage, selectedComponentType string) string {
	var tag string
	prompt := &survey.Select{
		Message: fmt.Sprintf("Which version of '%s' component type would you wish to create", selectedComponentType),
		Options: getTagCandidates(options, selectedComponentType),
	}
	err := survey.AskOne(prompt, &tag, survey.Required)
	ui.HandleError(err)
	return tag
}

func getTagCandidates(options []catalog.CatalogImage, selectedComponentType string) []string {
	for _, option := range options {
		if option.Name == selectedComponentType {
			sort.Strings(option.NonHiddenTags)
			return option.NonHiddenTags
		}
	}
	glog.V(4).Infof("Selected component type %s was not part of the catalog images", selectedComponentType)
	return []string{}
}

// SelectSourceType lets the user select a specific occlient.CreateType in a prompty
func SelectSourceType(sourceTypes []occlient.CreateType) occlient.CreateType {
	options := make([]string, len(sourceTypes))
	for i, sourceType := range sourceTypes {
		options[i] = fmt.Sprint(sourceType)
	}

	var selectedSourceType string
	prompt := &survey.Select{
		Message: "Which input type would wish to use for the component",
		Options: options,
	}
	err := survey.AskOne(prompt, &selectedSourceType, survey.Required)
	ui.HandleError(err)

	for _, sourceType := range sourceTypes {
		if selectedSourceType == fmt.Sprint(sourceType) {
			return sourceType
		}
	}
	glog.V(4).Infof("Selected source type %s was not part of the source type options", selectedSourceType)
	return occlient.NONE
}

// EnterInputTypePath allows the user to specify the path on the filesystem in a prompt
func EnterInputTypePath(inputType string, defaultPath string) string {
	var path string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Where does the %s input for the component reside on the local file system", inputType),
	}
	if len(defaultPath) > 0 {
		prompt.Default = defaultPath
	}
	err := survey.AskOne(prompt, &path, validation.PathValidator)
	ui.HandleError(err)
	return path
}

// we need this because the validator for the component name needs use info from the Context
// so we effectively return a closure that references the context
func createComponentNameValidator(context *genericclioptions.Context) survey.Validator {
	return func(input interface{}) error {
		if s, ok := input.(string); ok {
			err := validation.ValidateName(s)
			if err != nil {
				return err
			}

			exists, err := component.Exists(context.Client, s, context.Application)
			if err != nil {
				glog.V(4).Info(err)
				return fmt.Errorf("Unable to determine if component '%s' exists or not", s)
			}
			if exists {
				return fmt.Errorf("Component with name '%s' already exists in application '%s'", s, context.Application)
			}

			return nil
		}

		return fmt.Errorf("can only validate strings, got %v", input)
	}
}

// EnterComponentName allows the user to specify the component name in a prompt
func EnterComponentName(defaultName string, context *genericclioptions.Context) string {
	var path string
	prompt := &survey.Input{
		Message: "How would you wish to name the new component",
		Default: defaultName,
	}
	err := survey.AskOne(prompt, &path, createComponentNameValidator(context))
	ui.HandleError(err)
	return path
}

// EnterGitInfo will display two prompts, one of the URL of the project and one of the ref
func EnterGitInfo() (string, string) {
	gitURL := enterGitInputTypePath()
	gitRef := enterGitRef("master")

	return gitURL, gitRef
}

func enterGitInputTypePath() string {
	var path string
	prompt := &survey.Input{
		Message: "What is the URL of the git repository you would wish the new component to use",
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

func enterGitRef(defaultRef string) string {
	var path string
	prompt := &survey.Input{
		Message: "What git ref (branch, tag, commit) would you wish to use",
		Default: defaultRef,
	}
	err := survey.AskOne(prompt, &path, survey.Required)
	ui.HandleError(err)
	return path
}

// Proceed displays a given message and asks the user if they want to proceed
func Proceed(message string) bool {
	var response bool
	prompt := &survey.Confirm{
		Message: message,
	}
	err := survey.AskOne(prompt, &response, survey.Required)
	ui.HandleError(err)

	return response
}

func getSplitValuesFromStr(inputStr string) []string {
	if len(inputStr) == 0 {
		return []string{}
	}

	result := strings.Split(inputStr, ",")
	for i, port := range result {
		result[i] = strings.TrimSpace(port)
	}
	return result
}

// EnterPorts allows the user to specify the ports to be used in a prompt
func EnterPorts() []string {
	var portsStr string
	prompt := &survey.Input{
		Message: "Enter the ports you wish to set (for example: 8080,8100/tcp,9100/udp)",
		Default: "",
	}
	err := survey.AskOne(prompt, &portsStr, nil)
	ui.HandleError(err)

	return getSplitValuesFromStr(portsStr)
}

// EnterEnvVars allows the user to specify the environment variables to be used in a prompt
func EnterEnvVars() []string {
	var envVarsStr string
	prompt := &survey.Input{
		Message: "Enter the environment variables you would like to set (for example: MY_TYPE=backed,PROFILE=dev)",
		Default: "",
	}
	err := survey.AskOne(prompt, &envVarsStr, nil)
	ui.HandleError(err)

	return getSplitValuesFromStr(envVarsStr)
}

// EnterMemory allows the user to specify the memory limits to be used in a prompt
func EnterMemory(typeStr string, defaultValue string) string {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter the %s memory (for example 100Mi)", typeStr),
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result, survey.Required)
	ui.HandleError(err)

	return result
}

// EnterCPU allows the user to specify the cpu limits to be used in a prompt
func EnterCPU(typeStr string, defaultValue string) string {
	var result string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Enter the %s CPU (for example 100m or 2)", typeStr),
		Default: defaultValue,
	}
	err := survey.AskOne(prompt, &result, survey.Required)
	ui.HandleError(err)

	return result
}
