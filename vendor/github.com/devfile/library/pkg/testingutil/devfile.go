package testingutil

import (
	"fmt"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components        []v1.Component
	ExecCommands      []v1.ExecCommand
	CompositeCommands []v1.CompositeCommand
	Commands          []v1.Command
	Events            v1.Events
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() devfilepkg.DevfileMetadata {
	return devfilepkg.DevfileMetadata{}
}

// SetMetadata sets metadata for the test devfile
func (d TestDevfileData) SetMetadata(name, version string) {}

// GetSchemaVersion gets the schema version for the test devfile
func (d TestDevfileData) GetSchemaVersion() string { return "testSchema" }

// SetSchemaVersion sets the schema version for the test devfile
func (d TestDevfileData) SetSchemaVersion(version string) {}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() *v1.Parent {
	return &v1.Parent{}
}

// SetParent is a mock function to set parent of the test devfile
func (d TestDevfileData) SetParent(parent *v1.Parent) {}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() v1.Events {
	return d.Events
}

// AddEvents is a mock function to add events to the test devfile
func (d TestDevfileData) AddEvents(events v1.Events) error { return nil }

// UpdateEvents is a mock function to update the events of the test devfile
func (d TestDevfileData) UpdateEvents(postStart, postStop, preStart, preStop []string) {}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents(options common.DevfileOptions) ([]v1.Component, error) {
	if len(options.Filter) == 0 {
		return d.Components, nil
	}

	var components []v1.Component
	for _, comp := range d.Components {
		filterIn, err := common.FilterDevfileObject(comp.Attributes, options)
		if err != nil {
			return nil, err
		}

		if filterIn {
			components = append(components, comp)
		}
	}

	return components, nil
}

// AddComponents is a mock function to add components to the test devfile
func (d TestDevfileData) AddComponents(components []v1.Component) error { return nil }

// UpdateComponent is a mock function to update the component of the test devfile
func (d TestDevfileData) UpdateComponent(component v1.Component) {}

// GetProjects is a mock function to get the projects from a test devfile
func (d TestDevfileData) GetProjects(options common.DevfileOptions) ([]v1.Project, error) {
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
	return []v1.Project{project1, project2}, nil

}

// AddProjects is a mock function to add projects to the test devfile
func (d TestDevfileData) AddProjects(projects []v1.Project) error { return nil }

// UpdateProject is a mock function to update a project for the test devfile
func (d TestDevfileData) UpdateProject(project v1.Project) {}

// GetStarterProjects is a mock function to get the starter projects from a test devfile
func (d TestDevfileData) GetStarterProjects(options common.DevfileOptions) ([]v1.StarterProject, error) {
	return []v1.StarterProject{}, nil
}

// AddStarterProjects is a mock func to add the starter projects to the test devfile
func (d TestDevfileData) AddStarterProjects(projects []v1.StarterProject) error {
	return nil
}

// UpdateStarterProject is a mock func to update the starter project for a test devfile
func (d TestDevfileData) UpdateStarterProject(project v1.StarterProject) {}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands(options common.DevfileOptions) ([]v1.Command, error) {

	var commands []v1.Command

	for _, command := range d.Commands {
		// we convert devfile command id to lowercase so that we can handle
		// cases efficiently without being error prone
		command.Id = strings.ToLower(command.Id)
		commands = append(commands, command)
	}

	return commands, nil
}

// AddCommands is a mock func that adds commands to the test devfile
func (d *TestDevfileData) AddCommands(commands ...v1.Command) error {
	devfileCommands, err := d.GetCommands(common.DevfileOptions{})
	if err != nil {
		return err
	}

	for _, command := range commands {
		id := command.Id
		for _, devfileCommand := range devfileCommands {
			if id == devfileCommand.Id {
				return fmt.Errorf("command %s already exist in the devfile", id)
			}
		}

		d.Commands = append(d.Commands, command)
	}
	return nil
}

// UpdateCommand is a mock func to update the command in a test devfile
func (d TestDevfileData) UpdateCommand(command v1.Command) {}

// AddVolume is a mock func that adds volume to the test devfile
func (d TestDevfileData) AddVolume(volumeComponent v1.Component, path string) error {
	return nil
}

// DeleteVolume is a mock func that deletes volume from the test devfile
func (d TestDevfileData) DeleteVolume(name string) error { return nil }

// GetVolumeMountPath is a mock func that gets the volume mount path of a container
func (d TestDevfileData) GetVolumeMountPath(name string) (string, error) {
	return "", nil
}

// GetDevfileContainerComponents gets the container components from the test devfile
func (d TestDevfileData) GetDevfileContainerComponents(options common.DevfileOptions) ([]v1.Component, error) {
	var components []v1.Component
	devfileComponents, err := d.GetComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range devfileComponents {
		if comp.Container != nil {
			components = append(components, comp)
		}
	}
	return components, nil
}

// GetDevfileVolumeComponents gets the volume components from the test devfile
func (d TestDevfileData) GetDevfileVolumeComponents(options common.DevfileOptions) ([]v1.Component, error) {
	var components []v1.Component
	devfileComponents, err := d.GetComponents(options)
	if err != nil {
		return nil, err
	}
	for _, comp := range devfileComponents {
		if comp.Volume != nil {
			components = append(components, comp)
		}
	}
	return components, nil
}

// GetDevfileWorkspace is a mock func to get the DevfileWorkspace in a test devfile
func (d TestDevfileData) GetDevfileWorkspace() *v1.DevWorkspaceTemplateSpecContent {
	return &v1.DevWorkspaceTemplateSpecContent{}
}

// SetDevfileWorkspace is a mock func to set the DevfileWorkspace in a test devfile
func (d TestDevfileData) SetDevfileWorkspace(content v1.DevWorkspaceTemplateSpecContent) {}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}

// GetFakeContainerComponent returns a fake container component for testing.
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
					VolumeMounts: []v1.VolumeMount{
						{
							Name: volumeName,
							Path: volumePath,
						},
					},
					MountSources: &mountSources,
				},
			},
		},
	}
}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) v1.Component {

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Volume: &v1.VolumeComponent{
				Volume: v1.Volume{
					Size: size,
				},
			},
		},
	}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []v1.ExecCommand {
	return []v1.ExecCommand{
		{
			CommandLine: "ls -a",
			Component:   "alias1",
			LabeledCommand: v1.LabeledCommand{
				BaseCommand: v1.BaseCommand{
					Group: &v1.CommandGroup{
						Kind: v1.RunCommandGroupKind,
					},
				},
			},

			WorkingDir: "/root",
		},
	}
}

// GetFakeEnv returns a fake env for testing
func GetFakeEnv(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

// GetFakeEnvParentOverride returns a fake envParentOverride for testing
func GetFakeEnvParentOverride(name, value string) v1.EnvVarParentOverride {
	return v1.EnvVarParentOverride{
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

// GetFakeVolumeMountParentOverride returns a fake volumeMountParentOverride for testing
func GetFakeVolumeMountParentOverride(name, path string) v1.VolumeMountParentOverride {
	return v1.VolumeMountParentOverride{
		Name: name,
		Path: path,
	}
}
