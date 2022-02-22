package asker

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/redhat-developer/odo/pkg/log"
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

func (o *Survey) AskCorrect() (bool, error) {
	question := &survey.Confirm{
		Message: "Is this correct?",
		Default: true,
	}
	var answer bool
	err := survey.AskOne(question, &answer)
	if err != nil {
		return false, err
	}
	return answer, nil
}

// AskPersonalizeConfiguration asks the configuration user wants to change
func (o *Survey) AskPersonalizeConfiguration(configuration ContainerConfiguration) (ContainerMap, error) {
	options := []string{
		"NOTHING - configuration is correct",
	}
	tracker := []ContainerMap{{Ops: "Nothing"}}
	for _, port := range configuration.Ports {
		options = append(options, fmt.Sprintf("Delete port %q", port))
		tracker = append(tracker, ContainerMap{
			Ops:  "Delete",
			Kind: "Port",
			Key:  port,
		})
	}
	options = append(options, "Add new port")
	tracker = append(tracker, ContainerMap{
		Ops:  "Add",
		Kind: "Port",
	})

	for key := range configuration.Envs {
		options = append(options, fmt.Sprintf("Delete environment variable %q", key))
		tracker = append(tracker, ContainerMap{
			Ops:  "Delete",
			Kind: "EnvVar",
			Key:  key,
		})
	}
	options = append(options, "Add new environment variable")
	tracker = append(tracker, ContainerMap{
		Ops:  "Add",
		Kind: "EnvVar",
	})

	configChangeQuestion := &survey.Select{
		Message: "What configuration do you want change?",
		Default: options[0],
		Options: options,
	}
	var configChangeIndex int
	err := survey.AskOne(configChangeQuestion, &configChangeIndex)
	if err != nil {
		return ContainerMap{}, err
	}
	return tracker[configChangeIndex], nil
}

// AskAddEnvVar asks the key and value for env var
func (o *Survey) AskAddEnvVar() (string, string, error) {
	newEnvNameQuesion := &survey.Input{
		Message: "Enter new environment variable name:",
	}
	var newEnvNameAnswer string
	err := survey.AskOne(newEnvNameQuesion, &newEnvNameAnswer)
	if err != nil {
		return "", "", err
	}
	newEnvValueQuestion := &survey.Input{
		Message: fmt.Sprintf("Enter value for %q environment variable:", newEnvNameAnswer),
	}
	var newEnvValueAnswer string
	err = survey.AskOne(newEnvValueQuestion, &newEnvValueAnswer)
	if err != nil {
		return "", "", err
	}
	return newEnvNameAnswer, newEnvValueAnswer, nil
}

// AskAddPort asks the container name and port that user wants to add
func (o *Survey) AskAddPort() (string, error) {
	newPortQuestion := &survey.Input{
		Message: "Enter port number:",
	}
	var newPortAnswer string
	log.Warning("Please ensure that you do not add a duplicate port number")
	err := survey.AskOne(newPortQuestion, &newPortAnswer)
	if err != nil {
		return "", err
	}
	return newPortAnswer, nil
}

func (o *Survey) AskContainerName(containers []string) (string, error) {
	selectContainerQuestion := &survey.Select{
		Message: "Select container for which you want to change configuration?",
		Default: containers[len(containers)-1],
		Options: containers,
	}
	var selectContainerAnswer string
	err := survey.AskOne(selectContainerQuestion, &selectContainerAnswer)
	if err != nil {
		return selectContainerAnswer, err
	}
	return selectContainerAnswer, nil
}

type ContainerConfiguration struct {
	Ports []string
	Envs  map[string]string
}

type ContainerMap struct {
	Ops  string
	Kind string
	Key  string
}

// key is container name
type DevfileConfiguration map[string]ContainerConfiguration

func (dc *DevfileConfiguration) GetContainers() []string {
	keys := []string{}
	for k := range *dc {
		keys = append(keys, k)
	}
	return keys
}
