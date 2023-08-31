package devstate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"k8s.io/utils/pointer"
)

const (
	SEPARATOR = ","
)

// getContent returns the YAML content of the global devfile as string
func (o *DevfileState) GetContent() (DevfileContent, error) {
	err := o.Devfile.WriteYamlDevfile()
	if err != nil {
		return DevfileContent{}, errors.New("error writing file")
	}
	result, err := o.FS.ReadFile("/devfile.yaml")
	if err != nil {
		return DevfileContent{}, errors.New("error reading file")
	}

	commands, err := o.getCommands()
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error getting commands: %w", err)
	}

	containers, err := o.getContainers()
	if err != nil {
		return DevfileContent{}, errors.New("error getting containers")
	}
	images, err := o.getImages()
	if err != nil {
		return DevfileContent{}, errors.New("error getting images")
	}

	resources, err := o.getResources()
	if err != nil {
		return DevfileContent{}, errors.New("error getting Kubernetes resources")
	}

	volumes, err := o.getVolumes()
	if err != nil {
		return DevfileContent{}, errors.New("error getting volumes")
	}

	return DevfileContent{
		Content:    string(result),
		Commands:   commands,
		Containers: containers,
		Images:     images,
		Resources:  resources,
		Volumes:    volumes,
		Events:     o.getEvents(),
		Metadata:   o.getMetadata(),
	}, nil
}

func (o *DevfileState) getMetadata() Metadata {
	metadata := o.Devfile.Data.GetMetadata()
	return Metadata{
		Name:              metadata.Name,
		Version:           metadata.Version,
		DisplayName:       metadata.DisplayName,
		Description:       metadata.Description,
		Tags:              strings.Join(metadata.Tags, SEPARATOR),
		Architectures:     joinArchitectures(metadata.Architectures),
		Icon:              metadata.Icon,
		GlobalMemoryLimit: metadata.GlobalMemoryLimit,
		ProjectType:       metadata.ProjectType,
		Language:          metadata.Language,
		Website:           metadata.Website,
		Provider:          metadata.Provider,
		SupportUrl:        metadata.SupportUrl,
	}
}

func joinArchitectures(architectures []devfile.Architecture) string {
	strArchs := make([]string, len(architectures))
	for i, arch := range architectures {
		strArchs[i] = string(arch)
	}
	return strings.Join(strArchs, SEPARATOR)
}

func (o *DevfileState) getCommands() ([]Command, error) {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	result := make([]Command, 0, len(commands))
	for _, command := range commands {
		newCommand := Command{
			Name:    command.Id,
			Group:   GetGroup(command),
			Default: GetDefault(command),
		}

		if command.Exec != nil {
			newCommand.Type = "exec"
			newCommand.Exec = ExecCommand{
				Component:        command.Exec.Component,
				CommandLine:      command.Exec.CommandLine,
				WorkingDir:       command.Exec.WorkingDir,
				HotReloadCapable: pointer.BoolDeref(command.Exec.HotReloadCapable, false),
			}
		}

		if command.Apply != nil {
			components, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
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
				newCommand.Type = "apply"
				newCommand.Apply = ApplyCommand{
					Component: command.Apply.Component,
				}
			}
			if component.Image != nil {
				newCommand.Type = "image"
				newCommand.Image = ImageCommand{
					Component: command.Apply.Component,
				}
			}
		}

		if command.Composite != nil {
			newCommand.Type = "composite"
			newCommand.Composite = CompositeCommand{
				Commands: command.Composite.Commands,
				Parallel: pointer.BoolDeref(command.Composite.Parallel, false),
			}
		}
		result = append(result, newCommand)
	}
	return result, nil
}

func (o *DevfileState) getContainers() ([]Container, error) {
	containers, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ContainerComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]Container, 0, len(containers))
	for _, container := range containers {
		result = append(result, Container{
			Name:             container.Name,
			Image:            container.ComponentUnion.Container.Image,
			Command:          container.ComponentUnion.Container.Command,
			Args:             container.ComponentUnion.Container.Args,
			MemoryRequest:    container.ComponentUnion.Container.MemoryRequest,
			MemoryLimit:      container.ComponentUnion.Container.MemoryLimit,
			CpuRequest:       container.ComponentUnion.Container.CpuRequest,
			CpuLimit:         container.ComponentUnion.Container.CpuLimit,
			VolumeMounts:     o.getVolumeMounts(container.Container.Container),
			Annotation:       o.getAnnotation(container.Container.Annotation),
			Endpoints:        o.getEndpoints(container.Container.Endpoints),
			Env:              o.getEnv(container.Container.Env),
			ConfigureSources: container.Container.MountSources != nil,
			MountSources:     pointer.BoolDeref(container.Container.MountSources, true), // TODO(feloy) default value will depend on dedicatedPod
			SourceMapping:    container.Container.SourceMapping,
		})
	}
	return result, nil
}

