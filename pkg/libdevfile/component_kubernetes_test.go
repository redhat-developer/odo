package libdevfile

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
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

func TestGetK8sAndOcComponentsToPush(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	buildK8sOrOcComponent := func(k8s bool, name string, deployByDefault *bool, referenced bool) (v1alpha2.Component, v1alpha2.Command) {
		k8sLikeComponent := v1alpha2.K8sLikeComponent{
			DeployByDefault: deployByDefault,
			K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
				Inlined: name,
			},
		}
		comp := v1alpha2.Component{Name: name}
		if k8s {
			comp.ComponentUnion.Kubernetes = &v1alpha2.KubernetesComponent{K8sLikeComponent: k8sLikeComponent}
		} else {
			comp.ComponentUnion.Openshift = &v1alpha2.OpenshiftComponent{K8sLikeComponent: k8sLikeComponent}
		}
		if referenced {
			cmd := v1alpha2.Command{
				Id: "apply-" + name,
				CommandUnion: v1alpha2.CommandUnion{
					Apply: &v1alpha2.ApplyCommand{
						Component: name,
					},
				},
			}
			return comp, cmd
		}
		return comp, v1alpha2.Command{}
	}

	var (
		k8sDeployByDefaultTrueReferenced, applyK8sDeployByDefaultTrueReferenced = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultTrueReferenced", pointer.Bool(true), true)
		ocDeployByDefaultTrueReferenced, applyOcDeployByDefaultTrueReferenced = buildK8sOrOcComponent(
			false, "ocDeployByDefaultTrueReferenced", pointer.Bool(true), true)

		k8sDeployByDefaultTrueNotReferenced, _ = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultTrueNotReferenced", pointer.Bool(true), false)
		ocDeployByDefaultTrueNotReferenced, _ = buildK8sOrOcComponent(
			false, "ocDeployByDefaultTrueNotReferenced", pointer.Bool(true), false)

		k8sDeployByDefaultFalseReferenced, applyK8sDeployByDefaultFalseReferenced = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultFalseReferenced", pointer.Bool(false), true)
		ocDeployByDefaultFalseReferenced, applyOcDeployByDefaultFalseReferenced = buildK8sOrOcComponent(
			false, "ocDeployByDefaultFalseReferenced", pointer.Bool(false), true)

		k8sDeployByDefaultFalseNotReferenced, _ = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultFalseNotReferenced", pointer.Bool(false), false)
		ocDeployByDefaultFalseNotReferenced, _ = buildK8sOrOcComponent(
			false, "ocDeployByDefaultFalseNotReferenced", pointer.Bool(false), false)

		k8sDeployByDefaultNotSetReferenced, applyK8sDeployByDefaultNotSetReferenced = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultNotSetReferenced", nil, true)
		ocDeployByDefaultNotSetReferenced, applyOcDeployByDefaultNotSetReferenced = buildK8sOrOcComponent(
			false, "ocDeployByDefaultNotSetReferenced", nil, true)

		k8sDeployByDefaultNotSetNotReferenced, _ = buildK8sOrOcComponent(
			true, "k8sDeployByDefaultNotSetNotReferenced", nil, false)
		ocDeployByDefaultNotSetNotReferenced, _ = buildK8sOrOcComponent(
			false, "ocDeployByDefaultNotSetNotReferenced", nil, false)
	)

	buildFullDevfile := func() (parser.DevfileObj, error) {
		devfileData, err := data.NewDevfileData(string(data.APISchemaVersion220))
		if err != nil {
			return parser.DevfileObj{}, err
		}
		devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-devfile"})
		err = devfileData.AddComponents([]v1alpha2.Component{
			k8sDeployByDefaultNotSetNotReferenced,
			k8sDeployByDefaultNotSetReferenced,
			ocDeployByDefaultNotSetReferenced,
			ocDeployByDefaultNotSetNotReferenced,

			k8sDeployByDefaultTrueNotReferenced,
			k8sDeployByDefaultTrueReferenced,
			ocDeployByDefaultTrueReferenced,
			ocDeployByDefaultTrueNotReferenced,

			k8sDeployByDefaultFalseNotReferenced,
			k8sDeployByDefaultFalseReferenced,
			ocDeployByDefaultFalseReferenced,
			ocDeployByDefaultFalseNotReferenced,

			//Add other kinds of components
			{
				Name: "my-image-component",
				ComponentUnion: v1alpha2.ComponentUnion{
					Image: &v1alpha2.ImageComponent{
						Image: v1alpha2.Image{
							ImageName: "image-component-1",
						},
					},
				},
			},
			{
				Name: "container-component",
				ComponentUnion: v1alpha2.ComponentUnion{
					Container: &v1alpha2.ContainerComponent{
						Container: v1alpha2.Container{
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
		err = devfileData.AddCommands([]v1alpha2.Command{
			applyK8sDeployByDefaultNotSetReferenced,
			applyOcDeployByDefaultNotSetReferenced,
			applyK8sDeployByDefaultTrueReferenced,
			applyOcDeployByDefaultTrueReferenced,
			applyK8sDeployByDefaultFalseReferenced,
			applyOcDeployByDefaultFalseReferenced,

			//Add other kinds of components
			{
				Id: "apply-image",
				CommandUnion: v1alpha2.CommandUnion{
					Apply: &v1alpha2.ApplyCommand{
						Component: "my-image-component",
					},
				},
			},
			{
				Id: "exec-command",
				CommandUnion: v1alpha2.CommandUnion{
					Apply: &v1alpha2.ApplyCommand{
						Component: "my-image-component",
					},
					Exec: &v1alpha2.ExecCommand{
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
		allowApply bool
	}

	tests := []struct {
		name    string
		args    args
		want    []v1alpha2.Component
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
			want:    []v1alpha2.Component{},
			wantErr: false,
		},
		{
			name: "allowApply=false => return components that need to be created automatically on startup",
			args: args{
				devfileObj: buildFullDevfile,
				allowApply: false,
			},
			want: []v1alpha2.Component{
				k8sDeployByDefaultTrueNotReferenced,
				k8sDeployByDefaultTrueReferenced,
				ocDeployByDefaultTrueNotReferenced,
				ocDeployByDefaultTrueReferenced,
				k8sDeployByDefaultNotSetNotReferenced,
				ocDeployByDefaultNotSetNotReferenced,
			},
		},
		{
			name: "allowApply=true => return components that need to be created automatically on startup and those referenced",
			args: args{
				devfileObj: buildFullDevfile,
				allowApply: true,
			},
			want: []v1alpha2.Component{
				k8sDeployByDefaultTrueNotReferenced,
				k8sDeployByDefaultTrueReferenced,
				ocDeployByDefaultTrueNotReferenced,
				ocDeployByDefaultTrueReferenced,
				k8sDeployByDefaultNotSetNotReferenced,
				ocDeployByDefaultNotSetNotReferenced,

				k8sDeployByDefaultFalseReferenced,
				ocDeployByDefaultFalseReferenced,
				k8sDeployByDefaultNotSetReferenced,
				ocDeployByDefaultNotSetReferenced,
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
			got, err := GetK8sAndOcComponentsToPush(devfileObj, tt.args.allowApply)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Got error %v, expected %v\n", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Got %d components, expected %d\n", len(got), len(tt.want))
			}

			lessFn := func(x, y v1alpha2.Component) bool {
				return x.Name < y.Name
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty(), cmpopts.SortSlices(lessFn)); diff != "" {
				t.Errorf("GetK8sAndOcComponentsToPush() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
