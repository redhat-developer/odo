package image

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestGetShellCommand(t *testing.T) {
	devfilePath := filepath.Join("home", "user", "project1")
	tests := []struct {
		name        string
		cmdName     string
		image       *devfile.ImageComponent
		devfilePath string
		want        []string
	}{
		{
			name:    "test 1",
			cmdName: "cli",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "registry.io/myimagename:tag",
					ImageUnion: devfile.ImageUnion{
						Dockerfile: &devfile.DockerfileImage{
							DockerfileSrc: devfile.DockerfileSrc{
								Uri: "./Dockerfile",
							},
							Dockerfile: devfile.Dockerfile{
								BuildContext: "${PROJECTS_ROOT}",
							},
						},
					},
				},
			},
			devfilePath: devfilePath,
			want: []string{
				"cli", "build", "-t", "registry.io/myimagename:tag", "-f", filepath.Join(devfilePath, "Dockerfile"), "${PROJECTS_ROOT}",
			},
		},
		{
			name:    "test 2",
			cmdName: "cli",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "registry.io/myimagename:tag",
					ImageUnion: devfile.ImageUnion{
						Dockerfile: &devfile.DockerfileImage{
							DockerfileSrc: devfile.DockerfileSrc{
								Uri: "Dockerfile",
							},
							Dockerfile: devfile.Dockerfile{
								BuildContext: "${PROJECTS_ROOT}",
							},
						},
					},
				},
			},
			devfilePath: devfilePath,
			want: []string{
				"cli", "build", "-t", "registry.io/myimagename:tag", "-f", filepath.Join(devfilePath, "Dockerfile"), "${PROJECTS_ROOT}",
			},
		},
		{
			name:    "test with args",
			cmdName: "cli",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "registry.io/myimagename:tag",
					ImageUnion: devfile.ImageUnion{
						Dockerfile: &devfile.DockerfileImage{
							DockerfileSrc: devfile.DockerfileSrc{
								Uri: "Dockerfile",
							},
							Dockerfile: devfile.Dockerfile{
								BuildContext: "${PROJECTS_ROOT}",
								Args:         []string{"--flag", "value"},
							},
						},
					},
				},
			},
			devfilePath: devfilePath,
			want: []string{
				"cli", "build", "-t", "registry.io/myimagename:tag", "-f", filepath.Join(devfilePath, "Dockerfile"), "${PROJECTS_ROOT}", "--flag", "value",
			},
		},
		{
			name:    "test with no build context in Devfile",
			cmdName: "cli",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "registry.io/myimagename:tag",
					ImageUnion: devfile.ImageUnion{
						Dockerfile: &devfile.DockerfileImage{
							DockerfileSrc: devfile.DockerfileSrc{
								Uri: "Dockerfile.rhel",
							},
						},
					},
				},
			},
			devfilePath: devfilePath,
			want: []string{
				"cli", "build", "-t", "registry.io/myimagename:tag", "-f", filepath.Join(devfilePath, "Dockerfile.rhel"), devfilePath,
			},
		},
		{
			name:    "using an absolute Dockerfile path",
			cmdName: "cli",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "registry.io/myimagename:tag",
					ImageUnion: devfile.ImageUnion{
						Dockerfile: &devfile.DockerfileImage{
							DockerfileSrc: devfile.DockerfileSrc{
								Uri: filepath.Join("/", "path", "to", "Dockerfile.rhel"),
							},
						},
					},
				},
			},
			devfilePath: devfilePath,
			want: []string{
				"cli", "build", "-t", "registry.io/myimagename:tag", "-f", filepath.Join("/", "path", "to", "Dockerfile.rhel"), devfilePath,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShellCommand(tt.cmdName, tt.image, tt.devfilePath, tt.image.Dockerfile.Uri)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s:\n  Expected %v,\n       got %v", tt.name, tt.want, got)
			}
		})
	}
}

func Test_resolveDockerfile(t *testing.T) {
	fakeFs := filesystem.NewFakeFs()

	for _, tt := range []struct {
		name       string
		uriFunc    func() (*httptest.Server, string)
		wantErr    bool
		wantIsTemp bool
		want       string
	}{
		{
			name:    "local file",
			uriFunc: func() (*httptest.Server, string) { return nil, "Dockerfile" },
			want:    "Dockerfile",
		},
		{
			name:    "remote file (non-HTTP)",
			uriFunc: func() (*httptest.Server, string) { return nil, "ftp://example.com/Dockerfile" },
			want:    "ftp://example.com/Dockerfile",
		},
		{
			name: "remote file with error (HTTP)",
			uriFunc: func() (*httptest.Server, string) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
				return s, s.URL + "/404"
			},
			wantErr:    true,
			wantIsTemp: true,
		},
		{
			name: "remote file (HTTP)",
			uriFunc: func() (*httptest.Server, string) {
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, "FROM alpine:3.6")
					fmt.Fprintln(w, "RUN echo Hello World")
					fmt.Fprintln(w, "ENTRYPOINT [\"/bin/tail\"]")
					fmt.Fprintln(w, "CMD [\"-f\", \"/dev/null\"]")
				}))
				return s, s.URL
			},
			wantIsTemp: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server, uri := tt.uriFunc()
			if server != nil {
				defer server.Close()
			}
			got, gotIsTemp, err := resolveDockerfile(fakeFs, uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s:\n  Expected error %v,\n       got %v", tt.name, tt.wantErr, err)
			}
			if gotIsTemp != tt.wantIsTemp {
				t.Errorf("%s:\n  For 'isTemp', expected %v,\n       got %v", tt.name, tt.wantIsTemp, gotIsTemp)
			}
			if gotIsTemp {
				defer func(fs filesystem.Filesystem, name string) {
					_ = fs.Remove(name)
				}(fakeFs, got)
				// temp file is created, so we can't compare the path, but we can check the path is not blank
				if strings.TrimSpace(got) == "" {
					t.Errorf("%s:\n  Expected non-blank path,\n       got blank path: %s", tt.name, got)
				}
			} else if got != tt.want {
				t.Errorf("%s:\n  Expected %v,\n       got %v", tt.name, tt.want, got)
			}
		})
	}
}