func (o *DevfileState) getVolumeMounts(container v1alpha2.Container) []VolumeMount {
	result := make([]VolumeMount, 0, len(container.VolumeMounts))
	for _, vm := range container.VolumeMounts {
		result = append(result, VolumeMount{
			Name: vm.Name,
			Path: vm.Path,
		})
	}
	return result
}

func (o *DevfileState) getAnnotation(annotation *v1alpha2.Annotation) Annotation {
	if annotation == nil {
		return Annotation{}
	}
	return Annotation{
		Deployment: annotation.Deployment,
		Service:    annotation.Service,
	}
}

func (o *DevfileState) getEndpoints(endpoints []v1alpha2.Endpoint) []Endpoint {
	result := make([]Endpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		result = append(result, Endpoint{
			Name:       ep.Name,
			Exposure:   string(ep.Exposure),
			Path:       ep.Path,
			Protocol:   string(ep.Protocol),
			Secure:     pointer.BoolDeref(ep.Secure, false),
			TargetPort: int32(ep.TargetPort),
		})
	}
	return result
}

func (o *DevfileState) getEnv(envs []v1alpha2.EnvVar) []Env {
	result := make([]Env, 0, len(envs))
	for _, env := range envs {
		result = append(result, Env{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return result
}

func (o *DevfileState) getImages() ([]Image, error) {
	allApplyCommands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ApplyCommandType,
		},
	})
	if err != nil {
		return nil, err
	}

	images, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ImageComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]Image, 0, len(images))
	for _, image := range images {
		result = append(result, Image{
			Name:         image.Name,
			ImageName:    image.Image.ImageName,
			Args:         image.Image.Dockerfile.Args,
			BuildContext: image.Image.Dockerfile.BuildContext,
			RootRequired: pointer.BoolDeref(image.Image.Dockerfile.RootRequired, false),
			Uri:          image.Image.Dockerfile.Uri,
			AutoBuild:    getAutoBuild(image.Image.AutoBuild),
			Orphan:       !libdevfile.IsComponentReferenced(allApplyCommands, image.Name),
		})
	}
	return result, nil
}

func getAutoBuild(v *bool) string {
	if v == nil {
		return "undefined"
	}
	if *v {
		return "always"
	}
	return "never"
}

func (o *DevfileState) getResources() ([]Resource, error) {
	allApplyCommands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ApplyCommandType,
		},
	})
	if err != nil {
		return nil, err
	}

	resources, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.KubernetesComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]Resource, 0, len(resources))
	for _, resource := range resources {
		result = append(result, Resource{
			Name:            resource.Name,
			Inlined:         resource.ComponentUnion.Kubernetes.Inlined,
			Uri:             resource.ComponentUnion.Kubernetes.Uri,
			DeployByDefault: pointer.BoolDeref(resource.ComponentUnion.Kubernetes.DeployByDefault, false),
			Orphan:          !libdevfile.IsComponentReferenced(allApplyCommands, resource.Name),
		})
	}
	return result, nil
}

func (o *DevfileState) getVolumes() ([]Volume, error) {
	volumes, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.VolumeComponentType,
		},
	})
	if err != nil {
		return nil, err
	}
	result := make([]Volume, 0, len(volumes))
	for _, volume := range volumes {
		result = append(result, Volume{
			Name:      volume.Name,
			Ephemeral: *volume.Volume.Ephemeral,
			Size:      volume.Volume.Size,
		})
	}
	return result, nil
}

func (o *DevfileState) getEvents() Events {
	events := o.Devfile.Data.GetEvents()
	return Events{
		PreStart:  events.PreStart,
		PostStart: events.PostStart,
		PreStop:   events.PreStop,
		PostStop:  events.PostStop,
	}
}
