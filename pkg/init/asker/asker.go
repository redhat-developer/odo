package asker

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"sort"

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

func (o *Survey) AskAddPort(containers []string) (containerNameAnswer, newPortAnswer string, err error) {
	containerNameQuestion := &survey.Select{
		Message: "Enter container name: ",
		Options: containers,
	}
	err = survey.AskOne(containerNameQuestion, &containerNameAnswer)
	if err != nil {
		return
	}
	newPortQuestion := &survey.Input{
		Message: "Enter port number:",
	}
	err = survey.AskOne(newPortQuestion, &newPortAnswer)
	if err != nil {
		return
	}
	return
}

func (o *Survey) AskAddEnvVar() (newEnvNameAnswer, newEnvValueAnswer string, err error) {
	newEnvNameQuesion := &survey.Input{
		Message: "Enter new environment variable name:",
	}
	// Ask for env name
	survey.AskOne(newEnvNameQuesion, &newEnvNameAnswer)
	newEnvValueQuestion := &survey.Input{
		Message: fmt.Sprintf("Enter value for %q environment variable:", newEnvNameAnswer),
	}

	// Ask for env value
	err = survey.AskOne(newEnvValueQuestion, &newEnvValueAnswer)
	if err != nil {
		return
	}
	return
}

func (o *Survey) AskPersonalizeConfiguration(options []string) (configChangeAnswer string, configChangeIndex int, err error) {
	configChangeQuestion := &survey.Select{
		Message: "What configuration do you want change?",
		Default: options[0],
		Options: options,
	}

	err = survey.AskOne(configChangeQuestion, &configChangeIndex)
	if err != nil {
		return
	}
	configChangeAnswer = options[configChangeIndex]
	return
}
