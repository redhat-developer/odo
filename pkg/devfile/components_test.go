package devfile

import (
	"reflect"
	"testing"

	devfiletesting "github.com/redhat-developer/odo/pkg/devfile/testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
)

func TestGetKubernetesComponentsToPush(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	getDevfileWithoutApplyCommand := func() parser.DevfileObj {
		devfileObj := parser.DevfileObj{
			Data: devfiletesting.GetDevfileData(t, []devfiletesting.InlinedComponent{
				{
					Name:    "component1",
					Inlined: "Component 1",
				},
			}, nil),
			Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		}
		return devfileObj
	}

	getDevfileWithApplyCommand := func(applyComponentName string) parser.DevfileObj {
		devfileObj := parser.DevfileObj{
			Data: devfiletesting.GetDevfileData(t, []devfiletesting.InlinedComponent{
				{
					Name:    "component1",
					Inlined: "Component 1",
				},
			}, nil),
			Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		}
		applyCommand := devfilev1.Command{
			CommandUnion: devfilev1.CommandUnion{
				Apply: &devfilev1.ApplyCommand{
					Component: applyComponentName,
				},
			},
		}
		_ = devfileObj.Data.AddCommands([]devfilev1.Command{applyCommand})
		return devfileObj
	}

	tests := []struct {
		name       string
		devfileObj parser.DevfileObj
		want       []devfilev1.Component
		wantErr    bool
	}{
		{
			name: "empty devfile",
			devfileObj: parser.DevfileObj{
				Data: devfiletesting.GetDevfileData(t, nil, nil),
				Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
			},
			want:    []devfilev1.Component{},
			wantErr: false,
		},
		{
			name:       "no apply command",
			devfileObj: getDevfileWithoutApplyCommand(),
			want: []devfilev1.Component{
				{
					Name: "component1",
					ComponentUnion: devfilev1.ComponentUnion{
						Kubernetes: &devfilev1.KubernetesComponent{
							K8sLikeComponent: devfilev1.K8sLikeComponent{
								K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
									Inlined: "Component 1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "apply command referencing the component",
			devfileObj: getDevfileWithApplyCommand("component1"),
			want:       []devfilev1.Component{},
			wantErr:    false,
		},
		{
			name:       "apply command not referencing the component",
			devfileObj: getDevfileWithApplyCommand("other component"),
			want: []devfilev1.Component{
				{
					Name: "component1",
					ComponentUnion: devfilev1.ComponentUnion{
						Kubernetes: &devfilev1.KubernetesComponent{
							K8sLikeComponent: devfilev1.K8sLikeComponent{
								K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
									Inlined: "Component 1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKubernetesComponentsToPush(tt.devfileObj)
			gotErr := err != nil
			if len(got) != len(tt.want) {
				t.Errorf("Got %d components, expected %d\n", len(got), len(tt.want))
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nGot      %+v\nExpected %+v\n", got, tt.want)
			}
			if gotErr != tt.wantErr {
				t.Errorf("Got error %v, expected %v\n", gotErr, tt.wantErr)
			}
		})
	}
}
