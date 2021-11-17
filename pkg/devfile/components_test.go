package devfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/testingutil"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
)

func TestGetKubernetesComponentsToPush(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	getDevfileWithoutApplyCommand := func() parser.DevfileObj {
		devfileObj := parser.DevfileObj{
			Data: GetDevfileData(t, []InlinedComponent{
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
			Data: GetDevfileData(t, []InlinedComponent{
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
				Data: GetDevfileData(t, nil, nil),
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

func TestAddKubernetesComponentToDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		crd        string
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Add service CRD to devfile.yaml",
			args: args{
				crd:  "test CRD",
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: GetDevfileData(t, nil, nil),
					Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want: []v1alpha2.Component{{
				Name: "testName",
				ComponentUnion: devfilev1.ComponentUnion{
					Kubernetes: &devfilev1.KubernetesComponent{
						K8sLikeComponent: devfilev1.K8sLikeComponent{
							BaseComponent: devfilev1.BaseComponent{},
							K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
								Inlined: "test CRD",
							},
						},
					},
				},
			},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddKubernetesComponentToDevfile(tt.args.crd, tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("AddKubernetesComponentToDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addKubernetesComponent(t *testing.T) {

	type args struct {
		crd               string
		name              string
		componentContext  string
		devfileObj        parser.DevfileObj
		fs                devfileFileSystem.Filesystem
		uriFolderExists   bool
		fileAlreadyExists bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: the uri folder doesn't exist",
			args: args{
				crd:              "example",
				name:             "redis-service",
				componentContext: "/",
			},
		},
		{
			name: "case 2: the uri folder exist",
			args: args{
				crd:             "example",
				name:            "redis-service",
				uriFolderExists: true,
			},
		},
		{
			name: "case 3: the file already exists",
			args: args{
				crd:               "example",
				name:              "redis-service",
				uriFolderExists:   true,
				fileAlreadyExists: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := devfileFileSystem.NewFakeFs()
			tt.args.devfileObj = testingutil.GetTestDevfileObj(fs)
			tt.args.fs = fs

			if tt.args.uriFolderExists || tt.args.fileAlreadyExists {
				err := fs.MkdirAll(UriFolder, os.ModePerm)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				defer os.RemoveAll(UriFolder)
			}

			if tt.args.fileAlreadyExists {
				testFileName, err := fs.Create(filepath.Join(UriFolder, filePrefix+tt.args.name+".yaml"))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				defer os.RemoveAll(testFileName.Name())
			}

			if err := addKubernetesComponent(tt.args.crd, tt.args.name, tt.args.componentContext, tt.args.devfileObj, tt.args.fs); (err != nil) != tt.wantErr {
				t.Errorf("addKubernetesComponent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteKubernetesComponentFromDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	testFolderName := "someFolder"
	testFileName, err := SetupTestFolder(testFolderName, fs)
	if err != nil {
		t.Errorf("unexpected error : %v", err)
		return
	}

	type args struct {
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Remove a CRD from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: GetDevfileData(t, []InlinedComponent{
						{
							Name:    "testName",
							Inlined: "test CRD",
						},
					}, nil),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
		{
			name: "Case 2: Remove a uri based component from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]v1alpha2.Component{{
							Name: "testName",
							ComponentUnion: devfilev1.ComponentUnion{
								Kubernetes: &devfilev1.KubernetesComponent{
									K8sLikeComponent: devfilev1.K8sLikeComponent{
										BaseComponent: devfilev1.BaseComponent{},
										K8sLikeComponentLocation: devfilev1.K8sLikeComponentLocation{
											Uri: filepath.Join(UriFolder, filepath.Base(testFileName.Name())),
										},
									},
								},
							},
						},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteKubernetesComponentFromDevfile(tt.args.name, tt.args.devfileObj, testFolderName, fs); (err != nil) != tt.wantErr {
				t.Errorf("DeleteKubernetesComponentFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
