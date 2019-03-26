package ui

import (
	"fmt"
	"sort"

	"github.com/golang/glog"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/openshift/odo/pkg/util"
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

// SelectImageTag lets the user to select a specific tag for the previously selected builder image in a prompt
func SelectImageTag(options []catalog.CatalogImage, selectedComponentType string) string {
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

// SelectSourceType lets the user select a specific config.SrcType in a prompty
func SelectSourceType(sourceTypes []config.SrcType) config.SrcType {
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
	return config.NONE
}

// EnterInputTypePath allows the user to specify the path on the filesystem in a prompt
func EnterInputTypePath(inputType string, currentDir string, defaultPath ...string) string {
	var path string
	prompt := &survey.Input{
		Message: fmt.Sprintf("Location of %s component, relative to '%s'", inputType, currentDir),
	}

	if len(defaultPath) == 1 {
		prompt.Default = defaultPath[0]
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

// EnterOpenshiftName allows the user to specify the app name in a prompt
func EnterOpenshiftName(defaultName string, message string, context *genericclioptions.Context) string {
	var name string
	prompt := &survey.Input{
		Message: message,
		Default: defaultName,
	}
	err := survey.AskOne(prompt, &name, validation.NameValidator)
	ui.HandleError(err)
	return name
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

// EnterPorts allows the user to specify the ports to be used in a prompt
func EnterPorts() []string {
	var portsStr string
	prompt := &survey.Input{
		Message: "Enter the ports you wish to set (for example: 8080,8100/tcp,9100/udp). Simply press 'Enter' to avoid setting them",
		Default: "",
	}
	err := survey.AskOne(prompt, &portsStr, validation.PortsValidator)
	ui.HandleError(err)

	return util.GetSplitValuesFromStr(portsStr)
}

// EnterEnvVars allows the user to specify the environment variables to be used in a prompt
func EnterEnvVars() []string {
	var envVarsStr string
	prompt := &survey.Input{
		Message: "Enter the environment variables you would like to set (for example: MY_TYPE=backed,PROFILE=dev). Simply press 'Enter' to avoid setting them",
		Default: "",
	}
	err := survey.AskOne(prompt, &envVarsStr, validation.KeyEqValFormatValidator)
	ui.HandleError(err)

	return util.GetSplitValuesFromStr(envVarsStr)
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
