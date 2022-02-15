package backend

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/redhat-developer/odo/pkg/log"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
)

const (
	STATE_ASK_LANG = iota
	STATE_ASK_TYPE
	STATE_END
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	asker         asker.Asker
	catalogClient catalog.Client
}

func NewInteractiveBackend(asker asker.Asker, catalogClient catalog.Client) *InteractiveBackend {
	return &InteractiveBackend{
		asker:         asker,
		catalogClient: catalogClient,
	}
}

func (o *InteractiveBackend) Validate(flags map[string]string) error {
	return nil
}

func (o *InteractiveBackend) SelectDevfile(flags map[string]string) (*DevfileLocation, error) {
	result := &DevfileLocation{}
	devfileEntries, _ := o.catalogClient.ListDevfileComponents("")

	langs := devfileEntries.GetLanguages()
	state := STATE_ASK_LANG
	var lang string
	var err error
	var details catalog.DevfileComponentType
loop:
	for {
		switch state {

		case STATE_ASK_LANG:
			lang, err = o.asker.AskLanguage(langs)
			if err != nil {
				return nil, err
			}
			state = STATE_ASK_TYPE

		case STATE_ASK_TYPE:
			types := devfileEntries.GetProjectTypes(lang)
			var back bool
			back, details, err = o.asker.AskType(types)
			if err != nil {
				return nil, err
			}
			if back {
				state = STATE_ASK_LANG
				continue loop
			}
			result.DevfileRegistry = details.Registry.Name
			result.Devfile = details.Name
			state = STATE_END
		case STATE_END:
			break loop
		}
	}

	return result, nil
}

func (o *InteractiveBackend) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	starterProjects, err := devfile.Data.GetStarterProjects(parsercommon.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(starterProjects))
	for _, starterProject := range starterProjects {
		names = append(names, starterProject.Name)
	}

	ok, starter, err := o.asker.AskStarterProject(names)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &starterProjects[starter], nil
}

func (o *InteractiveBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	name, err := o.asker.AskName(fmt.Sprintf("my-%s-app", devfile.Data.GetMetadata().Name))
	if err != nil {
		return err
	}
	return devfile.SetMetadataName(name)
}

