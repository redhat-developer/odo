package ui

import (
	"sort"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/v2/pkg/catalog"
	"github.com/openshift/odo/v2/pkg/odo/cli/ui"
	"github.com/openshift/odo/v2/pkg/odo/util/validation"
	"gopkg.in/AlecAivazis/survey.v1"
)

// SelectStarterProject allows user to select starter project in the prompt
func SelectStarterProject(projects []devfilev1.StarterProject) string {

	if len(projects) == 0 {
		return ""
	}

	projectNames := getProjectNames(projects)

	var download = false
	var selectedProject string
	prompt := &survey.Confirm{Message: "Do you want to download a starter project"}
	err := survey.AskOne(prompt, &download, nil)
	ui.HandleError(err)

	if !download {
		return ""
	}

	// select the only starter project in devfile
	if len(projectNames) == 1 {
		return projectNames[0]
	}

	// If multiple projects present give options to select
	promptSelect := &survey.Select{
		Message: "Which starter project do you want to download",
		Options: projectNames,
	}

	err = survey.AskOne(promptSelect, &selectedProject, survey.Required)
	ui.HandleError(err)
	return selectedProject

}

// SelectDevfileComponentType lets the user to select the devfile component type in the prompt
func SelectDevfileComponentType(options []catalog.DevfileComponentType) string {
	var componentType string
	prompt := &survey.Select{
		Message: "Which devfile component type do you wish to create",
		Options: getDevfileComponentTypeNameCandidates(options),
	}
	err := survey.AskOne(prompt, &componentType, survey.Required)
	ui.HandleError(err)
	return componentType
}

// EnterDevfileComponentName lets the user to specify the component name in the prompt
func EnterDevfileComponentName(defaultComponentName string) string {
	var componentName string
	prompt := &survey.Input{
		Message: "What do you wish to name the new devfile component",
		Default: defaultComponentName,
	}
	err := survey.AskOne(prompt, &componentName, survey.Required)
	ui.HandleError(err)
	return componentName
}

// EnterDevfileComponentProject lets the user to specify the component project in the prompt
func EnterDevfileComponentProject(defaultComponentNamespace string) string {
	var name string
	prompt := &survey.Input{
		Message: "What project do you want the devfile component to be created in",
		Default: defaultComponentNamespace,
	}
	err := survey.AskOne(prompt, &name, validation.NameValidator)
	ui.HandleError(err)
	return name
}

func getDevfileComponentTypeNameCandidates(options []catalog.DevfileComponentType) []string {
	result := make([]string, len(options))
	for i, option := range options {
		result[i] = option.Name
	}
	sort.Strings(result)
	return result
}

func getProjectNames(projects []devfilev1.StarterProject) []string {
	result := make([]string, len(projects))
	for i, project := range projects {
		result[i] = project.Name
	}
	sort.Strings(result)
	return result
}
