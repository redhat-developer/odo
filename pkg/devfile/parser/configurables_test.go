package parser

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openshift/odo/pkg/config"
	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestSetConfiguration(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		args           map[string]string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: set memory to 500Mi",
			args: map[string]string{
				"memory": "500Mi",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image:       "quay.io/nodejs-12",
								MemoryLimit: "500Mi",
								Endpoints: []common.Endpoint{
									{
										Name:       "port-3030",
										TargetPort: 3000,
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							Container: &common.Container{
								Image:       "quay.io/nginx",
								MemoryLimit: "500Mi",
							},
						},
					},
					Events: common.DevfileEvents{
						PostStop: []string{"post-stop"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/projects",
							Name:      "starter-project-2",
						},
					},
				},
			},
		},
		{
			name: "case 2: set ports array",
			args: map[string]string{
				"ports": "8080,8081/UDP,8080/TCP",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
								Endpoints: []common.Endpoint{
									{
										Name:       "port-3030",
										TargetPort: 3000,
									},
									{
										Name:       "port-8080-tcp",
										TargetPort: 8080,
										Protocol:   "tcp",
									}, {
										Name:       "port-8081-udp",
										TargetPort: 8081,
										Protocol:   "udp",
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							Container: &common.Container{
								Image: "quay.io/nginx",
								Endpoints: []common.Endpoint{
									{
										Name:       "port-8080-tcp",
										TargetPort: 8080,
										Protocol:   "tcp",
									}, {
										Name:       "port-8081-udp",
										TargetPort: 8081,
										Protocol:   "udp",
									},
								},
							},
						},
					},
					Events: common.DevfileEvents{
						PostStop: []string{"post-stop"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/projects",
							Name:      "starter-project-2",
						},
					},
				},
			},
		},
		{
			name: "case 3: set ports array fails due to validation",
			args: map[string]string{
				"ports": "8080,8081/UDP,8083/",
			},
			currentDevfile: testDevfileObj(fs),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for key, value := range tt.args {
				err := tt.currentDevfile.SetConfiguration(key, value)
				if tt.wantErr {
					if err == nil {
						t.Errorf("expected error but got nil")
					}
					// we dont expect an error here
				} else {
					if err != nil {
						t.Errorf("error while setting configuration %+v", err.Error())
					}
				}
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
					t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
				}
			}

		})
	}

}

func TestAddAndRemoveEnvVars(t *testing.T) {

	// Use fakeFs
	fs := filesystem.NewFakeFs()

	tests := []struct {
		name           string
		listToAdd      config.EnvVarList
		listToRemove   []string
		currentDevfile DevfileObj
		wantDevFile    DevfileObj
	}{
		{
			name: "case 1: add and remove env vars",
			listToAdd: config.EnvVarList{
				{
					Name:  "DATABASE_PASSWORD",
					Value: "苦痛",
				},
				{
					Name:  "PORT",
					Value: "3003",
				},
				{
					Name:  "PORT",
					Value: "4342",
				},
			},
			listToRemove: []string{
				"PORT",
			},
			currentDevfile: testDevfileObj(fs),
			wantDevFile: DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
								Endpoints: []common.Endpoint{
									{
										Name:       "port-3030",
										TargetPort: 3000,
									},
								},
								Env: []common.Env{
									{
										Name:  "DATABASE_PASSWORD",
										Value: "苦痛",
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							Container: &common.Container{
								Image: "quay.io/nginx",
								Env: []common.Env{
									{
										Name:  "DATABASE_PASSWORD",
										Value: "苦痛",
									},
								},
							},
						},
					},
					Events: common.DevfileEvents{
						PostStop: []string{"post-stop"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/projects",
							Name:      "starter-project-2",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.currentDevfile.AddEnvVars(tt.listToAdd)

			if err != nil {
				t.Errorf("error while adding env vars %+v", err.Error())
			}

			err = tt.currentDevfile.RemoveEnvVars(tt.listToRemove)

			if err != nil {
				t.Errorf("error while removing env vars %+v", err.Error())
			}

			if !reflect.DeepEqual(tt.currentDevfile.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.currentDevfile, pretty.Compare(tt.currentDevfile.Data, tt.wantDevFile.Data))
			}

		})
	}

}

func testDevfileObj(fs filesystem.Filesystem) DevfileObj {
	return DevfileObj{
		Ctx: devfileCtx.FakeContext(fs, OutputDevfileYamlPath),
		Data: &v200.Devfile200{
			Commands: []common.DevfileCommand{
				{
					Id: "devbuild",
					Exec: &common.Exec{
						WorkingDir: "/projects/nodejs-starter",
					},
				},
			},
			Components: []common.DevfileComponent{
				{
					Name: "runtime",
					Container: &common.Container{
						Image: "quay.io/nodejs-12",
						Endpoints: []common.Endpoint{
							{
								Name:       "port-3030",
								TargetPort: 3000,
							},
						},
					},
				},
				{
					Name: "loadbalancer",
					Container: &common.Container{
						Image: "quay.io/nginx",
					},
				},
			},
			Events: common.DevfileEvents{
				PostStop: []string{"post-stop"},
			},
			Projects: []common.DevfileProject{
				{
					ClonePath: "/projects",
					Name:      "nodejs-starter-build",
				},
			},
			StarterProjects: []common.DevfileStarterProject{
				{
					ClonePath: "/projects",
					Name:      "starter-project-2",
				},
			},
		},
	}
}
