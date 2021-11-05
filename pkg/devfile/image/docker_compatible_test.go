package image

import (
	"path/filepath"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func TestGetShellCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		image       *devfile.ImageComponent
		devfilePath string
		want        string
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
								BuildContext: "${PROJECT_ROOT}",
							},
						},
					},
				},
			},
			devfilePath: filepath.Join("home", "user", "project1"),
			want:        `cli build -t "registry.io/myimagename:tag" -f "` + filepath.Join("home", "user", "project1", "Dockerfile") + `" ${PROJECT_ROOT}`,
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
								BuildContext: "${PROJECT_ROOT}",
							},
						},
					},
				},
			},
			devfilePath: filepath.Join("home", "user", "project1"),
			want:        `cli build -t "registry.io/myimagename:tag" -f "` + filepath.Join("home", "user", "project1", "Dockerfile") + `" ${PROJECT_ROOT}`,
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
								BuildContext: "${PROJECT_ROOT}",
								Args:         []string{"--flag", "value"},
							},
						},
					},
				},
			},
			devfilePath: filepath.Join("home", "user", "project1"),
			want:        `cli build -t "registry.io/myimagename:tag" -f "` + filepath.Join("home", "user", "project1", "Dockerfile") + `" ${PROJECT_ROOT} --flag value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShellCommand(tt.cmdName, tt.image, tt.devfilePath)
			if got != tt.want {
				t.Errorf("%s:\n  Expected %q,\n       got %q", tt.name, tt.want, got)
			}
		})
	}
}