func (o *InteractiveBackend) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) error {
	var envs = map[string]string{}
	var portsMap = map[string][]string{}
	var deletePortMessage = "Delete port (container: %q): %q"
	var deleteEnvMessage = "Delete environment variable: %q"
	options := []string{
		"NOTHING - configuration is correct",
		"Add new port",
		"Add new environment variable",
	}
	options2 := [][2]string{{""}, {""}, {""}}
	components, err := devfileobj.Data.GetComponents(parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			for _, ep := range component.Container.Endpoints {
				portsMap[component.Name] = append(portsMap[component.Name], strconv.Itoa(ep.TargetPort))
				options = append(options, fmt.Sprintf(deletePortMessage, component.Name, strconv.Itoa(ep.TargetPort)))
				options2 = append(options2, [2]string{component.Name, strconv.Itoa(ep.TargetPort)})
			}
			for _, env := range component.Container.Env {
				envs[env.Name] = env.Value
				options = append(options, fmt.Sprintf(deleteEnvMessage, env.Name))
				options2 = append(options2, [2]string{env.Name})
			}
		}
	}

	var configChangeAnswer string
	var configChangeIndex int
	for configChangeAnswer != "NOTHING - configuration is correct" {
		printConfiguration(portsMap, envs)

		configChangeQuestion := &survey.Select{
			Message: "What configuration do you want change?",
			Default: options[0],
			Options: options,
		}

		err = survey.AskOne(configChangeQuestion, &configChangeIndex)
		if err != nil {
			return err
		}
		configChangeAnswer = options[configChangeIndex]

		if strings.HasPrefix(configChangeAnswer, "Delete port") {
			containerName, portToDelete := options2[configChangeIndex][0], options2[configChangeIndex][1]
			if !parser.InArray(portsMap[containerName], portToDelete) {
				log.Warningf("unable to delete port %q, not found", portToDelete)
				continue
			}
			// Delete port from the devfile
			err = devfileobj.RemovePorts(map[string][]string{containerName: {portToDelete}})
			if err != nil {
				return err
			}
			// Delete port from portsMap
			for i, port := range portsMap[containerName] {
				if port == portToDelete {
					portsMap[containerName] = append(portsMap[containerName][:i], portsMap[containerName][i+1:]...)
					break
				}
			}
			// Delete port from the options
			for i, opt := range options {
				if opt == fmt.Sprintf(deletePortMessage, containerName, portToDelete) {
					options = append(options[:i], options[i+1:]...)
					options2 = append(options2[:i], options2[i+1:]...)
					break
				}
			}
		} else if strings.HasPrefix(configChangeAnswer, "Delete environment variable") {
			envToDelete := options2[configChangeIndex][0]
			if _, ok := envs[envToDelete]; !ok {
				log.Warningf("unable to delete env %q, not found", envToDelete)
			}

			err = devfileobj.RemoveEnvVars([]string{envToDelete})
			if err != nil {
				return err
			}
			delete(envs, envToDelete)
			// Delete env from the options
			for i, opt := range options {
				if opt == fmt.Sprintf(deleteEnvMessage, envToDelete) {
					options = append(options[:i], options[i+1:]...)
					options2 = append(options2[:i], options2[i+1:]...)
					break
				}
			}
		} else if configChangeAnswer == "Add new port" {
			var containers []string
			for containerName, _ := range portsMap {
				containers = append(containers, containerName)
			}
			containerNameQuestion := &survey.Select{
				Message: "Enter container name: ",
				Options: containers,
			}
			var containerNameAnswer string
			survey.AskOne(containerNameQuestion, &containerNameAnswer)

			newPortQuestion := &survey.Input{
				Message: "Enter port number:",
			}
			var newPortAnswer string
			survey.AskOne(newPortQuestion, &newPortAnswer)

			// Ensure the newPortAnswer is not already present; otherwise it will cause a duplicate endpoint error while parsing the devfile
			if parser.InArray(portsMap[containerNameAnswer], newPortAnswer) {
				log.Warningf("Port is %q already present in container %q.", newPortAnswer, containerNameAnswer)
				continue
			}

			// Add port
			err = devfileobj.SetPorts(map[string][]string{containerNameAnswer: []string{newPortAnswer}})
			if err != nil {
				return err
			}
			portsMap[containerNameAnswer] = append(portsMap[containerNameAnswer], newPortAnswer)
			options = append(options, fmt.Sprintf(deletePortMessage, containerNameAnswer, newPortAnswer))
			options2 = append(options2, [2]string{containerNameAnswer, newPortAnswer})
		} else if configChangeAnswer == "Add new environment variable" {
			newEnvNameQuesion := &survey.Input{
				Message: "Enter new environment variable name:",
			}
			// Ask for env name
			var newEnvNameAnswer string
			survey.AskOne(newEnvNameQuesion, &newEnvNameAnswer)
			newEnvValueQuestion := &survey.Input{
				Message: fmt.Sprintf("Enter value for %q environment variable:", newEnvNameAnswer),
			}

			// Ask for env value
			var newEnvValueAnswer string
			survey.AskOne(newEnvValueQuestion, &newEnvValueAnswer)

			// Add env var
			err = devfileobj.AddEnvVars([]v1alpha2.EnvVar{
				{
					Name:  newEnvNameAnswer,
					Value: newEnvValueAnswer,
				},
			})
			if err != nil {
				return err
			}
			envs[newEnvNameAnswer] = newEnvValueAnswer
			options = append(options, fmt.Sprintf(deleteEnvMessage, newEnvNameAnswer))
			options2 = append(options2, [2]string{newEnvNameAnswer})
		} else if configChangeAnswer == "NOTHING - configuration is correct" {
			// nothing to do
		} else {
			return fmt.Errorf("Unknown configuration selected %q", configChangeAnswer)
		}
	}
	return nil
}

func printConfiguration(portsMap map[string][]string, envs map[string]string) {
	color.New(color.Bold, color.FgGreen).Println("Current component configuration:")
	color.Greenln("Opened ports:")
	for containerName, ports := range portsMap {
		color.New(color.Bold, color.FgWhite).Printf(" - Container %q:\n", containerName)
		for _, port := range ports {
			color.New(color.FgWhite).Printf("    · %s\n", port)
		}
	}

	color.Greenln("Environment variables:")
	for key, value := range envs {
		color.New(color.Bold, color.FgWhite).Printf(" - %s = %s\n", key, value)
	}
}
