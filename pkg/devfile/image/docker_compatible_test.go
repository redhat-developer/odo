package image

import (
	"path/filepath"
	"reflect"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShellCommand(tt.cmdName, tt.image, tt.devfilePath)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s:\n  Expected %v,\n       got %v", tt.name, tt.want, got)
			}
		})
	}
}
