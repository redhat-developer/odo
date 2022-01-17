package init

import (
	"sort"

	"github.com/AlecAivazis/survey/v2"

	"github.com/redhat-developer/odo/pkg/catalog"
)

type Survey struct{}

func NewSurveyAsker() *Survey {
	return &Survey{}
}

func (o *Survey) askLanguage(langs []string) (string, error) {
	sort.Strings(langs)
	question := &survey.Select{
		Message: "Select language:",
		Options: langs,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (o *Survey) askType(types catalog.TypesWithDetails) (catalog.DevfileComponentType, error) {
	stringTypes := types.GetOrderedLabels()
	question := &survey.Select{
		Message: "Select project type:",
		Options: stringTypes,
	}
	var answerPos int
	err := survey.AskOne(question, &answerPos)
	if err != nil {
		return catalog.DevfileComponentType{}, err
	}
	return types.GetAtOrderedPosition(answerPos)
}

func (o *Survey) askStarterProject(projects []string) (string, error) {
	sort.Strings(projects)
	question := &survey.Select{
		Message: "Which starter project do you want to use?",
		Options: projects,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (o *Survey) askName(defaultName string) (string, error) {
	question := &survey.Input{
		Message: "Enter component name:",
		Default: defaultName,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}
