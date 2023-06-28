package devstate

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	context "github.com/devfile/library/v2/pkg/devfile/parser/context"
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
