package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	apidevfile "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"

	"k8s.io/utils/pointer"

	"github.com/feloy/devfile-builder/wasm/pkg/global"
)

const (
	SEPARATOR = ","
)

// getContent returns the YAML content of the global devfile as string
func GetContent() (map[string]interface{}, error) {
	err := global.Devfile.WriteYamlDevfile()
	if err != nil {
		return nil, errors.New("error writing file")
	}
	result, err := global.FS.ReadFile("/devfile.yaml")
	if err != nil {
		return nil, errors.New("error reading file")
	}

	commands, err := getCommands()
	if err != nil {
		return nil, errors.New("error getting commands")
	}

	containers, err := getContainers()
	if err != nil {
		return nil, errors.New("error getting containers")
	}

	images, err := getImages()
	if err != nil {
		return nil, errors.New("error getting images")
	}

	resources, err := getResources()
	if err != nil {
		return nil, errors.New("error getting Kubernetes resources")
	}

	return map[string]interface{}{
		"content":    string(result),
		"metadata":   getMetadata(),
		"commands":   commands,
		"events":     getEvents(),
		"containers": containers,
		"images":     images,
		"resources":  resources,
	}, nil
}

func getMetadata() map[string]interface{} {
	metadata := global.Devfile.Data.GetMetadata()
	return map[string]interface{}{
		"name":              metadata.Name,
		"version":           metadata.Version,
		"displayName":       metadata.DisplayName,
		"description":       metadata.Description,
		"tags":              strings.Join(metadata.Tags, SEPARATOR),
		"architectures":     joinArchitectures(metadata.Architectures),
		"icon":              metadata.Icon,
		"globalMemoryLimit": metadata.GlobalMemoryLimit,
		"projectType":       metadata.ProjectType,
		"language":          metadata.Language,
		"website":           metadata.Website,
		"provider":          metadata.Provider,
		"supportUrl":        metadata.SupportUrl,
	}
}

func getCommands() ([]interface{}, error) {
	commands, err := global.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, len(commands))
	for _, command := range commands {
		newCommand := map[string]interface{}{
			"name":    command.Id,
			"group":   GetGroup(command),
			"default": GetDefault(command),
		}

		if command.Exec != nil {
			newCommand["type"] = "exec"
			newCommand["exec"] = map[string]interface{}{
				"component":        command.Exec.Component,
				"commandLine":      command.Exec.CommandLine,
				"workingDir":       command.Exec.WorkingDir,
				"hotReloadCapable": pointer.BoolDeref(command.Exec.HotReloadCapable, false),
			}
		}

		if command.Apply != nil {
			components, err := global.Devfile.Data.GetComponents(common.DevfileOptions{
				FilterByName: command.Apply.Component,
			})
			if err != nil {
				return nil, err
			}
			if len(components) == 0 {
				return nil, fmt.Errorf("component %q not found", command.Apply.Component)
			}
			component := components[0]
			if component.Kubernetes != nil || component.Openshift != nil {
				newCommand["type"] = "apply"
				newCommand["apply"] = map[string]interface{}{
					"component": command.Apply.Component,
				}
			}
			if component.Image != nil {
				newCommand["type"] = "image"
				newCommand["image"] = map[string]interface{}{
					"component": command.Apply.Component,
				}
			}

		}

		if command.Composite != nil {
			commands := make([]interface{}, 0, len(command.Composite.Commands))
			for _, cmd := range command.Composite.Commands {
				commands = append(commands, cmd)
			}
			newCommand["type"] = "composite"
			newCommand["composite"] = map[string]interface{}{
				"commands": commands,
				"parallel": pointer.BoolDeref(command.Composite.Parallel, false),
			}
		}
		result = append(result, newCommand)
	}
	return result, nil
}

func getEvents() map[string]interface{} {
	events := global.Devfile.Data.GetEvents()

	return map[string]interface{}{
		"preStart":  StringArrayToInterfaceArray(events.PreStart),
		"postStart": StringArrayToInterfaceArray(events.PostStart),
		"preStop":   StringArrayToInterfaceArray(events.PreStop),
		"postStop":  StringArrayToInterfaceArray(events.PostStop),
	}
}

func getContainers() ([]interface{}, error) {
	containers, err := global.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ContainerComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, len(containers))
	for _, container := range containers {
		commands := make([]interface{}, len(container.ComponentUnion.Container.Command))
		for i, command := range container.ComponentUnion.Container.Command {
			commands[i] = command
		}

		args := make([]interface{}, len(container.ComponentUnion.Container.Args))
		for i, arg := range container.ComponentUnion.Container.Args {
			args[i] = arg
		}

		result = append(result, map[string]interface{}{
			"name":          container.Name,
			"image":         container.ComponentUnion.Container.Image,
			"command":       commands,
			"args":          args,
			"memoryRequest": container.ComponentUnion.Container.MemoryRequest,
			"memoryLimit":   container.ComponentUnion.Container.MemoryLimit,
			"cpuRequest":    container.ComponentUnion.Container.CpuRequest,
			"cpuLimit":      container.ComponentUnion.Container.CpuLimit,
		})
	}
	return result, nil
}

func getImages() ([]interface{}, error) {
	images, err := global.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ImageComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, len(images))
	for _, image := range images {

		args := make([]interface{}, len(image.Image.Dockerfile.Args))
		for i, arg := range image.Image.Dockerfile.Args {
			args[i] = arg
		}

		result = append(result, map[string]interface{}{
			"name":         image.Name,
			"imageName":    image.Image.ImageName,
			"args":         args,
			"buildContext": image.Image.Dockerfile.BuildContext,
			"rootRequired": pointer.BoolDeref(image.Image.Dockerfile.RootRequired, false),
			"uri":          image.Image.Dockerfile.Uri,
		})
	}
	return result, nil
}

func getResources() ([]interface{}, error) {
	resources, err := global.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.KubernetesComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, 0, len(resources))
	for _, resource := range resources {
		result = append(result, map[string]interface{}{
			"name":    resource.Name,
			"inlined": resource.ComponentUnion.Kubernetes.Inlined,
			"uri":     resource.ComponentUnion.Kubernetes.Uri,
		})
	}
	return result, nil
}

func joinArchitectures(architectures []apidevfile.Architecture) string {
	strArchs := make([]string, len(architectures))
	for i, arch := range architectures {
		strArchs[i] = string(arch)
	}
	return strings.Join(strArchs, SEPARATOR)
}
