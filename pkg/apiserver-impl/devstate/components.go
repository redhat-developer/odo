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

	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image:         image,
					Command:       command,
					Args:          args,
					Env:           tov1alpha2EnvVars(envs),
					MemoryRequest: memRequest,
					MemoryLimit:   memLimit,
					CpuRequest:    cpuRequest,
					CpuLimit:      cpuLimit,
					VolumeMounts:  tov1alpha2VolumeMounts(volumeMounts),
					Annotation:    tov1alpha2Annotation(annotation),
				},
				Endpoints: tov1alpha2Endpoints(endpoints),
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

func (o *DevfileState) PatchContainer(
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
	found, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ContainerComponentType,
		},
		FilterByName: name,
	})
	if err != nil {
		return DevfileContent{}, err
	}
	if len(found) != 1 {
		return DevfileContent{}, fmt.Errorf("%d Container found with name %q", len(found), name)
	}

	container := found[0]
	container.Container.Image = image
	container.Container.Command = command
	container.Container.Args = args
	container.Container.Env = tov1alpha2EnvVars(envs)
	container.Container.MemoryRequest = memRequest
	container.Container.MemoryLimit = memLimit
	container.Container.CpuRequest = cpuRequest
	container.Container.CpuLimit = cpuLimit
	container.Container.VolumeMounts = tov1alpha2VolumeMounts(volumeMounts)

	container.Container.MountSources = nil
	container.Container.SourceMapping = ""
	if configureSources {
		container.Container.MountSources = &mountSources
		container.Container.SourceMapping = sourceMapping
	}
	container.Container.Annotation = tov1alpha2Annotation(annotation)
	container.Container.Endpoints = tov1alpha2Endpoints(endpoints)

	err = o.Devfile.Data.UpdateComponent(container)
	if err != nil {
		return DevfileContent{}, err
	}
	return o.GetContent()
}

func tov1alpha2EnvVars(envs []Env) []v1alpha2.EnvVar {
	result := make([]v1alpha2.EnvVar, 0, len(envs))
	for _, env := range envs {
		result = append(result, v1alpha2.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return result
}

func tov1alpha2VolumeMounts(volumeMounts []VolumeMount) []v1alpha2.VolumeMount {
	result := make([]v1alpha2.VolumeMount, 0, len(volumeMounts))
	for _, vm := range volumeMounts {
		result = append(result, v1alpha2.VolumeMount{
			Name: vm.Name,
			Path: vm.Path,
		})
	}
	return result
}

func tov1alpha2Annotation(annotation Annotation) *v1alpha2.Annotation {
	var result *v1alpha2.Annotation
	if len(annotation.Deployment) > 0 || len(annotation.Service) > 0 {
		result = &v1alpha2.Annotation{}
		if len(annotation.Deployment) > 0 {
			result.Deployment = annotation.Deployment
		}
		if len(annotation.Service) > 0 {
			result.Service = annotation.Service
		}
	}
	return result
}

func tov1alpha2Endpoints(endpoints []Endpoint) []v1alpha2.Endpoint {
	result := make([]v1alpha2.Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpoint := endpoint
		result = append(result, v1alpha2.Endpoint{
			Name:       endpoint.Name,
			TargetPort: int(endpoint.TargetPort),
			Exposure:   v1alpha2.EndpointExposure(endpoint.Exposure),
			Protocol:   v1alpha2.EndpointProtocol(endpoint.Protocol),
			Secure:     &endpoint.Secure,
			Path:       endpoint.Path,
		})
	}
	return result
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

func (o *DevfileState) PatchImage(name string, imageName string, args []string, buildContext string, rootRequired bool, uri string, autoBuild string) (DevfileContent, error) {
	found, err := o.Devfile.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{
			ComponentType: v1alpha2.ImageComponentType,
		},
		FilterByName: name,
	})
	if err != nil {
		return DevfileContent{}, err
	}
	if len(found) != 1 {
		return DevfileContent{}, fmt.Errorf("%d Image found with name %q", len(found), name)
	}

	image := found[0]
	if image.Image == nil {
		image.Image = &v1alpha2.ImageComponent{}
	}
	image.Image.ImageName = imageName
	if image.Image.Dockerfile == nil {
		image.Image.Dockerfile = &v1alpha2.DockerfileImage{}
	}
	image.Image.Dockerfile.Args = args
	image.Image.Dockerfile.BuildContext = buildContext
	image.Image.Dockerfile.RootRequired = &rootRequired
	image.Image.Dockerfile.DockerfileSrc.Uri = uri
	image.Image.AutoBuild = nil
	if autoBuild == "never" {
		image.Image.AutoBuild = pointer.Bool(false)
	} else if autoBuild == "always" {
		image.Image.AutoBuild = pointer.Bool(true)
	}
	err = o.Devfile.Data.UpdateComponent(image)
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
