package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDevfileState_AddContainer(t *testing.T) {
	type args struct {
		name       string
		image      string
		command    []string
		args       []string
		memRequest string
		memLimit   string
		cpuRequest string
		cpuLimit   string
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add a container",
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
					},
				},
				Images:    []Image{},
				Resources: []Resource{},
				Events:    Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddContainer(tt.args.name, tt.args.image, tt.args.command, tt.args.args, tt.args.memRequest, tt.args.memLimit, tt.args.cpuRequest, tt.args.cpuLimit)
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
					"1Gi",
					"2Gi",
					"100m",
					"200m",
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
					"1Gi",
					"2Gi",
					"100m",
					"200m",
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
						URI:          "an-uri",
					},
				},
				Resources: []Resource{},
				Events:    Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddImage(tt.args.name, tt.args.imageName, tt.args.args, tt.args.buildContext, tt.args.rootRequired, tt.args.uri)
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
		name   string
		inline string
		uri    string
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
						Name: "a-name",
						URI:  "an-uri",
					},
				},
				Events: Events{},
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
					},
				},
				Events: Events{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state()
			got, err := o.AddResource(tt.args.name, tt.args.inline, tt.args.uri)
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
