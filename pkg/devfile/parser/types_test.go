package parser

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

const devfileTempPath = "devfile.yaml"

func TestDevfileObj_OverrideCommands(t *testing.T) {
	componentName0 := "component-0"
	overrideComponent0 := "override-component-0"

	commandLineBuild := "npm build"
	overrideBuild := "npm custom build"
	commandLineRun := "npm run"

	workingDir := "/project"
	overrideWorkingDir := "/data"

	type args struct {
		overridePatch []common.DevfileCommand
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override a command's non list/map fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								CommandLine: commandLineBuild,
								Component:   componentName0,
								Env:         nil,
								Group: &common.Group{
									IsDefault: false,
									Kind:      common.BuildCommandGroupType,
								},
								WorkingDir: workingDir,
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileCommand{
					{
						Id: "devbuild",
						Exec: &common.Exec{
							CommandLine: overrideBuild,
							Component:   overrideComponent0,
							Group: &common.Group{
								IsDefault: true,
								Kind:      common.BuildCommandGroupType,
							},
							WorkingDir: overrideWorkingDir,
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								CommandLine: overrideBuild,
								Component:   overrideComponent0,
								Group: &common.Group{
									IsDefault: true,
									Kind:      common.BuildCommandGroupType,
								},
								WorkingDir: overrideWorkingDir,
							},
						},
					},
				},
			},
		},
		{
			name: "case 2: append/override a command's list fields based on the key",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								Attributes: map[string]string{
									"key-0": "value-0",
								},
								Env: []common.Env{
									testingutil.GetFakeEnv("env-0", "value-0"),
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileCommand{
					{
						Id: "devbuild",
						Exec: &common.Exec{
							Attributes: map[string]string{
								"key-1": "value-1",
							},
							Env: []common.Env{
								testingutil.GetFakeEnv("env-0", "value-0-0"),
								testingutil.GetFakeEnv("env-1", "value-1"),
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								Attributes: map[string]string{
									"key-0": "value-0",
									"key-1": "value-1",
								},
								Env: []common.Env{
									testingutil.GetFakeEnv("env-0", "value-0-0"),
									testingutil.GetFakeEnv("env-1", "value-1"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "case 3: if multiple, override the correct command",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								CommandLine: commandLineBuild,
							},
						},
						{
							Id: "devrun",
							Exec: &common.Exec{
								CommandLine: commandLineRun,
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileCommand{
					{
						Id: "devbuild",
						Exec: &common.Exec{
							CommandLine: overrideBuild,
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								CommandLine: overrideBuild,
							},
						},
						{
							Id: "devrun",
							Exec: &common.Exec{
								CommandLine: commandLineRun,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 4: throw error if command to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								Env: []common.Env{
									testingutil.GetFakeEnv("env-0", "value-0"),
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileCommand{
					{
						Id: "devbuild-custom",
						Exec: &common.Exec{
							Env: []common.Env{
								testingutil.GetFakeEnv("env-0", "value-0-0"),
								testingutil.GetFakeEnv("env-1", "value-1"),
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideCommands(tt.args.overridePatch)

			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideCommands() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

func TestDevfileObj_OverrideComponents(t *testing.T) {

	containerImage0 := "image-0"
	containerImage1 := "image-1"

	overrideContainerImage := "image-0-override"

	type args struct {
		overridePatch []common.DevfileComponent
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override a container's non list/map fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Args:          []string{"arg-0", "arg-1"},
								Command:       []string{"cmd-0", "cmd-1"},
								Image:         containerImage0,
								MemoryLimit:   "512Mi",
								MountSources:  false,
								SourceMapping: "/source",
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileComponent{
					{
						Name: "nodejs",
						Container: &common.Container{
							Args:          []string{"arg-0-0", "arg-1-1"},
							Command:       []string{"cmd-0-0", "cmd-1-1"},
							Image:         overrideContainerImage,
							MemoryLimit:   "1Gi",
							MountSources:  true,
							SourceMapping: "/data",
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Args:          []string{"arg-0-0", "arg-1-1"},
								Command:       []string{"cmd-0-0", "cmd-1-1"},
								Image:         overrideContainerImage,
								MemoryLimit:   "1Gi",
								MountSources:  true,
								SourceMapping: "/data",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 2: append/override a command's list fields based on the key",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Endpoints: []common.Endpoint{
									{
										Attributes: map[string]string{
											"key-0": "value-0",
											"key-1": "value-1",
										},
										Name:       "endpoint-0",
										TargetPort: 8080,
									},
								},
								Env: []common.Env{
									testingutil.GetFakeEnv("env-0", "value-0"),
								},
								VolumeMounts: []common.VolumeMount{
									testingutil.GetFakeVolumeMount("volume-0", "path-0"),
								},
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileComponent{
					{
						Name: "nodejs",
						Container: &common.Container{
							Endpoints: []common.Endpoint{
								{
									Attributes: map[string]string{
										"key-1":      "value-1-1",
										"key-append": "value-append",
									},
									Name:       "endpoint-0",
									TargetPort: 9090,
								},
								{
									Attributes: map[string]string{
										"key-0": "value-0",
									},
									Name:       "endpoint-1",
									TargetPort: 3000,
								},
							},
							Env: []common.Env{
								testingutil.GetFakeEnv("env-0", "value-0-0"),
								testingutil.GetFakeEnv("env-1", "value-1"),
							},
							VolumeMounts: []common.VolumeMount{
								testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
								testingutil.GetFakeVolumeMount("volume-1", "path-1"),
							},
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Env: []common.Env{
									testingutil.GetFakeEnv("env-0", "value-0-0"),
									testingutil.GetFakeEnv("env-1", "value-1"),
								},
								VolumeMounts: []common.VolumeMount{
									testingutil.GetFakeVolumeMount("volume-0", "path-0-0"),
									testingutil.GetFakeVolumeMount("volume-1", "path-1"),
								},
								Endpoints: []common.Endpoint{
									{
										Attributes: map[string]string{
											"key-0":      "value-0",
											"key-1":      "value-1-1",
											"key-append": "value-append",
										},
										Name:       "endpoint-0",
										TargetPort: 9090,
									},
									{
										Attributes: map[string]string{
											"key-0": "value-0",
										},
										Name:       "endpoint-1",
										TargetPort: 3000,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 3: if multiple, override the correct command",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: containerImage0,
							},
						},
						{
							Name: "runtime",
							Container: &common.Container{
								Image: containerImage1,
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileComponent{
					{
						Name: "nodejs",
						Container: &common.Container{
							Image: overrideContainerImage,
						},
					},
					{
						Name: "runtime",
						Container: &common.Container{
							Image: containerImage1,
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: overrideContainerImage,
							},
						},
						{
							Name: "runtime",
							Container: &common.Container{
								Image: containerImage1,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "case 4: throw error if component to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: containerImage0,
							},
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileComponent{
					{
						Name: "nodejs-custom",
						Container: &common.Container{
							Image: containerImage0,
						},
					},
				},
			},
			wantDevFileObj: DevfileObj{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideComponents(tt.args.overridePatch)
			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideComponents() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

func TestDevfileObj_OverrideProjects(t *testing.T) {
	projectName0 := "project-0"
	projectName1 := "project-1"

	type args struct {
		overridePatch []common.DevfileProject
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		wantDevFileObj DevfileObj
		args           args
		wantErr        bool
	}{
		{
			name: "case 1: override a project's fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []common.DevfileProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName0,
							Zip:  nil,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: projectName0,
						Zip:  nil,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []common.DevfileProject{
						{
							ClonePath: "/source",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
								},
							},
							Name: projectName0,
							Zip:  nil,
						},
					},
				},
			},
		},
		{
			name: "case 2: if multiple, override the correct project",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []common.DevfileProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName0,
							Zip:  nil,
						},
						{
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: projectName0,
						Zip:  nil,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []common.DevfileProject{
						{
							ClonePath: "/source",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
								},
							},
							Name: projectName0,
							Zip:  nil,
						},
						{
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
		},
		{
			name: "case 3: throw error if project to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Projects: []common.DevfileProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName0,
							Zip:  nil,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: "custom-project",
						Zip:  nil,
					},
				},
			},
			wantDevFileObj: DevfileObj{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideProjects(tt.args.overridePatch)

			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

func TestDevfileObj_OverrideStarterProjects(t *testing.T) {
	projectName1 := "starter-1"
	projectName2 := "starter-2"

	type args struct {
		overridePatch []common.DevfileStarterProject
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		wantDevFileObj DevfileObj
		args           args
		wantErr        bool
	}{
		{
			name: "Case 1: override a starter projects fields",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileStarterProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: projectName1,
						Zip:  nil,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/source",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
								},
							},
							Name: projectName1,
						},
					},
				},
			},
		},
		{
			name: "Case 2: if multiple, override the correct starter project",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName1,
						},
						{
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName2,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileStarterProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: projectName1,
					},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/source",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
								},
							},
							Name: projectName1,
							Zip:  nil,
						},
						{
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName2,
						},
					},
				},
			},
		},
		{
			name: "Case 3: throw error if starter project to override is not found",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: projectName1,
							Zip:  nil,
						},
					},
				},
			},
			args: args{
				overridePatch: []common.DevfileStarterProject{
					{
						ClonePath: "/source",
						Github: &common.Github{
							GitLikeProjectSource: common.GitLikeProjectSource{
								Remotes:      map[string]string{"origin": "url"},
								CheckoutFrom: &common.CheckoutFrom{Revision: "release-1.0.0"},
							},
						},
						Name: "custom-starter-project",
						Zip:  nil,
					},
				},
			},
			wantDevFileObj: DevfileObj{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.devFileObj.OverrideStarterProjects(tt.args.overridePatch)

			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideStarterProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}

func TestDevfileObj_OverrideEvents(t *testing.T) {
	type args struct {
		overridePatch common.DevfileEvents
	}
	tests := []struct {
		name           string
		devFileObj     DevfileObj
		args           args
		wantDevFileObj DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: override the events",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: common.DevfileEvents{
						PostStart: []string{"post-start-0", "post-start-1"},
						PostStop:  []string{"post-stop-0", "post-stop-1"},
						PreStart:  []string{"pre-start-0", "pre-start-1"},
						PreStop:   []string{"pre-stop-0", "pre-stop-1"},
					},
				},
			},
			args: args{
				overridePatch: common.DevfileEvents{
					PostStart: []string{"override-post-start-0", "override-post-start-1"},
					PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
					PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
					PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: common.DevfileEvents{
						PostStart: []string{"override-post-start-0", "override-post-start-1"},
						PostStop:  []string{"override-post-stop-0", "override-post-stop-1"},
						PreStart:  []string{"override-pre-start-0", "override-pre-start-1"},
						PreStop:   []string{"override-pre-stop-0", "override-pre-stop-1"},
					},
				},
			},
		},
		{
			name: "case 2: override some of the events",
			devFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: common.DevfileEvents{
						PostStart: []string{"post-start-0", "post-start-1"},
						PostStop:  []string{"post-stop-0", "post-stop-1"},
					},
				},
			},
			args: args{
				overridePatch: common.DevfileEvents{
					PostStart: []string{"override-post-start-0", "override-post-start-1"},
				},
			},
			wantDevFileObj: DevfileObj{
				Ctx: devfileCtx.NewDevfileCtx(devfileTempPath),
				Data: &v200.Devfile200{
					Events: common.DevfileEvents{
						PostStart: []string{"override-post-start-0", "override-post-start-1"},
						PostStop:  []string{"post-stop-0", "post-stop-1"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.devFileObj.OverrideEvents(tt.args.overridePatch); (err != nil) != tt.wantErr {
				t.Errorf("OverrideEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.wantDevFileObj, tt.devFileObj) {
				t.Errorf("expected devfile and got devfile are different: %v", pretty.Compare(tt.wantDevFileObj, tt.devFileObj))
			}
		})
	}
}
