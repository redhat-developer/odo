package asker

import (
	"sort"

	"github.com/AlecAivazis/survey/v2"
)

type Survey struct{}

func NewSurveyAsker() *Survey {
	return &Survey{}
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
	question := &survey.Confirm{
		Message: "Bind as files?",
		Default: true,
	}
	var answer bool
	err := survey.AskOne(question, &answer)
	if err != nil {
		return false, err
	}
	return answer, nil
}
