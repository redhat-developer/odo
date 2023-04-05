package libdevfile

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/pointer"

	devfiletesting "github.com/redhat-developer/odo/pkg/devfile/testing"
)

func TestGetImageComponentsToPush(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	buildImageComponent := func(name string, autoBuild *bool, referenced bool) (devfilev1.Component, devfilev1.Command) {
		comp := devfilev1.Component{
			Name: name,
			ComponentUnion: devfilev1.ComponentUnion{
				Image: &devfilev1.ImageComponent{
					Image: devfilev1.Image{
						ImageName: "my-image:" + name,
						ImageUnion: devfilev1.ImageUnion{
							AutoBuild: autoBuild,
						},
					},
				},
			},
		}
		if referenced {
			cmd := devfilev1.Command{
				Id: "apply-" + name,
				CommandUnion: devfilev1.CommandUnion{
					Apply: &devfilev1.ApplyCommand{
						Component: name,
					},
				},
			}
			return comp, cmd
		}
		return comp, devfilev1.Command{}
	}

	var (
		autoBuildTrueReferenced, applyAutoBuildTrueReferenced = buildImageComponent(
			"autoBuildTrueReferenced", pointer.Bool(true), true)
		autoBuildTrueNotReferenced, _ = buildImageComponent(
			"autoBuildTrueNotReferenced", pointer.Bool(true), false)
		autoBuildFalseReferenced, applyAutoBuildFalseReferenced = buildImageComponent(
			"autoBuildFalseReferenced", pointer.Bool(false), true)
		autoBuildFalseNotReferenced, _ = buildImageComponent(
			"autoBuildFalseNotReferenced", pointer.Bool(false), false)
		autoBuildNotSetReferenced, applyAutoBuildNotSetReferenced = buildImageComponent(
			"autoBuildNotSetReferenced", nil, true)
		autoBuildNotSetNotReferenced, _ = buildImageComponent(
			"autoBuildNotSetNotReferenced", nil, false)
	)

	buildFullDevfile := func() (parser.DevfileObj, error) {
		devfileData, err := data.NewDevfileData(string(data.APISchemaVersion220))
		if err != nil {
			return parser.DevfileObj{}, err
		}
		devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-devfile"})
		err = devfileData.AddComponents([]devfilev1.Component{
			autoBuildTrueReferenced,
			autoBuildTrueNotReferenced,
			autoBuildFalseReferenced,
			autoBuildFalseNotReferenced,
			autoBuildNotSetReferenced,
			autoBuildNotSetNotReferenced,

			//Add other kinds of components
			{
				Name: "my-k8s-component",
				ComponentUnion: devfilev1.ComponentUnion{
					Kubernetes: &devfilev1.KubernetesComponent{
						K8sLikeComponent: devfilev1.K8sLikeComponent{
							K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
								Inlined: "my-k8s-component-inlined",
							},
						},
					},
				},
			},
			{
				Name: "container-component",
				ComponentUnion: devfilev1.ComponentUnion{
					Container: &devfilev1.ContainerComponent{
						Container: devfilev1.Container{
							DedicatedPod: pointer.Bool(true),
							Image:        "my-container-image",
						},
					},
				},
			},
		})
		if err != nil {
			return parser.DevfileObj{}, err
		}
		err = devfileData.AddCommands([]devfilev1.Command{
			applyAutoBuildTrueReferenced,
			applyAutoBuildFalseReferenced,
			applyAutoBuildNotSetReferenced,

			//Add other kinds of components
			{
				Id: "apply-k8s-component",
				CommandUnion: devfilev1.CommandUnion{
					Apply: &devfilev1.ApplyCommand{
						Component: "my-k8s-component",
					},
				},
			},
			{
				Id: "exec-command",
				CommandUnion: devfilev1.CommandUnion{
					Apply: &devfilev1.ApplyCommand{
						Component: "my-image-component",
					},
					Exec: &devfilev1.ExecCommand{
						CommandLine: "/path/to/my/command -success",
						Component:   "container-component",
					},
				},
			},
		})
		if err != nil {
			return parser.DevfileObj{}, err
		}
		return parser.DevfileObj{
			Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			Data: devfileData,
		}, nil
	}

	type args struct {
		devfileObj func() (parser.DevfileObj, error)
	}
	tests := []struct {
		name    string
		args    args
		want    []devfilev1.Component
		wantErr bool
	}{
		{
			name: "empty devfile",
			args: args{
				devfileObj: func() (parser.DevfileObj, error) {
					return parser.DevfileObj{
						Data: devfiletesting.GetDevfileData(t, nil, nil),
						Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
					}, nil
				},
			},
			want:    []devfilev1.Component{},
			wantErr: false,
		},
		{
			name: "return components that need to be created automatically on startup",
			args: args{
				devfileObj: buildFullDevfile,
			},
			want: []devfilev1.Component{
				autoBuildTrueReferenced,
				autoBuildTrueNotReferenced,
				autoBuildNotSetNotReferenced,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devfileObj, err := tt.args.devfileObj()
			if err != nil {
				t.Errorf("unable to create Devfile object: %v", err)
				return
			}

			got, err := GetImageComponentsToPush(devfileObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageComponentsToPush() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Got %d components, expected %d\n", len(got), len(tt.want))
			}

			lessFn := func(x, y devfilev1.Component) bool {
				return x.Name < y.Name
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty(), cmpopts.SortSlices(lessFn)); diff != "" {
				t.Errorf("GetImageComponentsToPush() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
