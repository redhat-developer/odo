package devstate

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/google/go-cmp/cmp"
	. "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
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
					nil,
					"1Gi",
					"2Gi",
					"100m",
					"200m",
					nil,
					true,
					true,
					"",
					openapi.Annotation{},
					nil,
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
    mountSources: true
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name: "an-exec-command",
						Type: "exec",
						Exec: ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
				},
				Containers: []Container{
					{
						Name:             "a-container",
						Image:            "an-image",
						Command:          []string{"run", "command"},
						Args:             []string{"arg1", "arg2"},
						MemoryRequest:    "1Gi",
						MemoryLimit:      "2Gi",
						CpuRequest:       "100m",
						CpuLimit:         "200m",
						VolumeMounts:     []openapi.VolumeMount{},
						Endpoints:        []openapi.Endpoint{},
						Env:              []openapi.Env{},
						ConfigureSources: true,
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
						Image: ImageCommand{
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
				Volumes:   []Volume{},
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
					nil,
					"1Gi",
					"2Gi",
					"100m",
					"200m",
					nil,
					true,
					true,
					"",
					openapi.Annotation{},
					nil,
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
    mountSources: true
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name: "an-exec-command",
						Type: "exec",
						Exec: ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
					{
						Name:      "a-composite-command",
						Type:      "composite",
						Composite: CompositeCommand{Commands: []string{"an-exec-command"}, Parallel: true},
					},
				},
				Containers: []Container{
					{
						Name:             "a-container",
						Image:            "an-image",
						Command:          []string{"run", "command"},
						Args:             []string{"arg1", "arg2"},
						MemoryRequest:    "1Gi",
						MemoryLimit:      "2Gi",
						CpuRequest:       "100m",
						CpuLimit:         "200m",
						VolumeMounts:     []openapi.VolumeMount{},
						Endpoints:        []openapi.Endpoint{},
						Env:              []openapi.Env{},
						ConfigureSources: true,
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
					nil,
					"1Gi",
					"2Gi",
					"100m",
					"200m",
					nil,
					true,
					true,
					"",
					openapi.Annotation{},
					nil,
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
    mountSources: true
  name: a-container
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{},
				Containers: []Container{
					{
						Name:             "a-container",
						Image:            "an-image",
						Command:          []string{"run", "command"},
						Args:             []string{"arg1", "arg2"},
						MemoryRequest:    "1Gi",
						MemoryLimit:      "2Gi",
						CpuRequest:       "100m",
						CpuLimit:         "200m",
						VolumeMounts:     []openapi.VolumeMount{},
						Endpoints:        []openapi.Endpoint{},
						Env:              []openapi.Env{},
						ConfigureSources: true,
						MountSources:     true,
					},
				},
				Images:    []Image{},
				Resources: []Resource{},
				Volumes:   []Volume{},
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

func newCommand(group string, id string) v1alpha2.Command {
	return v1alpha2.Command{
		Id: id,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind: v1alpha2.CommandGroupKind(group),
						},
					},
				},
			},
		},
	}
}

