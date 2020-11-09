package testingutil

import (
	"strings"

	v1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components []v1.Component
	Commands   []v1.Command
	Events     v1.Events
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []v1.Component {
	return d.Components
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() devfilepkg.DevfileMetadata {
	return devfilepkg.DevfileMetadata{}
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() v1.Events {
	return d.Events
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() *v1.Parent {
	return &v1.Parent{}
}

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []v1.Project {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"test-project/", "anotherproject/"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}

	project1 := v1.Project{
		ClonePath: clonePath[0],
		Name:      projectName[0],
		ProjectSource: v1.ProjectSource{
			Git: &v1.GitProjectSource{
				GitLikeProjectSource: v1.GitLikeProjectSource{
					Remotes: map[string]string{
						"origin": sourceLocation[0],
					},
				},
			},
		},
	}

	project2 := v1.Project{
		ClonePath: clonePath[1],
		Name:      projectName[1],
		ProjectSource: v1.ProjectSource{
			Git: &v1.GitProjectSource{
				GitLikeProjectSource: v1.GitLikeProjectSource{
					Remotes: map[string]string{
						"origin": sourceLocation[1],
					},
				},
			},
		},
	}
	return []v1.Project{project1, project2}

}

// GetStarterProjects returns the fake starter projects
func (d TestDevfileData) GetStarterProjects() []v1.StarterProject {
	return []v1.StarterProject{}
}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands() map[string]v1.Command {

	commands := make(map[string]v1.Command, len(d.Commands))

	for _, command := range d.Commands {
		// we convert devfile command id to lowercase so that we can handle
		// cases efficiently without being error prone
		// we also convert the odo push commands from build-command and run-command flags
		command.Id = strings.ToLower(command.Id)
		commands[command.Id] = command
	}

	return commands
}

func (d TestDevfileData) AddVolume(volume v1.Component, path string) error {
	return nil
}

func (d TestDevfileData) DeleteVolume(name string) error { return nil }

func (d TestDevfileData) GetVolumeMountPath(name string) (string, error) {
	return "", nil
}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}

// SetMetadata sets metadata for devfile
func (d TestDevfileData) SetMetadata(name, version string) {}

// SetSchemaVersion sets schema version for devfile
func (d TestDevfileData) SetSchemaVersion(version string) {}

func (d TestDevfileData) AddComponents(components []v1.Component) error { return nil }

func (d TestDevfileData) UpdateComponent(component v1.Component) {}

func (d TestDevfileData) UpdateCommand(command v1.Command) {}

func (d TestDevfileData) SetEvents(events v1.Events) {}

func (d TestDevfileData) AddProjects(projects []v1.Project) error { return nil }

func (d TestDevfileData) UpdateProject(project v1.Project) {}

func (d TestDevfileData) AddEvents(events v1.Events) error { return nil }

func (d TestDevfileData) UpdateEvents(postStart, postStop, preStart, preStop []string) {}

func (d TestDevfileData) SetParent(parent *v1.Parent) {}

func (d *TestDevfileData) AddCommands(commands ...v1.Command) error {
	commandsMap := d.GetCommands()

	for _, command := range commands {
		id := command.Id
		if _, ok := commandsMap[id]; !ok {
			d.Commands = append(d.Commands, command)
		} else {
			return &common.FieldAlreadyExistError{Name: id, Field: "command"}
		}
	}
	return nil
}

func (d TestDevfileData) AddStarterProjects(projects []v1.StarterProject) error {
	return nil
}

func (d TestDevfileData) UpdateStarterProject(project v1.StarterProject) {}

// GetFakeContainerComponent returns a fake container component for testing
func GetFakeContainerComponent(name string) v1.Component {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"
	mountSources := true

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Container: &v1.ContainerComponent{
				Container: v1.Container{
					Image:       image,
					Env:         []v1.EnvVar{},
					MemoryLimit: memoryLimit,
					VolumeMounts: []v1.VolumeMount{{
						Name: volumeName,
						Path: volumePath,
					}},
					MountSources: &mountSources,
				},
			}}}

}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) v1.Component {
	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Volume: &v1.VolumeComponent{
				Volume: v1.Volume{
					Size: size,
				}}}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []v1.Command {
	return []v1.Command{
		{
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					LabeledCommand: v1.LabeledCommand{
						BaseCommand: v1.BaseCommand{
							Group: &v1.CommandGroup{
								Kind: v1.RunCommandGroupKind,
							},
						},
					},
					CommandLine: "ls -a",
					Component:   "alias1",
					WorkingDir:  "/root",
				},
			},
		},
	}
}

// GetFakeExecRunCommands returns a fake env for testing
func GetFakeEnv(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

// GetFakeVolumeMount returns a fake volume mount for testing
func GetFakeVolumeMount(name, path string) v1.VolumeMount {
	return v1.VolumeMount{
		Name: name,
		Path: path,
	}
}
