package parser

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kylelemons/godebug/pretty"
	parser "github.com/openshift/odo/pkg/devfile/parser/context"
	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

const schemaV200 = "2.0.0"

func Test_parseParent(t *testing.T) {
	type args struct {
		devFileObj DevfileObj
	}
	tests := []struct {
		name          string
		args          args
		parentDevFile DevfileObj
		wantDevFile   DevfileObj
		wantErr       bool
	}{
		{
			name: "case 1: it should override the requested parent's data and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Parent: common.DevfileParent{
							Commands: []common.DevfileCommand{
								{
									Id: "devrun",
									Exec: &common.Exec{
										WorkingDir: "/projects/nodejs-starter",
									},
								},
							},
							Components: []common.DevfileComponent{
								{
									Name: "nodejs",
									Container: &common.Container{
										Image: "quay.io/nodejs-12",
									},
								},
							},
							Projects: []common.DevfileProject{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter",
								},
							},
							StarterProjects: []common.DevfileStarterProject{
								{
									ClonePath: "/projects",
									Name:      "starter-project-1",
								},
							},
						},

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
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Commands: []common.DevfileCommand{
						{
							Id: "devrun",
							Exec: &common.Exec{
								WorkingDir:  "/projects",
								CommandLine: "npm run",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: "quay.io/nodejs-10",
							},
						},
					},
					Events: common.DevfileEvents{
						PostStart: []string{"post-start-0"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "nodejs-starter",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "starter-project-1",
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
						{
							Id: "devrun",
							Exec: &common.Exec{
								CommandLine: "npm run",
								WorkingDir:  "/projects/nodejs-starter",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
							},
						},
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
							},
						},
					},
					Events: common.DevfileEvents{
						PostStop:  []string{"post-stop"},
						PostStart: []string{"post-start-0"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
						{
							ClonePath: "/projects",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "nodejs-starter",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/projects",
							Name:      "starter-project-2",
						},
						{
							ClonePath: "/projects",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "starter-project-1",
						},
					},
				},
			},
		},
		{
			name: "case 2: handle a parent'data without any local override and add the local devfile's data",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
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
								Name:      "starter-project-1",
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Commands: []common.DevfileCommand{
						{
							Id: "devrun",
							Exec: &common.Exec{
								WorkingDir:  "/projects",
								CommandLine: "npm run",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: "quay.io/nodejs-10",
							},
						},
					},
					Events: common.DevfileEvents{
						PostStart: []string{"post-start-0"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "nodejs-starter",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "starter-project-2",
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
						{
							Id: "devrun",
							Exec: &common.Exec{
								CommandLine: "npm run",
								WorkingDir:  "/projects",
							},
						},
					},
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
							},
						},
						{
							Name: "nodejs",
							Container: &common.Container{
								Image: "quay.io/nodejs-10",
							},
						},
					},
					Events: common.DevfileEvents{
						PostStart: []string{"post-start-0"},
						PostStop:  []string{"post-stop"},
					},
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "nodejs-starter",
						},
					},
					StarterProjects: []common.DevfileStarterProject{
						{
							ClonePath: "/projects",
							Name:      "starter-project-1",
						},
						{
							ClonePath: "/data",
							Github: &common.Github{
								GitLikeProjectSource: common.GitLikeProjectSource{
									Remotes:      map[string]string{"origin": "url"},
									CheckoutFrom: &common.CheckoutFrom{Revision: "master"},
								},
							},
							Name: "starter-project-2",
						},
					},
				},
			},
		},
		{
			name: "case 3: it should error out when the override is invalid",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Parent: common.DevfileParent{
							Commands: []common.DevfileCommand{
								{
									Id: "devrun",
									Exec: &common.Exec{
										WorkingDir: "/projects/nodejs-starter",
									},
								},
							},
							Components: []common.DevfileComponent{
								{
									Name: "nodejs",
									Container: &common.Container{
										Image: "quay.io/nodejs-12",
									},
								},
							},
							Projects: []common.DevfileProject{
								{
									ClonePath: "/projects",
									Name:      "nodejs-starter",
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Commands:      []common.DevfileCommand{},
					Components:    []common.DevfileComponent{},
					Projects:      []common.DevfileProject{},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{},
			},
			wantErr: true,
		},
		{
			name: "case 4: error out if the same parent command is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Commands: []common.DevfileCommand{
							{
								Id: "devbuild",
								Exec: &common.Exec{
									WorkingDir: "/projects/nodejs-starter",
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Commands: []common.DevfileCommand{
						{
							Id: "devbuild",
							Exec: &common.Exec{
								WorkingDir: "/projects/nodejs-starter",
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{},
			},
			wantErr: true,
		},
		{
			name: "case 5: error out if the same parent component is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Components: []common.DevfileComponent{
							{
								Name: "runtime",
								Container: &common.Container{
									Image: "quay.io/nodejs-12",
								},
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Components: []common.DevfileComponent{
						{
							Name: "runtime",
							Container: &common.Container{
								Image: "quay.io/nodejs-12",
							},
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{},
			},
			wantErr: true,
		},
		{
			name: "case 6: error out if the same event is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Events: common.DevfileEvents{
							PostStop: []string{"post-stop"},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Events: common.DevfileEvents{
						PostStop: []string{"post-stop"},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{},
			},
			wantErr: true,
		},
		{
			name: "case 7: error out if the same project is defined again in the local devfile",
			args: args{
				devFileObj: DevfileObj{
					Ctx: parser.NewDevfileCtx(devfileTempPath),
					Data: &v200.Devfile200{
						Projects: []common.DevfileProject{
							{
								ClonePath: "/projects",
								Name:      "nodejs-starter-build",
							},
						},
					},
				},
			},
			parentDevFile: DevfileObj{
				Data: &v200.Devfile200{
					SchemaVersion: schemaV200,
					Projects: []common.DevfileProject{
						{
							ClonePath: "/projects",
							Name:      "nodejs-starter-build",
						},
					},
				},
			},
			wantDevFile: DevfileObj{
				Data: &v200.Devfile200{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := yaml.Marshal(tt.parentDevFile.Data)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				_, err = w.Write(data)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}))

			defer testServer.Close()

			parent := tt.args.devFileObj.Data.GetParent()
			parent.Uri = testServer.URL

			tt.args.devFileObj.Data.SetParent(parent)
			tt.wantDevFile.Data.SetParent(parent)

			err := parseParent(tt.args.devFileObj)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseParent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				return
			}

			if !reflect.DeepEqual(tt.args.devFileObj.Data, tt.wantDevFile.Data) {
				t.Errorf("wanted: %v, got: %v, difference at %v", tt.wantDevFile, tt.args.devFileObj, pretty.Compare(tt.args.devFileObj.Data, tt.wantDevFile.Data))
			}
		})
	}
}