func Test_subMoveCommand(t *testing.T) {
	type args struct {
		commands      []v1alpha2.Command
		previousGroup string
		newGroup      string
		previousIndex int
		newIndex      int
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]v1alpha2.Command
		wantErr bool
	}{
		{
			name: "Move from run to test",
			args: args{
				commands: []v1alpha2.Command{
					newCommand("build", "build1"),
					newCommand("run", "runToTest"),
					newCommand("", "other1"),
				},
				previousGroup: "run",
				previousIndex: 0,
				newGroup:      "test",
				newIndex:      0,
			},
			want: map[string][]v1alpha2.Command{
				"build": {
					newCommand("build", "build1"),
				},
				"run": {},
				"test": {
					newCommand("test", "runToTest"),
				},
				"": {
					newCommand("", "other1"),
				},
			},
		},
		{
			name: "Move from other to build",
			args: args{
				commands: []v1alpha2.Command{
					newCommand("build", "build1"),
					newCommand("run", "run"),
					newCommand("other", "otherToBuild"),
				},
				previousGroup: "other",
				previousIndex: 0,
				newGroup:      "build",
				newIndex:      1,
			},
			want: map[string][]v1alpha2.Command{
				"build": {
					newCommand("build", "build1"),
					newCommand("build", "otherToBuild"),
				},
				"run": {
					newCommand("run", "run"),
				},
				"other": {},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := subMoveCommand(tt.args.commands, tt.args.previousGroup, tt.args.newGroup, tt.args.previousIndex, tt.args.newIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("moveCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("moveCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_MoveCommand(t *testing.T) {
	type args struct {
		previousGroup string
		newGroup      string
		previousIndex int
		newIndex      int
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "not found command",
			state: func(t *testing.T) DevfileState {
				return NewDevfileState()
			},
			args: args{
				previousGroup: "build",
				previousIndex: 0,
				newGroup:      "run",
				newIndex:      0,
			},
			wantErr: true,
		},
		{
			name: "command moved from no group to run group",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddExecCommand(
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
				previousGroup: "",
				previousIndex: 0,
				newGroup:      "run",
				newIndex:      0,
			},
			want: DevfileContent{
				Content: `commands:
- exec:
    commandLine: run command
    component: a-container
    group:
      kind: run
    hotReloadCapable: true
    workingDir: /path/to/work
  id: an-exec-command
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name:    "an-exec-command",
						Group:   "run",
						Default: false,
						Type:    "exec",
						Exec: ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
				},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.MoveCommand(tt.args.previousGroup, tt.args.newGroup, tt.args.previousIndex, tt.args.newIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.MoveCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.MoveCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_SetDefaultCommand(t *testing.T) {
	type args struct {
		commandName string
		group       string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "command set to default in run group",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddExecCommand(
					"an-exec-command",
					"a-container",
					"run command",
					"/path/to/work",
					true,
				)
				if err != nil {
					t.Fatal(err)
				}
				_, err = state.MoveCommand("", "run", 0, 0)
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				commandName: "an-exec-command",
				group:       "run",
			},
			want: DevfileContent{
				Content: `commands:
- exec:
    commandLine: run command
    component: a-container
    group:
      isDefault: true
      kind: run
    hotReloadCapable: true
    workingDir: /path/to/work
  id: an-exec-command
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name:    "an-exec-command",
						Group:   "run",
						Default: true,
						Type:    "exec",
						Exec: ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
				},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.SetDefaultCommand(tt.args.commandName, tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.SetDefaultCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.SetDefaultCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevfileState_UnsetDefaultCommand(t *testing.T) {
	type args struct {
		commandName string
	}
	tests := []struct {
		name    string
		state   func(t *testing.T) DevfileState
		args    args
		want    DevfileContent
		wantErr bool
	}{
		{
			name: "command unset default",
			state: func(t *testing.T) DevfileState {
				state := NewDevfileState()
				_, err := state.AddExecCommand(
					"an-exec-command",
					"a-container",
					"run command",
					"/path/to/work",
					true,
				)
				if err != nil {
					t.Fatal(err)
				}
				_, err = state.MoveCommand("", "run", 0, 0)
				if err != nil {
					t.Fatal(err)
				}
				_, err = state.SetDefaultCommand("an-exec-command", "run")
				if err != nil {
					t.Fatal(err)
				}
				return state
			},
			args: args{
				commandName: "an-exec-command",
			},
			want: DevfileContent{
				Content: `commands:
- exec:
    commandLine: run command
    component: a-container
    group:
      isDefault: false
      kind: run
    hotReloadCapable: true
    workingDir: /path/to/work
  id: an-exec-command
metadata: {}
schemaVersion: 2.2.0
`,
				Commands: []Command{
					{
						Name:    "an-exec-command",
						Group:   "run",
						Default: false,
						Type:    "exec",
						Exec: ExecCommand{
							Component:        "a-container",
							CommandLine:      "run command",
							WorkingDir:       "/path/to/work",
							HotReloadCapable: true,
						},
					},
				},
				Containers: []Container{},
				Images:     []Image{},
				Resources:  []Resource{},
				Volumes:    []Volume{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.state(t)
			got, err := o.UnsetDefaultCommand(tt.args.commandName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DevfileState.UnsetDefaultCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DevfileState.UnsetDefaultCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
