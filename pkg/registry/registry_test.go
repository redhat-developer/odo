package registry

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestGetDevfileRegistries(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal("Fail to create temporary config file")
	}
	defer os.Remove(tempConfigFile.Name())
	defer tempConfigFile.Close()
	_, err = tempConfigFile.Write([]byte(
		`kind: Preference
apiversion: odo.openshift.io/v1alpha1
OdoSettings:
  RegistryList:
  - Name: DefaultDevfileRegistry
    URL: https://registry.devfile.io
  - Name: CheDevfileRegistry
    URL: https://che-devfile-registry.openshift.io/`,
	))
	if err != nil {
		t.Error(err)
	}
	tempConfigFileName := tempConfigFile.Name()

	tests := []struct {
		name         string
		registryName string
		want         []api.Registry
	}{
		{
			name:         "Case 1: Test get all devfile registries",
			registryName: "",
			want: []api.Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
				{
					Name:   "DefaultDevfileRegistry",
					URL:    "https://registry.devfile.io",
					Secure: false,
				},
			},
		},
		{
			name:         "Case 2: Test get specific devfile registry",
			registryName: "CheDevfileRegistry",
			want: []api.Registry{
				{
					Name:   "CheDevfileRegistry",
					URL:    "https://che-devfile-registry.openshift.io/",
					Secure: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{
				Globalodoconfig: &tempConfigFileName,
			})
			prefClient, _ := preference.NewClient(ctx)
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient)
			got, err := catClient.GetDevfileRegistries(tt.registryName)
			if err != nil {
				t.Errorf("Error message is %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("RegistryClient.GetDevfileRegistries() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListDevfileStacks(t *testing.T) {
	// Start a local HTTP server
	// to test getting multiple devfiles via ListDevfileStacks
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Send response to be tested
		_, err := rw.Write([]byte(
			`
			[
				{
					"name": "nodejs",
					"displayName": "NodeJS Angular Web Application",
					"description": "Stack for developing NodeJS Angular Web Application",
					"tags": [
						"NodeJS",
						"Angular",
						"Alpine"
					],
					"language": "nodejs",
					"icon": "/images/angular.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/angular/devfile.yaml"
					}
				},
				{
					"name": "python",
					"displayName": "Python",
					"description": "Python Stack with Python 3.7",
					"tags": [
						"Python",
						"pip"
					],
					"language": "python",
					"icon": "/images/foobar.svg",
					"globalMemoryLimit": "2686Mi",
					"links": {
						"self": "/devfiles/python/devfile.yaml"
					}
				}
			]
			`,
		))
		if err != nil {
			t.Error(err)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	const registryName = "TestRegistry"
	tests := []struct {
		name         string
		registryName string
		devfileName  string
		filter       string
		want         DevfileStackList
	}{
		{
			name:         "Case 1: Test getting ALL registries and looking for nodejs",
			registryName: "",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "nodejs",
						DisplayName: "NodeJS Angular Web Application",
						Description: "Stack for developing NodeJS Angular Web Application",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "nodejs",
						Tags:     []string{"NodeJS", "Angular", "Alpine"},
					},
					{
						Name:        "python",
						DisplayName: "Python",
						Description: "Python Stack with Python 3.7",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "python",
						Tags:     []string{"Python", "pip"},
					},
				},
			},
		},
		{
			name:         "Case 2: Test getting from only one specific devfile and from a specific registry",
			registryName: "TestRegistry",
			devfileName:  "nodejs",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "nodejs",
						DisplayName: "NodeJS Angular Web Application",
						Description: "Stack for developing NodeJS Angular Web Application",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "nodejs",
						Tags:     []string{"NodeJS", "Angular", "Alpine"},
					},
				},
			},
		},
		{
			name:         "Case 3: Test getting a devfile using a filter from the description",
			registryName: "TestRegistry",
			filter:       "Python Stack",
			want: DevfileStackList{
				DevfileRegistries: []api.Registry{
					{
						Name:   "TestRegistry",
						URL:    server.URL,
						Secure: false,
					},
				},
				Items: []api.DevfileStack{
					{
						Name:        "python",
						DisplayName: "Python",
						Description: "Python Stack with Python 3.7",
						Registry: api.Registry{
							Name: registryName,
							URL:  server.URL,
						},
						Language: "python",
						Tags:     []string{"Python", "pip"},
					},
				},
			},
		},
		{
			name:         "Case 4: Expect nothing back if registry is not found",
			registryName: "Foobar",
			want:         DevfileStackList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryList().Return([]preference.Registry{
				{
					Name: "TestRegistry",
					URL:  server.URL,
				},
			}).AnyTimes()
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient)
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{})
			got, err := catClient.ListDevfileStacks(ctx, tt.registryName, tt.devfileName, tt.filter, false)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("RegistryClient.ListDevfileStacks() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetRegistryDevfiles(t *testing.T) {
	const registryName = "some registry"
	const (
		v1IndexResponse = `
[
	{
		"name": "nodejs",
		"displayName": "NodeJS Angular Web Application",
		"description": "Stack for developing NodeJS Angular Web Application",
		"version": "1.2.3",
		"tags": [
			"NodeJS",
			"Angular",
			"Alpine"
		],
		"language": "nodejs",
		"icon": "/images/angular.svg",
		"globalMemoryLimit": "2686Mi",
		"links": {
			"self": "/devfiles/angular/devfile.yaml"
		}
	}
]
`
		v2IndexResponse = `
[
	{
    "name": "go",
    "displayName": "Go Runtime",
    "description": "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
    "type": "stack",
    "tags": [
      "Go"
    ],
    "icon": "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
    "projectType": "Go",
    "language": "Go",
    "provider": "Red Hat",
    "versions": [
      {
        "version": "2.0.0",
        "schemaVersion": "2.2.0",
        "description": "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
        "tags": [
          "Go"
        ],
        "icon": "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
        "links": {
          "self": "devfile-catalog/go:2.0.0"
        },
        "resources": [
          "devfile.yaml"
        ],
        "starterProjects": [
          "go-starter"
        ]
      },
      {
        "version": "1.0.2",
        "schemaVersion": "2.1.0",
        "default": true,
        "description": "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
        "tags": [
          "Go"
        ],
        "icon": "https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg",
        "links": {
          "self": "devfile-catalog/go:1.0.2"
        },
        "resources": [
          "devfile.yaml"
        ],
        "starterProjects": [
          "go-starter"
        ]
      }
    ]
  }
]
`
	)

	type test struct {
		name                   string
		registryServerProvider func(t *testing.T) (*httptest.Server, string)
		wantErr                bool
		wantProvider           func(registryUrl string) []api.DevfileStack
	}
	tests := []test{
		{
			name: "GitHub-based registry: github.com",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				return nil, "https://github.com/redhat-developer/odo"
			},
			wantErr: true,
		},
		{
			name: "GitHub-based registry: raw.githubusercontent.com",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				return nil, "https://raw.githubusercontent.com/redhat-developer/odo"
			},
			wantErr: true,
		},
		{
			name: "GitHub-based registry: *.github.com",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				return nil, "https://redhat-developer.github.com/odo"
			},
			wantErr: true,
		},
		{
			name: "GitHub-based registry: *.raw.githubusercontent.com",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				return nil, "https://redhat-developer.raw.githubusercontent.com/odo"
			},
			wantErr: true,
		},
		{
			name: "Devfile registry server: client error (4xx)",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusNotFound)
				}))
				return server, server.URL
			},
			wantErr: true,
		},
		{
			name: "Devfile registry server: server error (5xx)",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusInternalServerError)
				}))
				return server, server.URL
			},
			wantErr: true,
		},
		{
			name: "Devfile registry: only /index",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					if req.URL.Path != "/index" {
						rw.WriteHeader(http.StatusNotFound)
						return
					}
					_, err := rw.Write([]byte(v1IndexResponse))
					if err != nil {
						t.Error(err)
					}
				}))
				return server, server.URL
			},
			wantProvider: func(registryUrl string) []api.DevfileStack {
				return []api.DevfileStack{
					{
						Name:           "nodejs",
						DisplayName:    "NodeJS Angular Web Application",
						Description:    "Stack for developing NodeJS Angular Web Application",
						Registry:       api.Registry{Name: registryName, URL: registryUrl},
						Language:       "nodejs",
						Tags:           []string{"NodeJS", "Angular", "Alpine"},
						DefaultVersion: "1.2.3",
					},
				}
			},
		},
		{
			name: "Devfile registry: only /v2index",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					if req.URL.Path != "/v2index" {
						rw.WriteHeader(http.StatusNotFound)
						return
					}
					_, err := rw.Write([]byte(v2IndexResponse))
					if err != nil {
						t.Error(err)
					}
				}))
				return server, server.URL
			},
			wantProvider: func(registryUrl string) []api.DevfileStack {
				return []api.DevfileStack{
					{
						Name:                   "go",
						DisplayName:            "Go Runtime",
						Description:            "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
						Registry:               api.Registry{Name: registryName, URL: registryUrl},
						Language:               "Go",
						ProjectType:            "Go",
						Tags:                   []string{"Go"},
						DefaultVersion:         "1.0.2",
						DefaultStarterProjects: []string{"go-starter"},
						Versions: []api.DevfileStackVersion{
							{Version: "1.0.2", IsDefault: true, SchemaVersion: "2.1.0", StarterProjects: []string{"go-starter"}},
							{Version: "2.0.0", IsDefault: false, SchemaVersion: "2.2.0", StarterProjects: []string{"go-starter"}},
						},
					},
				}
			},
		},
		{
			name: "Devfile registry: both /index and /v2index => v2index has precedence",
			registryServerProvider: func(t *testing.T) (*httptest.Server, string) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					var resp string
					switch req.URL.Path {
					case "/index":
						resp = v1IndexResponse
					case "/v2index":
						resp = v2IndexResponse
					}
					if resp == "" {
						rw.WriteHeader(http.StatusNotFound)
						return
					}
					_, err := rw.Write([]byte(resp))
					if err != nil {
						t.Error(err)
					}
				}))
				return server, server.URL
			},
			wantProvider: func(registryUrl string) []api.DevfileStack {
				return []api.DevfileStack{
					{
						Name:                   "go",
						DisplayName:            "Go Runtime",
						Description:            "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
						Registry:               api.Registry{Name: registryName, URL: registryUrl},
						Language:               "Go",
						ProjectType:            "Go",
						Tags:                   []string{"Go"},
						DefaultVersion:         "1.0.2",
						DefaultStarterProjects: []string{"go-starter"},
						Versions: []api.DevfileStackVersion{
							{Version: "1.0.2", IsDefault: true, SchemaVersion: "2.1.0", StarterProjects: []string{"go-starter"}},
							{Version: "2.0.0", IsDefault: false, SchemaVersion: "2.2.0", StarterProjects: []string{"go-starter"}},
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			ctx := envcontext.WithEnvConfig(context.Background(), config.Configuration{})
			server, url := tt.registryServerProvider(t)
			if server != nil {
				defer server.Close()
			}

			got, err := getRegistryStacks(ctx, prefClient, api.Registry{Name: registryName, URL: url})

			if tt.wantErr != (err != nil) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantProvider != nil {
				want := tt.wantProvider(url)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("getRegistryStacks() mismatch (-want +got):\n%s", diff)
					t.Logf("Error message is: %v", err)
				}
			}
		})
	}
}
