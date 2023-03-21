package registry

import (
	"context"
	"errors"
	"fmt"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestGetDevfileRegistries(t *testing.T) {
	tempConfigFile, err := os.CreateTemp("", "odoconfig")
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
		kclient      func(ctrl *gomock.Controller) kclient.ClientInterface
		want         []api.Registry
		wantErr      bool
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
		{
			name: "with devfileRegistrylists in cluster",
			kclient: func(ctrl *gomock.Controller) kclient.ClientInterface {
				result := kclient.NewMockClientInterface(ctrl)
				list := []api.Registry{
					{
						Name:   "secure-name",
						URL:    "secure-url",
						Secure: true,
					},
					{
						Name:   "unsecure-name",
						URL:    "unsecure-url",
						Secure: false,
					},
				}
				result.EXPECT().GetRegistryList().Return(list, nil)
				return result
			},
			want: []api.Registry{
				{
					Name:   "secure-name",
					URL:    "secure-url",
					Secure: true,
				},
				{
					Name:   "unsecure-name",
					URL:    "unsecure-url",
					Secure: false,
				},
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
			name: "error getting devfileRegistrylists from cluster",
			kclient: func(ctrl *gomock.Controller) kclient.ClientInterface {
				result := kclient.NewMockClientInterface(ctrl)
				result.EXPECT().GetRegistryList().Return(nil, errors.New("an error"))
				return result
			},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = envcontext.WithEnvConfig(ctx, config.Configuration{
				Globalodoconfig: &tempConfigFileName,
			})

			ctrl := gomock.NewController(t)

			prefClient, _ := preference.NewClient(ctx)
			var kc kclient.ClientInterface
			if tt.kclient != nil {
				kc = tt.kclient(ctrl)
			}
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient, kc)
			got, err := catClient.GetDevfileRegistries(tt.registryName)

			if tt.wantErr != (err != nil) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
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
			prefClient.EXPECT().RegistryList().Return([]api.Registry{
				{
					Name: "TestRegistry",
					URL:  server.URL,
				},
			}).AnyTimes()
			// TODO(rm3l) Test with both nil and non-nil kubeclient
			catClient := NewRegistryClient(filesystem.NewFakeFs(), prefClient, nil)
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
			ctx := envcontext.WithEnvConfig(context.Background(), config.Configuration{})
			server, url := tt.registryServerProvider(t)
			if server != nil {
				defer server.Close()
			}

			got, err := getRegistryStacks(ctx, api.Registry{Name: registryName, URL: url})

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

func TestRegistryClient_DownloadStarterProject(t *testing.T) {
	setupFS := func(contextDir string, fs filesystem.Filesystem) {
		directoriesToBeCreated := []string{"kubernetes", "docker"}
		for _, dirToBeCreated := range directoriesToBeCreated {
			dirName := filepath.Join(contextDir, dirToBeCreated)
			err := fs.MkdirAll(dirName, os.FileMode(0750))
			if err != nil {
				t.Errorf("failed to create %s; cause: %s", dirName, err)
			}
		}

		filesToBeCreated := []string{"devfile.yaml", "kubernetes/deploy.yaml", "docker/Dockerfile"}
		for _, fileToBeCreated := range filesToBeCreated {
			fileName := filepath.Join(contextDir, fileToBeCreated)
			_, err := fs.Create(fileName)
			if err != nil {
				t.Errorf("failed to create %s; cause: %s", fileName, err)
			}
		}
	}

	getZipFilePath := func(name string) string {
		// filename of this file
		_, filename, _, _ := runtime.Caller(0)
		// path to the devfile
		return filepath.Join(filepath.Dir(filename), "..", "..", "tests", "examples", filepath.Join("source", "devfiles", "zips", name))
	}
	type fields struct {
		fsys             filesystem.Filesystem
		preferenceClient preference.Client
		kubeClient       kclient.ClientInterface
	}
	type args struct {
		starterProject *devfilev1.StarterProject
		decryptedToken string
		contextDir     string
		verbose        bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Starter project has a Devfile",
			fields: fields{
				fsys: filesystem.NewFakeFs(),
			},
			args: args{
				starterProject: &devfilev1.StarterProject{
					Name: "starter-project-with-devfile",
					ProjectSource: devfilev1.ProjectSource{
						SourceType: "",
						Zip: &devfilev1.ZipProjectSource{
							CommonProjectSource: devfilev1.CommonProjectSource{},
							Location:            fmt.Sprintf("file://%s", getZipFilePath("starterproject-with-devfile.zip")),
						},
					},
				},
			},
			want:    []string{"devfile.yaml", "docker", filepath.Join("docker", "Dockerfile"), "README.md", "main.go", "go.mod", "someFile.txt"},
			wantErr: false,
		},
		{
			name: "Starter project has conflicting files",
			fields: fields{
				fsys: filesystem.NewFakeFs(),
			},
			args: args{
				starterProject: &devfilev1.StarterProject{
					Name: "starter-project-with-conflicting-files",
					ProjectSource: devfilev1.ProjectSource{
						SourceType: "",
						Zip: &devfilev1.ZipProjectSource{
							CommonProjectSource: devfilev1.CommonProjectSource{},
							Location:            fmt.Sprintf("file://%s", getZipFilePath("starterproject-with-conflicts.zip")),
						},
					},
				},
			},
			want: []string{"devfile.yaml", "docker", filepath.Join("docker", "Dockerfile"), "kubernetes", filepath.Join("kubernetes", "deploy.yaml"),
				CONFLICT_DIR_NAME, filepath.Join(CONFLICT_DIR_NAME, "kubernetes"), filepath.Join(CONFLICT_DIR_NAME, "kubernetes", "deploy.yaml"),
				filepath.Join(CONFLICT_DIR_NAME, "main.go"), filepath.Join(CONFLICT_DIR_NAME, "go.mod"), filepath.Join(CONFLICT_DIR_NAME, "README.md"), filepath.Join(CONFLICT_DIR_NAME, "someFile.txt")},
			wantErr: false,
		},
		{
			name: "Starter project has conflicting files and an empty dir",
			fields: fields{
				fsys: filesystem.NewFakeFs(),
			},
			args: args{
				starterProject: &devfilev1.StarterProject{
					Name: "starter-project-with-conflicting-files-and-empty-dir",
					ProjectSource: devfilev1.ProjectSource{
						SourceType: "",
						Zip: &devfilev1.ZipProjectSource{
							CommonProjectSource: devfilev1.CommonProjectSource{},
							Location:            fmt.Sprintf("file://%s", getZipFilePath("starterproject-with-conflicts-and-empty-dir.zip")),
						},
					},
				},
			},
			want: []string{"devfile.yaml", "docker", filepath.Join("docker", "Dockerfile"), "kubernetes", filepath.Join("kubernetes", "deploy.yaml"),
				filepath.Join(CONFLICT_DIR_NAME, "kubernetes"), CONFLICT_DIR_NAME, filepath.Join(CONFLICT_DIR_NAME, "docker"), filepath.Join(CONFLICT_DIR_NAME, "docker", "Dockerfile"),
				filepath.Join(CONFLICT_DIR_NAME, "main.go"), filepath.Join(CONFLICT_DIR_NAME, "go.mod"), filepath.Join(CONFLICT_DIR_NAME, "README.md")},
			wantErr: false,
		},
		{
			name: "Starter project does not have any conflicting files",
			fields: fields{
				fsys: filesystem.NewFakeFs(),
			},
			args: args{
				starterProject: &devfilev1.StarterProject{
					Name: "starter-project-with-no-conflicting-files",
					ProjectSource: devfilev1.ProjectSource{
						SourceType: "",
						Zip: &devfilev1.ZipProjectSource{
							CommonProjectSource: devfilev1.CommonProjectSource{},
							Location:            fmt.Sprintf("file://%s", getZipFilePath("starterproject-with-no-conflicts.zip")),
						},
					},
				},
			},
			want:    []string{"devfile.yaml", "docker", filepath.Join("docker", "Dockerfile"), "kubernetes", filepath.Join("kubernetes", "deploy.yaml"), "README.md", "main.go", "go.mod"},
			wantErr: false,
		},
		{
			name: "Starter project does not have any conflicting files but has empty dir",
			fields: fields{
				fsys: filesystem.NewFakeFs(),
			},
			args: args{
				starterProject: &devfilev1.StarterProject{
					Name: "starter-project-with-no-conflicting-files-and-empty-dir",
					ProjectSource: devfilev1.ProjectSource{
						SourceType: "",
						Zip: &devfilev1.ZipProjectSource{
							CommonProjectSource: devfilev1.CommonProjectSource{},
							Location:            fmt.Sprintf("file://%s", getZipFilePath("starterproject-with-no-conflicts-and-empty-dir.zip")),
						},
					},
				},
			},
			want:    []string{"devfile.yaml", "docker", filepath.Join("docker", "Dockerfile"), "kubernetes", filepath.Join("kubernetes", "deploy.yaml"), "README.md", "main.go", "go.mod"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextDir, ferr := tt.fields.fsys.TempDir("", "downloadstarterproject")
			if ferr != nil {
				t.Errorf("failed to create temp dir; cause: %s", ferr)
			}
			tt.args.contextDir = contextDir

			setupFS(tt.args.contextDir, tt.fields.fsys)

			o := RegistryClient{
				fsys:             tt.fields.fsys,
				preferenceClient: tt.fields.preferenceClient,
				kubeClient:       tt.fields.kubeClient,
			}
			if err := o.DownloadStarterProject(tt.args.starterProject, tt.args.decryptedToken, tt.args.contextDir, tt.args.verbose); (err != nil) != tt.wantErr {
				t.Errorf("DownloadStarterProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got []string
			err := o.fsys.Walk(contextDir, func(path string, info fs.FileInfo, err error) error {
				path = strings.TrimPrefix(path, contextDir)
				path = strings.TrimPrefix(path, string(os.PathSeparator))
				if path != "" {
					got = append(got, path)
				}
				return nil
			})
			if err != nil {
				t.Errorf("failed to walk %s; cause:%s", contextDir, err.Error())
				return
			}
			sort.Strings(got)
			sort.Strings(tt.want)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("DownloadStarterProject() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
