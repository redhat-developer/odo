package backend

import (
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/log"
	"strconv"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const (
	STATE_ASK_LANG = iota
	STATE_ASK_TYPE
	STATE_END
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	askerClient   asker.Asker
	catalogClient catalog.Client
}

func NewInteractiveBackend(askerClient asker.Asker, catalogClient catalog.Client) *InteractiveBackend {
	return &InteractiveBackend{
		askerClient:   askerClient,
		catalogClient: catalogClient,
	}
}

func (o *InteractiveBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

func (o *InteractiveBackend) SelectDevfile(flags map[string]string, _ filesystem.Filesystem, _ string) (*DevfileLocation, error) {
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
			lang, err = o.askerClient.AskLanguage(langs)
			if err != nil {
				return nil, err
			}
			state = STATE_ASK_TYPE

		case STATE_ASK_TYPE:
			types := devfileEntries.GetProjectTypes(lang)
			var back bool
			back, details, err = o.askerClient.AskType(types)
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

	ok, starter, err := o.askerClient.AskStarterProject(names)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &starterProjects[starter], nil
}

func (o *InteractiveBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	name, err := o.askerClient.AskName(fmt.Sprintf("my-%s-app", devfile.Data.GetMetadata().Name))
	if err != nil {
		return err
	}
	return devfile.SetMetadataName(name)
}

func (o *InteractiveBackend) PersonalizeDevfileconfig(devfileobj parser.DevfileObj) error {
	// TODO: Add tests
	config, err := getPortsAndEnvVar(devfileobj)
	if err != nil {
		return err
	}

	var selectContainerAnswer string
	containerOptions := config.GetContainers()
	containerOptions = append(containerOptions, "NONE - configuration is correct")

	for selectContainerAnswer != "NONE - configuration is correct" {
		PrintConfiguration(config)
		selectContainerAnswer, err = o.asker.AskContainerName(containerOptions)
		if err != nil {
			return err
		}

		selectedContainer := config[selectContainerAnswer]
		if selectContainerAnswer == "NONE - configuration is correct" {
			break
		}

		var configOps asker.OperationOnContainer
		for configOps.Ops != "Nothing" {
			configOps, err = o.asker.AskPersonalizeConfiguration(selectedContainer)
			if err != nil {
				return err
			}
			switch {
			case configOps.Ops == "Delete" && configOps.Kind == "Port":

				portToDelete := configOps.Key
				indexToDelete := -1
				for i, port := range selectedContainer.Ports {
					if port == portToDelete {
						indexToDelete = i
					}
				}
				if indexToDelete == -1 {
					log.Warningf(fmt.Sprintf("unable to delete port %q, not found", portToDelete))
				}
				err = devfileobj.Data.RemovePorts(map[string][]string{selectContainerAnswer: {portToDelete}})
				if err != nil {
					return err
				}
				selectedContainer.Ports = append(selectedContainer.Ports[:indexToDelete], selectedContainer.Ports[indexToDelete+1:]...)

			case configOps.Ops == "Delete" && configOps.Kind == "EnvVar":
				envToDelete := configOps.Key
				if _, ok := selectedContainer.Envs[envToDelete]; !ok {
					log.Warningf(fmt.Sprintf("unable to delete env %q, not found", envToDelete))
				}
				err = devfileobj.Data.RemoveEnvVars(map[string][]string{selectContainerAnswer: {envToDelete}})
				if err != nil {
					return err
				}
				delete(selectedContainer.Envs, envToDelete)

			case configOps.Ops == "Add" && configOps.Kind == "Port":
				var newPort string
				newPort, err = o.asker.AskAddPort()
				if err != nil {
					return err
				}

				err = devfileobj.Data.SetPorts(map[string][]string{selectContainerAnswer: {newPort}})
				if err != nil {
					return err
				}
				selectedContainer.Ports = append(selectedContainer.Ports, newPort)

			case configOps.Ops == "Add" && configOps.Kind == "EnvVar":

				var newEnvNameAnswer, newEnvValueAnswer string
				newEnvNameAnswer, newEnvValueAnswer, err = o.asker.AskAddEnvVar()
				if err != nil {
					return err
				}
				err = devfileobj.Data.AddEnvVars(map[string][]v1alpha2.EnvVar{selectContainerAnswer: {{
					Name:  newEnvNameAnswer,
					Value: newEnvValueAnswer,
				}}})
				if err != nil {
					return err
				}
				selectedContainer.Envs[newEnvNameAnswer] = newEnvValueAnswer

			case configOps.Ops == "Nothing":
				continue
			default:
				return fmt.Errorf("Unknown configuration selected %q", fmt.Sprintf("%v %v %v", configOps.Ops, configOps.Kind, configOps.Key))
			}
			// Update the current configuration
			config[selectContainerAnswer] = selectedContainer
		}
	}
	return devfileobj.WriteYamlDevfile()
}

func PrintConfiguration(config asker.DevfileConfiguration) {
	color.New(color.Bold, color.FgGreen).Println("Current component configuration:")

	for key, container := range config {

		color.Green("Container %q:", key)
		color.Green("  Opened ports:")
		for _, port := range container.Ports {

			color.New(color.Bold, color.FgWhite).Printf("   - %s\n", port)
		}

		color.Green("  Environment variables:")
		for key, value := range container.Envs {
			color.New(color.Bold, color.FgWhite).Printf("   - %s = %s\n", key, value)
		}
	}
}

func getPortsAndEnvVar(obj parser.DevfileObj) (asker.DevfileConfiguration, error) {
	var config = asker.DevfileConfiguration{}
	components, err := obj.Data.GetComponents(parsercommon.DevfileOptions{})
	if err != nil {
		return config, err
	}
	for _, component := range components {
		var ports = []string{}
		var envMap = map[string]string{}
		if component.Container != nil {
			// Fix this for component that are not a container
			for _, ep := range component.Container.Endpoints {
				ports = append(ports, strconv.Itoa(ep.TargetPort))
			}
			for _, env := range component.Container.Env {
				envMap[env.Name] = env.Value
			}
		}
		config[component.Name] = asker.ContainerConfiguration{
			Ports: ports,
			Envs:  envMap,
		}
	}
	return config, nil
}
