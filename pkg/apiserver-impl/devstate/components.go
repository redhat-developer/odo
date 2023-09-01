package devstate

import (
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"k8s.io/utils/pointer"
)

func (o *DevfileState) AddContainer(
	name string,
	image string,
	command []string,
	args []string,
	envs []Env,
	memRequest string,
	memLimit string,
	cpuRequest string,
	cpuLimit string,
	volumeMounts []VolumeMount,
	configureSources bool,
	mountSources bool,
	sourceMapping string,
	annotation Annotation,
	endpoints []Endpoint,
) (DevfileContent, error) {
	v1alpha2VolumeMounts := make([]v1alpha2.VolumeMount, 0, len(volumeMounts))
	for _, vm := range volumeMounts {
		v1alpha2VolumeMounts = append(v1alpha2VolumeMounts, v1alpha2.VolumeMount{
			Name: vm.Name,
			Path: vm.Path,
		})
	}

	v1alpha2Envs := make([]v1alpha2.EnvVar, 0, len(envs))
	for _, env := range envs {
		v1alpha2Envs = append(v1alpha2Envs, v1alpha2.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	var annotations *v1alpha2.Annotation
	if len(annotation.Deployment) > 0 || len(annotation.Service) > 0 {
		annotations = &v1alpha2.Annotation{}
		if len(annotation.Deployment) > 0 {
			annotations.Deployment = annotation.Deployment
		}
		if len(annotation.Service) > 0 {
			annotations.Service = annotation.Service
		}
	}

	v1alpha2Endpoints := make([]v1alpha2.Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpoint := endpoint
		v1alpha2Endpoints = append(v1alpha2Endpoints, v1alpha2.Endpoint{
			Name:       endpoint.Name,
			TargetPort: int(endpoint.TargetPort),
			Exposure:   v1alpha2.EndpointExposure(endpoint.Exposure),
			Protocol:   v1alpha2.EndpointProtocol(endpoint.Protocol),
			Secure:     &endpoint.Secure,
			Path:       endpoint.Path,
		})
	}

	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image:         image,
					Command:       command,
					Args:          args,
					Env:           v1alpha2Envs,
					MemoryRequest: memRequest,
					MemoryLimit:   memLimit,
					CpuRequest:    cpuRequest,
					CpuLimit:      cpuLimit,
					VolumeMounts:  v1alpha2VolumeMounts,
					Annotation:    annotations,
				},
				Endpoints: v1alpha2Endpoints,
			},
		},
	}
	if configureSources {
		container.Container.MountSources = &mountSources
		container.Container.SourceMapping = sourceMapping
	}
	err := o.Devfile.Data.AddComponents([]v1alpha2.Component{container})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) DeleteContainer(name string) (DevfileContent, error) {

	err := o.checkContainerUsed(name)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error deleting container %q: %w", name, err)
	}

	// TODO check if it is a Container, not another component

	err = o.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) checkContainerUsed(name string) error {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ExecCommandType,
		},
	})
	if err != nil {
		return err
	}
	for _, command := range commands {
		if command.Exec.Component == name {
			return fmt.Errorf("container %q is used by exec command %q", name, command.Id)
		}
	}
	return nil
}

func (o *DevfileState) AddImage(name string, imageName string, args []string, buildContext string, rootRequired bool, uri string, autoBuild string) (DevfileContent, error) {
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Image: &v1alpha2.ImageComponent{
				Image: v1alpha2.Image{
					ImageName: imageName,
					ImageUnion: v1alpha2.ImageUnion{
						Dockerfile: &v1alpha2.DockerfileImage{
							Dockerfile: v1alpha2.Dockerfile{
								Args:         args,
								BuildContext: buildContext,
								RootRequired: &rootRequired,
							},
							DockerfileSrc: v1alpha2.DockerfileSrc{
								Uri: uri,
							},
						},
					},
				},
			},
		},
	}
	if autoBuild == "never" {
		container.Image.AutoBuild = pointer.Bool(false)
	} else if autoBuild == "always" {
		container.Image.AutoBuild = pointer.Bool(true)
	}
	err := o.Devfile.Data.AddComponents([]v1alpha2.Component{container})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) DeleteImage(name string) (DevfileContent, error) {

	err := o.checkImageUsed(name)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error deleting image %q: %w", name, err)
	}

	// TODO check if it is an Image, not another component

	err = o.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) checkImageUsed(name string) error {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ApplyCommandType,
		},
	})
	if err != nil {
		return err
	}
	for _, command := range commands {
		if command.Apply.Component == name {
			return fmt.Errorf("image %q is used by Image Command %q", name, command.Id)
		}
	}
	return nil
}

