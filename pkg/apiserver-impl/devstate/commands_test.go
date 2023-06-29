package devstate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDevfileState_AddExecCommand(t *testing.T) {
	type args struct {
		name             string
		component        string
		commandLine      string
		workingDir       string
		hotReloadCapable bool
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add an exec command",
			state: func() DevfileState {
				state := NewDevfileState()
				_, err := state.AddContainer(
					"a-container",
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
				name:             "an-exec-command",
				component:        "a-container",
				commandLine:      "run command",
				workingDir:       "/path/to/work",
				hotReloadCapable: true,
			},
			want: DevfileContent{
				Content: `commands:
- exec:
    commandLine: run command
    component: a-container
    hotReloadCapable: true
    workingDir: /path/to/work
  id: an-exec-command
components:
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
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name: "an-exec-command",
						Type: "exec",
						Exec: &ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
				},
				Containers: []Container{
					{
						Name:          "a-container",
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
			got, err := o.AddExecCommand(tt.args.name, tt.args.component, tt.args.commandLine, tt.args.workingDir, tt.args.hotReloadCapable)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddExecCommand() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddExecCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_AddApplyCommand(t *testing.T) {
	type args struct {
		name      string
		component string
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add an Apply command",
			state: func() DevfileState {
				state := NewDevfileState()
				_, err := state.AddImage(
					"an-image",
					"an-image-name",
					nil, "/context", false, "",
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name:      "an-apply-command",
				component: "an-image",
			},
			want: DevfileContent{
				Content: `commands:
- apply:
    component: an-image
  id: an-apply-command
components:
- image:
    dockerfile:
      buildContext: /context
      rootRequired: false
    imageName: an-image-name
  name: an-image
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name: "an-apply-command",
						Type: "image",
						Image: &ImageCommand{
							Component: "an-image",
						},
					},
				},
				Containers: []Container{},
				Images: []Image{
					{
						Name:         "an-image",
						ImageName:    "an-image-name",
						BuildContext: "/context",
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
			got, err := o.AddApplyCommand(tt.args.name, tt.args.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddApplyCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddApplyCommand() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddApplyCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_AddCompositeCommand(t *testing.T) {
	type args struct {
		name     string
		parallel bool
		commands []string
	}
	tests := []struct {
		name    string
		state   func() DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "Add an Apply command",
			state: func() DevfileState {
				state := NewDevfileState()
				_, err := state.AddContainer(
					"a-container",
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
				_, err = state.AddExecCommand(
					"an-exec-command",
					"a-container",
					"run command",
					"/path/to/work",
					true,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name:     "a-composite-command",
				parallel: true,
				commands: []string{"an-exec-command"},
			},
			want: DevfileContent{
				Content: `commands:
- exec:
    commandLine: run command
    component: a-container
    hotReloadCapable: true
    workingDir: /path/to/work
  id: an-exec-command
- composite:
    commands:
    - an-exec-command
    parallel: true
  id: a-composite-command
components:
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
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name: "an-exec-command",
						Type: "exec",
						Exec: &ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
					{
						Name:      "a-composite-command",
						Type:      "composite",
						Composite: &CompositeCommand{Commands: []string{"an-exec-command"}, Parallel: true},
					},
				},
				Containers: []Container{
					{
						Name:          "a-container",
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
			got, err := o.AddCompositeCommand(tt.args.name, tt.args.parallel, tt.args.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.AddApplyCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.AddApplyCommand() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.AddApplyCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_DeleteCommand(t *testing.T) {
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
			name: "Delete an existing command",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddContainer(
					"a-container",
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
				_, err = state.AddExecCommand(
					"an-exec-command",
					"a-container",
					"run command",
					"/path/to/work",
					true,
				)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				name: "an-exec-command",
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
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{},
				Containers: []Container{
					{
						Name:          "a-container",
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
		{
			name: "Delete a non existing command",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
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
			got, err := o.DeleteCommand(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.DeleteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want.Content, got.Content); diff != "" {
				t.Errorf("DevfileState.DeleteCommand() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.DeleteCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
