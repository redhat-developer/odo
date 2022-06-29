package asker

import (
	"sort"

	"github.com/AlecAivazis/survey/v2"
)

const (
	bindAsFiles  = "Bind As Files"
	bindAsEnvVar = "Bind As Environment Variables"
)

type Survey struct{}

var _ Asker = (*Survey)(nil)

func NewSurveyAsker() *Survey {
	return &Survey{}
}

func (s *Survey) SelectWorkloadResource(options []string) (int, error) {
	question := &survey.Select{
		Message: "Select workload resource you want to bind:",
		Options: options,
	}
	var answer int
	err := survey.AskOne(question, &answer)
	if err != nil {
		return 0, err
	}
	return answer, nil
}

func (s *Survey) SelectWorkloadResourceName(names []string) (bool, string, error) {
	sort.Strings(names)
	notFoundOption := "DOES NOT EXIST YET"
	goBackOption := "** GO BACK **"
	names = append(names, notFoundOption, goBackOption)
	question := &survey.Select{
		Message: "Select workload resource name you want to bind:",
		Options: names,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return false, "", err
	}
	if answer == notFoundOption {
		return false, "", nil
	}
	if answer == goBackOption {
		return true, "", nil
	}
	return false, answer, nil
}

func (s *Survey) AskWorkloadResourceName() (string, error) {
	question := &survey.Input{
		Message: "Enter the Workload's name:",
		Default: "",
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (s *Survey) AskServiceInstance(serviceInstances []string) (string, error) {
	sort.Strings(serviceInstances)
	question := &survey.Select{
		Message: "Select service instance you want to bind to:",
		Options: serviceInstances,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (s *Survey) AskServiceBindingName(defaultName string) (string, error) {
	question := &survey.Input{
		Message: "Enter the Binding's name:",
		Default: defaultName,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (o *Survey) AskBindAsFiles() (bool, error) {
	question := &survey.Select{
		Message: "How do you want to bind the service?",
		Options: []string{bindAsFiles, bindAsEnvVar},
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return true, err
	}

	var bindAsFile bool
	if answer == bindAsFiles {
		bindAsFile = true
	}
	return bindAsFile, nil
}

func (o *Survey) SelectCreationOptions() ([]CreationOption, error) {
	options := []int{}
	prompt := &survey.MultiSelect{
		Message: "Check(with Space Bar) one or more operations to perform with the ServiceBinding:",
		Options: []string{"create it on cluster", "display it", "save it to file"}, // respect order of CreationOption constants
		Help:    "Use the Space Bar to select one or more operations to perform with the ServiceBinding",
	}
	err := survey.AskOne(prompt, &options)
	if err != nil {
		return nil, err
	}
	result := make([]CreationOption, 0, len(options))
	for _, option := range options {
		result = append(result, CreationOption(option))
	}
	return result, nil
}

func (o *Survey) AskOutputFilePath(defaultValue string) (string, error) {
	question := &survey.Input{
		Message: "Save the ServiceBinding to file:",
		Default: defaultValue,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}