func (o *DevfileState) AddResource(name string, inlined string, uri string, deployByDefault string) (DevfileContent, error) {
	if inlined != "" && uri != "" {
		return DevfileContent{}, errors.New("both inlined and uri cannot be set at the same time")
	}
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: inlined,
						Uri:     uri,
					},
				},
			},
		},
	}
	if deployByDefault == "never" {
		container.Kubernetes.DeployByDefault = pointer.Bool(false)
	} else if deployByDefault == "always" {
		container.Kubernetes.DeployByDefault = pointer.Bool(true)
	}

	err := o.Devfile.Data.AddComponents([]v1alpha2.Component{container})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) PatchResource(name string, inlined string, uri string, deployByDefault string) (DevfileContent, error) {
	if inlined != "" && uri != "" {
		return DevfileContent{}, errors.New("both inlined and uri cannot be set at the same time")
	}
	found, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.KubernetesComponentType,
		},
		FilterByName: name,
	})
	if err != nil {
		return DevfileContent{}, err
	}
	if len(found) != 1 {
		return DevfileContent{}, fmt.Errorf("%d Resource found with name %q", len(found), name)
	}

	resource := found[0]
	resource.Kubernetes.Inlined = inlined
	resource.Kubernetes.Uri = uri
	resource.Kubernetes.DeployByDefault = nil
	if deployByDefault == "never" {
		resource.Kubernetes.DeployByDefault = pointer.Bool(false)
	} else if deployByDefault == "always" {
		resource.Kubernetes.DeployByDefault = pointer.Bool(true)
	}

	err = o.Devfile.Data.UpdateComponent(resource)
	if err != nil {
		return DevfileContent{}, err
	}

	return o.GetContent()
}

func (o *DevfileState) DeleteResource(name string) (DevfileContent, error) {

	err := o.checkResourceUsed(name)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error deleting resource %q: %w", name, err)
	}
	// TODO check if it is a Resource, not another component

	err = o.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) checkResourceUsed(name string) error {
	commands, err := o.Devfile.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandType: v1alpha2.ApplyCommandType,
		},
	})
	if err != nil {
		return err
	}
	for _, command := range commands {
		if command.Apply.Component == name {
			return fmt.Errorf("resource %q is used by Apply Command %q", name, command.Id)
		}
	}
	return nil
}

func (o *DevfileState) AddVolume(name string, ephemeral bool, size string) (DevfileContent, error) {
	volume := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Volume: &v1alpha2.VolumeComponent{
				Volume: v1alpha2.Volume{
					Ephemeral: &ephemeral,
					Size:      size,
				},
			},
		},
	}
	err := o.Devfile.Data.AddComponents([]v1alpha2.Component{volume})
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) PatchVolume(name string, ephemeral bool, size string) (DevfileContent, error) {
	found, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.VolumeComponentType,
		},
		FilterByName: name,
	})
	if err != nil {
		return DevfileContent{}, err
	}
	if len(found) != 1 {
		return DevfileContent{}, fmt.Errorf("%d Volume found with name %q", len(found), name)
	}

	volume := found[0]
	volume.Volume.Ephemeral = &ephemeral
	volume.Volume.Size = size

	err = o.Devfile.Data.UpdateComponent(volume)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) DeleteVolume(name string) (DevfileContent, error) {
	err := o.checkVolumeUsed(name)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error deleting volume %q: %w", name, err)
	}
	// TODO check if it is a Volume, not another component

	err = o.Devfile.Data.DeleteComponent(name)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func (o *DevfileState) checkVolumeUsed(name string) error {
	containers, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ContainerComponentType,
		},
	})
	if err != nil {
		return err
	}
	for _, container := range containers {
		for _, mount := range container.Container.VolumeMounts {
			if mount.Name == name {
				return fmt.Errorf("volume %q is mounted by Container %q", name, container.Name)
			}
		}
	}
	return nil
}
