package config

import (
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/kylelemons/godebug/pretty"
	odoTestingUtil "github.com/openshift/odo/pkg/testingutil"
)

func TestSetDevfileConfiguration(t *testing.T) {

	// Use fakeFs
	fs := devfilefs.NewFakeFs()

	tests := []struct {
		name           string
		args           map[string]string
		currentDevfile parser.DevfileObj
		wantDevFile    parser.DevfileObj
		wantErr        bool
	}{
		{
			name: "case 1: set memory to 500Mi",
			args: map[string]string{
				"memory": "500Mi",
			},
			currentDevfile: odoTestingUtil.GetTestDevfileObj(fs),
			wantDevFile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image:       "quay.io/nodejs-12",
										MemoryLimit: "500Mi",
									},
									Endpoints: []devfilev1.Endpoint{
										{
											Name:       "port-3030",
											TargetPort: 3000,
										},
									},
								},
							},
						},
						{
							Name: "loadbalancer",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image:       "quay.io/nginx",
										MemoryLimit: "500Mi",
									},
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands([]devfilev1.Command{
						{
							Id: "devbuild",
							CommandUnion: devfilev1.CommandUnion{
								Exec: &devfilev1.ExecCommand{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
		},
		{
			name: "case 2: set ports array",
			args: map[string]string{
				"ports": "8080,8081/UDP,8080/TCP",
			},
			currentDevfile: odoTestingUtil.GetTestDevfileObj(fs),
			wantDevFile: parser.DevfileObj{
				Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands([]devfilev1.Command{
						{
							Id: "devbuild",
							CommandUnion: devfilev1.CommandUnion{
								Exec: &devfilev1.ExecCommand{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{
						{
							Name: "runtime",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nodejs-12",
									},
									Endpoints: []devfilev1.Endpoint{
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
						},
						{
							Name: "loadbalancer",
							ComponentUnion: devfilev1.ComponentUnion{
								Container: &devfilev1.ContainerComponent{
									Container: devfilev1.Container{
										Image: "quay.io/nginx",
									},
									Endpoints: []devfilev1.Endpoint{
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
					})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			},
		},
		{
			name: "case 3: set ports array fails due to validation",
			args: map[string]string{
				"ports": "8080,8081/UDP,8083/",
			},
			currentDevfile: odoTestingUtil.GetTestDevfileObj(fs),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for key, value := range tt.args {
				err := SetDevfileConfiguration(tt.currentDevfile, key, value)
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
