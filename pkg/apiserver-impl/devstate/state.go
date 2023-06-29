package devstate

import (
	"errors"
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/v2/pkg/testingutil/filesystem"

	"k8s.io/utils/pointer"
)

type DevfileState struct {
	Devfile parser.DevfileObj
	FS      filesystem.Filesystem
}

func NewDevfileState() DevfileState {
	s := DevfileState{
		FS: filesystem.NewFakeFs(),
	}
	// this should never fail, as the parameters are constant
	_, _ = s.SetDevfileContent(`schemaVersion: 2.2.0`)
	return s
}

// SetDevfileContent replaces the devfile with a new content
// If an error occurs, the Devfile is not modified
func (o *DevfileState) SetDevfileContent(content string) (DevfileContent, error) {
	parserArgs := parser.ParserArgs{
		Data:                          []byte(content),
		ConvertKubernetesContentInUri: pointer.Bool(false),
	}
	var err error
	devfile, _, err := devfile.ParseDevfileAndValidate(parserArgs)
	if err != nil {
		return DevfileContent{}, fmt.Errorf("error parsing devfile YAML: %w", err)
	}
	o.Devfile = devfile
	o.Devfile.Ctx = context.FakeContext(o.FS, "/devfile.yaml")
	return o.GetContent()
}

func (o *DevfileState) AddContainer(name string, image string, command []string, args []string, memRequest string, memLimit string, cpuRequest string, cpuLimit string) (DevfileContent, error) {
	container := v1alpha2.Component{
		Name: name,
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image:         image,
					Command:       command,
					Args:          args,
					MemoryRequest: memRequest,
					MemoryLimit:   memLimit,
					CpuRequest:    cpuRequest,
					CpuLimit:      cpuLimit,
				},
			},
		},
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

func (o *DevfileState) AddImage(name string, imageName string, args []string, buildContext string, rootRequired bool, uri string) (DevfileContent, error) {
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

func (o *DevfileState) AddResource(name string, inlined string, uri string) (DevfileContent, error) {
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
	err := o.Devfile.Data.AddComponents([]v1alpha2.Component{container})
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
