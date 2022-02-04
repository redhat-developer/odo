package asker

import (
	"sort"

	"github.com/AlecAivazis/survey/v2"

	"github.com/redhat-developer/odo/pkg/catalog"
)

type Survey struct{}

func NewSurveyAsker() *Survey {
	return &Survey{}
}

func (o *Survey) AskLanguage(langs []string) (string, error) {
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

func (o *Survey) AskType(types catalog.TypesWithDetails) (back bool, _ catalog.DevfileComponentType, _ error) {
	stringTypes := types.GetOrderedLabels()
	stringTypes = append(stringTypes, "** GO BACK **")
	question := &survey.Select{
		Message: "Select project type:",
		Options: stringTypes,
	}
	var answerPos int
	err := survey.AskOne(question, &answerPos)
	if err != nil {
		return false, catalog.DevfileComponentType{}, err
	}
	if answerPos == len(stringTypes)-1 {
		return true, catalog.DevfileComponentType{}, nil
	}
	compType, err := types.GetAtOrderedPosition(answerPos)
	return false, compType, err
}

func (o *Survey) AskStarterProject(projects []string) (bool, int, error) {
	sort.Strings(projects)
	projects = append(projects, "** NO STARTER PROJECT **")
	question := &survey.Select{
		Message: "Which starter project do you want to use?",
		Options: projects,
	}
	var answer int
	err := survey.AskOne(question, &answer)
	if err != nil {
		return false, 0, err
	}
	if answer == len(projects)-1 {
		return false, 0, nil
	}
	return true, answer, nil
}

func (o *Survey) AskName(defaultName string) (string, error) {
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
