package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	dfutil "github.com/devfile/library/pkg/util"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

const (
	STATE_ASK_LANG = iota
	STATE_ASK_TYPE
	STATE_END
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	askerClient    asker.Asker
	registryClient registry.Client
	alizerClient   alizer.Client
}

var _ InitBackend = (*InteractiveBackend)(nil)

func NewInteractiveBackend(askerClient asker.Asker, registryClient registry.Client, alizerClient alizer.Client) *InteractiveBackend {
	return &InteractiveBackend{
		askerClient:    askerClient,
		registryClient: registryClient,
		alizerClient:   alizerClient,
	}
}

func (o *InteractiveBackend) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	return nil
}

func (o *InteractiveBackend) SelectDevfile(ctx context.Context, flags map[string]string, _ filesystem.Filesystem, _ string) (*api.DetectionResult, error) {
	result := &api.DetectionResult{}
	devfileEntries, _ := o.registryClient.ListDevfileStacks(ctx, "", "", "", false)

	langs := devfileEntries.GetLanguages()
	state := STATE_ASK_LANG
	var lang string
	var err error
	var details api.DevfileStack
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

	sort.Slice(starterProjects, func(i, j int) bool {
		return starterProjects[i].Name < starterProjects[j].Name
	})

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

func (o *InteractiveBackend) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {

	// We will retrieve the name using alizer and then suggest it as the default name.
	// 1. Check the pom.xml / package.json / etc. for the project name
	// 2. If not, use the directory name instead

	// Get the absolute path to the directory from the Devfile context
	path := devfile.Ctx.GetAbsPath()
	if path == "" {
		return "", fmt.Errorf("unable to get the absolute path to the directory: %q", path)
	}

	// Pass in the BASE directory (not the file name of devfile.yaml)
	// Convert path to base dir not file name
	baseDir := filepath.Dir(path)

	// Detect the name
	name, err := o.alizerClient.DetectName(baseDir)
	if err != nil {
		return "", fmt.Errorf("detecting name using alizer: %w", err)
	}

	klog.V(4).Infof("Detected name via alizer: %q from path: %q", name, baseDir)

	if name == "" {
		return "", fmt.Errorf("unable to detect the name")
	}

	var userReturnedName string
	// keep asking the name until the user enters a valid name
	for {
		userReturnedName, err = o.askerClient.AskName(name)
		if err != nil {
			return "", err
		}
		validK8sNameErr := dfutil.ValidateK8sResourceName("name", userReturnedName)
		if validK8sNameErr == nil {
			break
		}
		log.Error(validK8sNameErr)
	}
	return userReturnedName, nil
}

func (o *InteractiveBackend) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) (parser.DevfileObj, error) {
	config, err := getPortsAndEnvVar(devfileobj)
	var zeroDevfile parser.DevfileObj
	if err != nil {
		return zeroDevfile, err
	}

	var selectContainerAnswer string
	containerOptions := config.GetContainers()
	containerOptions = append(containerOptions, "NONE - configuration is correct")

	for selectContainerAnswer != "NONE - configuration is correct" {
		PrintConfiguration(config)
		selectContainerAnswer, err = o.askerClient.AskContainerName(containerOptions)
		if err != nil {
			return zeroDevfile, err
		}

		selectedContainer := config[selectContainerAnswer]
		if selectContainerAnswer == "NONE - configuration is correct" {
			break
		}

		var configOps asker.OperationOnContainer
		for configOps.Ops != "Nothing" {
			configOps, err = o.askerClient.AskPersonalizeConfiguration(selectedContainer)
			if err != nil {
				return zeroDevfile, err
			}
			switch configOps.Ops {
			case "Add":
				switch configOps.Kind {
				case "Port":
					var newPort string
					newPort, err = o.askerClient.AskAddPort()
					if err != nil {
						return zeroDevfile, err
					}

					err = devfileobj.Data.SetPorts(map[string][]string{selectContainerAnswer: {newPort}})
					if err != nil {
						return zeroDevfile, err
					}
					selectedContainer.Ports = append(selectedContainer.Ports, newPort)

				case "EnvVar":
					var newEnvNameAnswer, newEnvValueAnswer string
					newEnvNameAnswer, newEnvValueAnswer, err = o.askerClient.AskAddEnvVar()
					if err != nil {
						return zeroDevfile, err
					}
					err = devfileobj.Data.AddEnvVars(map[string][]v1alpha2.EnvVar{selectContainerAnswer: {{
						Name:  newEnvNameAnswer,
						Value: newEnvValueAnswer,
					}}})
					if err != nil {
						return zeroDevfile, err
					}
					selectedContainer.Envs[newEnvNameAnswer] = newEnvValueAnswer
				}
			case "Delete":
				switch configOps.Kind {
				case "Port":
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
						return zeroDevfile, err
					}
					selectedContainer.Ports = append(selectedContainer.Ports[:indexToDelete], selectedContainer.Ports[indexToDelete+1:]...)

				case "EnvVar":
					envToDelete := configOps.Key
					if _, ok := selectedContainer.Envs[envToDelete]; !ok {
						log.Warningf(fmt.Sprintf("unable to delete env %q, not found", envToDelete))
					}
					err = devfileobj.Data.RemoveEnvVars(map[string][]string{selectContainerAnswer: {envToDelete}})
					if err != nil {
						return zeroDevfile, err
					}
					delete(selectedContainer.Envs, envToDelete)
				}
			case "Nothing":
			default:
				return zeroDevfile, fmt.Errorf("unknown configuration selected %q", fmt.Sprintf("%v %v %v", configOps.Ops, configOps.Kind, configOps.Key))
			}
			// Update the current configuration
			config[selectContainerAnswer] = selectedContainer
		}
	}
	return devfileobj, nil
}

func (o *InteractiveBackend) HandleApplicationPorts(devfileobj parser.DevfileObj, ports []int, flags map[string]string) (parser.DevfileObj, error) {
	return handleApplicationPorts(log.GetStdout(), devfileobj, ports)
}

func PrintConfiguration(config asker.DevfileConfiguration) {

	var keys []string
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		container := config[key]
		log.Sectionf("Container Configuration %q:", key)

		stdout := log.GetStdout()

		fmt.Fprintf(stdout, "  OPEN PORTS:\n")

		for _, value := range container.Ports {
			fmt.Fprintf(stdout, "    - %s\n", value)
		}

		fmt.Fprintf(stdout, "  ENVIRONMENT VARIABLES:\n")

		for key, value := range container.Envs {
			fmt.Fprintf(stdout, "    - %s = %s\n", key, value)
		}

	}

	// Make sure we add a newline at the end
	fmt.Println()
}

func getPortsAndEnvVar(obj parser.DevfileObj) (asker.DevfileConfiguration, error) {
	var config = asker.DevfileConfiguration{}
	components, err := obj.Data.GetComponents(parsercommon.DevfileOptions{ComponentOptions: parsercommon.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType}})
	if err != nil {
		return config, err
	}
	for _, component := range components {
		var ports = []string{}
		var envMap = map[string]string{}

		for _, ep := range component.Container.Endpoints {
			ports = append(ports, strconv.Itoa(ep.TargetPort))
		}
		for _, env := range component.Container.Env {
			envMap[env.Name] = env.Value
		}
		config[component.Name] = asker.ContainerConfiguration{
			Ports: ports,
			Envs:  envMap,
		}
	}
	return config, nil
}
