package lclient

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// GenerateContainerConfig creates a containerConfig resource that can be used to create a local Docker container
func TestGenerateContainerConfig(t *testing.T) {
	fakeClient := FakeNew()
	tests := []struct {
		name       string
		image      string
		entrypoint []string
		cmd        []string
		envVars    []string
		portset    nat.PortSet
		labels     map[string]string
		want       container.Config
	}{
		{
			name:       "Case 1: Simple config, no env vars or labels",
			image:      "docker.io/fake-image:latest",
			entrypoint: []string{"bash"},
			cmd:        []string{"tail", "-f", "/dev/null"},
			envVars:    []string{},
			labels:     nil,
			want: container.Config{
				Image:      "docker.io/fake-image:latest",
				Entrypoint: []string{"bash"},
				Cmd:        []string{"tail", "-f", "/dev/null"},
				Env:        []string{},
				Labels:     nil,
				User:       "root",
			},
		},
		{
			name:       "Case 2: Simple config, env vars and labels set",
			image:      "docker.io/fake-image:latest",
			entrypoint: []string{"bash"},
			cmd:        []string{"tail", "-f", "/dev/null"},
			envVars:    []string{"test=hello", "sample=value"},
			labels: map[string]string{
				"component": "some-component",
				"alias":     "maven",
			},
			want: container.Config{
				Image:      "docker.io/fake-image:latest",
				Entrypoint: []string{"bash"},
				Cmd:        []string{"tail", "-f", "/dev/null"},
				Env:        []string{"test=hello", "sample=value"},
				Labels: map[string]string{
					"component": "some-component",
					"alias":     "maven",
				},
				User: "root",
			},
		},
		{
			name:       "Case 3: Simple config, adding portset",
			image:      "docker.io/fake-image:latest",
			entrypoint: []string{"bash"},
			cmd:        []string{"tail", "-f", "/dev/null"},
			portset: nat.PortSet{
				"8080/tcp": struct{}{},
				"9080/tcp": struct{}{},
			},
			want: container.Config{
				Image:      "docker.io/fake-image:latest",
				Entrypoint: []string{"bash"},
				Cmd:        []string{"tail", "-f", "/dev/null"},
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
					"9080/tcp": struct{}{},
				},
				User: "root",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := fakeClient.GenerateContainerConfig(tt.image, tt.entrypoint, tt.cmd, tt.envVars, tt.labels, tt.portset)
			if !reflect.DeepEqual(tt.want, config) {
				t.Errorf("expected %v, actual %v", tt.want, config)
			}
		})
	}
}

func TestGenerateHostConfig(t *testing.T) {
	fakeClient := FakeNew()
	tests := []struct {
		name string
		want container.HostConfig
	}{
		{
			name: "Case 1: Unprivileged and not publishing ports",
			want: container.HostConfig{
				Privileged:      false,
				PublishAllPorts: false,
				PortBindings:    nat.PortMap{},
			},
		},
		{
			name: "Case 2: Privileged and not publishing ports",
			want: container.HostConfig{
				Privileged:      true,
				PublishAllPorts: false,
				PortBindings:    nat.PortMap{},
			},
		},
		{
			name: "Case 3: Unprivileged and publishing ports",
			want: container.HostConfig{
				Privileged:      false,
				PublishAllPorts: true,
				PortBindings:    nat.PortMap{},
			},
		},
		{
			name: "Case 4: Privileged and publishing ports",
			want: container.HostConfig{
				Privileged:      true,
				PublishAllPorts: true,
				PortBindings:    nat.PortMap{},
			},
		},
		{
			name: "Case 5: With non-empty PortBindings",
			want: container.HostConfig{
				Privileged:      true,
				PublishAllPorts: true,
				PortBindings: nat.PortMap{
					"tcp/9090": []nat.PortBinding{
						nat.PortBinding{
							HostIP:   "127.0.0.1",
							HostPort: "65432",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := fakeClient.GenerateHostConfig(tt.want.Privileged, tt.want.PublishAllPorts, tt.want.PortBindings)
			if !reflect.DeepEqual(tt.want, config) {
				t.Errorf("expected %v, actual %v", tt.want, config)
			}
		})
	}
}
