package image

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestGetShellCommand(t *testing.T) {
	type test struct {
		name            string
		cmdName         string
		globalExtraArgs []string
		buildExtraArgs  []string
		image           *devfile.ImageComponent
		devfilePath     string
		want            []string
	}
	devfilePath := filepath.Join("home", "user", "project1")
	tests := []test{
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

	allTests := make([]test, len(tests))
	copy(allTests, tests)
	globalExtraArgs := []string{
		"--global-flag1=value1",
		"--global-flag2=value2",
	}
	buildExtraArgs := []string{
		"--flag1=value1",
		"--flag2=value2",
	}
	for _, tt := range tests {
		var want []string
		if len(tt.want) != 0 {
			want = append(want, tt.cmdName)
			want = append(want, globalExtraArgs...)
			if len(tt.want) >= 2 {
				want = append(want, tt.want[1])
			}
			want = append(want, buildExtraArgs...)
			if len(tt.want) > 3 {
				want = append(want, tt.want[2:]...)
			}
		}
		allTests = append(allTests, test{
			name:            tt.name + " - with extra args",
			cmdName:         tt.cmdName,
			globalExtraArgs: globalExtraArgs,
			buildExtraArgs:  buildExtraArgs,
			image:           tt.image,
			devfilePath:     devfilePath,
			want:            want,
		})
	}

	for _, tt := range allTests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShellCommand(tt.cmdName, tt.globalExtraArgs, tt.buildExtraArgs, tt.image, tt.devfilePath, tt.image.Dockerfile.Uri)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getShellCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_resolveAndDownloadDockerfile(t *testing.T) {
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
			got, gotIsTemp, err := resolveAndDownloadDockerfile(fakeFs, uri)
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
