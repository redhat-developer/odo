package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
)

func TestDevfileState_AddContainer(t *testing.T) {
	type args struct {
		name             string
		image            string
		command          []string
		args             []string
		envs             []Env
		memRequest       string
		memLimit         string
		cpuRequest       string
		cpuLimit         string
		volumeMounts     []openapi.VolumeMount
		configureSources bool
		mountSources     bool
		sourceMapping    string
		annotation       Annotation
		endpoints        []Endpoint
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add a container, with sources configured",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:       "a-name",
				image:      "an-image",
				command:    []string{"run", "command"},
				args:       []string{"arg1", "arg2"},
				memRequest: "1Gi",
				memLimit:   "2Gi",
				cpuRequest: "100m",
				cpuLimit:   "200m",
				volumeMounts: []openapi.VolumeMount{
					{
						Name: "vol1",
						Path: "/mnt/volume1",
					},
				},
				configureSources: true,
				mountSources:     false,
			},
			want: DevfileContent{
				Content: `components:
- container:
    args:
    - arg1
    - arg2
    command:
    - run
    - command
    cpuLimit: 200m
    cpuRequest: 100m
    image: an-image
    memoryLimit: 2Gi
    memoryRequest: 1Gi
    mountSources: false
    volumeMounts:
    - name: vol1
      path: /mnt/volume1
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{},
				Containers: []Container{
					{
						Name:          "a-name",
						Image:         "an-image",
						Command:       []string{"run", "command"},
						Args:          []string{"arg1", "arg2"},
						MemoryRequest: "1Gi",
						MemoryLimit:   "2Gi",
						CpuRequest:    "100m",
						CpuLimit:      "200m",
						VolumeMounts: []openapi.VolumeMount{
							{
								Name: "vol1",
								Path: "/mnt/volume1",
							},
						},
						Endpoints:        []openapi.Endpoint{},
						Env:              []openapi.Env{},
						ConfigureSources: true,
						MountSources:     false,
					},
				},
				Images:    []Image{},
				Resources: []Resource{},
				Volumes:   []Volume{},
				Events:    Events{},
			},
		},
		{
			name: "Add a container, without sources configured",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:       "a-name",
				image:      "an-image",
				command:    []string{"run", "command"},
				args:       []string{"arg1", "arg2"},
				memRequest: "1Gi",
				memLimit:   "2Gi",
				cpuRequest: "100m",
				cpuLimit:   "200m",
				volumeMounts: []openapi.VolumeMount{
					{
						Name: "vol1",
						Path: "/mnt/volume1",
					},
				},
				configureSources: false,
			},
			want: DevfileContent{
				Content: `components:
- container:
    args:
    - arg1
    - arg2
    command:
    - run
    - command
    cpuLimit: 200m
    cpuRequest: 100m
    image: an-image
    memoryLimit: 2Gi
    memoryRequest: 1Gi
    volumeMounts:
    - name: vol1
      path: /mnt/volume1
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{},
				Containers: []Container{
					{
						Name:          "a-name",
						Image:         "an-image",
						Command:       []string{"run", "command"},
						Args:          []string{"arg1", "arg2"},
						MemoryRequest: "1Gi",
						MemoryLimit:   "2Gi",
						CpuRequest:    "100m",
						CpuLimit:      "200m",
						VolumeMounts: []openapi.VolumeMount{
							{
								Name: "vol1",
								Path: "/mnt/volume1",
							},
						},
						Endpoints:        []openapi.Endpoint{},
						Env:              []openapi.Env{},
						ConfigureSources: false,
						MountSources:     true,
					},
				},
				Images:    []Image{},
				Resources: []Resource{},
				Volumes:   []Volume{},
				Events:    Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddContainer(tt.args.name, tt.args.image, tt.args.command, tt.args.args, tt.args.envs, tt.args.memRequest, tt.args.memLimit, tt.args.cpuRequest, tt.args.cpuLimit, tt.args.volumeMounts, tt.args.configureSources, tt.args.mountSources, tt.args.sourceMapping, tt.args.annotation, tt.args.endpoints)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddContainer() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddContainer() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_DeleteContainer(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Delete an existing container",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddContainer(
					"a-name",
					"an-image",
					[]string{"run", "command"},
					[]string{"arg1", "arg2"},
					nil,
					"1Gi",
					"2Gi",
					"100m",
					"200m",
					nil,
					true,
					false,
					"",
					Annotation{},
					nil,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "a-name",
			},
			want: DevfileContent{
				Content: `metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events:     Events{},
			},
		},
		{
			name: "Delete a non existing container",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddContainer(
					"a-name",
					"an-image",
					[]string{"run", "command"},
					[]string{"arg1", "arg2"},
					nil,
					"1Gi",
					"2Gi",
					"100m",
					"200m",
					nil,
					true,
					false,
					"",
					Annotation{},
					nil,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "another-name",
			},
			want:    DevfileContent{},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.DeleteContainer(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.DeleteContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.DeleteContainer() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.DeleteContainer() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_AddImage(t *testing.T) {
	type args struct {
		name         string
		imageName    string
		args         []string
		buildContext string
		rootRequired bool
		uri          string
		autoBuild    bool
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add an image",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:         "a-name",
				imageName:    "an-image-name",
				args:         []string{"start", "command"},
				buildContext: "path/to/context",
				rootRequired: true,
				uri:          "an-uri",
			},
			want: DevfileContent{
				Content: `components:
- image:
    autoBuild: false
    dockerfile:
      args:
      - start
      - command
      buildContext: path/to/context
      rootRequired: true
      uri: an-uri
    imageName: an-image-name
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images: []Image{
					{
						Name:         "a-name",
						ImageName:    "an-image-name",
						Args:         []string{"start", "command"},
						BuildContext: "path/to/context",
						RootRequired: true,
						Uri:          "an-uri",
						Orphan:       true,
					},
				},
				Resources: []Resource{},
				Volumes:   []Volume{},
				Events:    Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddImage(tt.args.name, tt.args.imageName, tt.args.args, tt.args.buildContext, tt.args.rootRequired, tt.args.uri, tt.args.autoBuild)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddImage() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddImage() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_DeleteImage(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Delete an existing image",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddImage(
					"a-name",
					"an-image-name",
					[]string{"start", "command"},
					"path/to/context",
					true,
					"an-uri",
					false,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "a-name",
			},
			want: DevfileContent{
				Content: `metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events:     Events{},
			},
		},
		{
			name: "Delete a non existing image",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddImage(
					"a-name",
					"an-image-name",
					[]string{"start", "command"},
					"path/to/context",
					true,
					"an-uri",
					false,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "another-name",
			},
			want:    DevfileContent{},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.DeleteImage(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.DeleteImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.DeleteImage() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.DeleteImage() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_AddResource(t *testing.T) {
	type args struct {
		name            string
		inline          string
		uri             string
		deployByDefault bool
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add a resource with uri",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name: "a-name",
				uri:  "an-uri",
			},
			want: DevfileContent{
				Content: `components:
- kubernetes:
    deployByDefault: false
    uri: an-uri
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources: []Resource{
					{
						Name:   "a-name",
						Uri:    "an-uri",
						Orphan: true,
					},
				},
				Volumes: []Volume{},
				Events:  Events{},
			},
		},
		{
			name: "Add an inline resource",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:   "a-name",
				inline: "inline resource...",
			},
			want: DevfileContent{
				Content: `components:
- kubernetes:
    deployByDefault: false
    inlined: inline resource...
  name: a-name
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources: []Resource{
					{
						Name:    "a-name",
						Inlined: "inline resource...",
						Orphan:  true,
					},
				},
				Volumes: []Volume{},
				Events:  Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddResource(tt.args.name, tt.args.inline, tt.args.uri, tt.args.deployByDefault)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddResource() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddResource() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_Deleteresource(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Delete an existing resource",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddResource(
					"a-name",
					"",
					"an-uri",
					false,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "a-name",
			},
			want: DevfileContent{
				Content: `metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events:     Events{},
			},
		},
		{
			name: "Delete a non existing resource",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddResource(
					"a-name",
					"",
					"an-uri",
					false,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "another-name",
			},
			want:    DevfileContent{},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.DeleteResource(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.DeleteResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.DeleteResource() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.DeleteResource() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_AddVolume(t *testing.T) {
	type args struct {
		name      string
		size      string
		ephemeral bool
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add a volume",
			state: func() DevfileState {
				return NewDevfileState()
			},
			args: args{
				name:      "a-name",
				size:      "1Gi",
				ephemeral: true,
			},
			want: DevfileContent{
				Content: `components:
- name: a-name
  volume:
    ephemeral: true
    size: 1Gi
metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes: []Volume{
					{
						Name:      "a-name",
						Size:      "1Gi",
						Ephemeral: true,
					},
				},
				Events: Events{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddVolume(tt.args.name, tt.args.ephemeral, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddVolume() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddVolume() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_DeleteVolume(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Delete an existing volume",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddVolume(
					"a-name",
					true,
					"1Gi",
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "a-name",
			},
			want: DevfileContent{
				Content: `metadata: {}
schemaVersion: 2.2.0
`,
				Commands:   []Command{},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
				Events:     Events{},
			},
		},
		{
			name: "Delete a non existing resource",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddVolume(
					"a-name",
					true,
					"1Gi",
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "another-name",
			},
			want:    DevfileContent{},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.DeleteVolume(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.DeleteVolume() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.DeleteVolume() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
