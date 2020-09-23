package testingutil

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components []common.DevfileComponent
	Commands   []common.DevfileCommand
	Events     common.DevfileEvents
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []common.DevfileComponent {
	return d.Components
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() common.DevfileMetadata {
	return common.DevfileMetadata{}
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() common.DevfileEvents {
	return d.Events
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() common.DevfileParent {
	return common.DevfileParent{}
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []common.DevfileComponent {
	var aliasedComponents = []common.DevfileComponent{}

	for _, comp := range d.Components {
		if comp.Container != nil {
			if comp.Name != "" {
				aliasedComponents = append(aliasedComponents, comp)
			}
		}
	}
	return aliasedComponents

}

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []common.DevfileProject {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"test-project/", "anotherproject/"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}

	project1 := common.DevfileProject{
		ClonePath: clonePath[0],
		Name:      projectName[0],
		Git: &common.Git{
			GitLikeProjectSource: common.GitLikeProjectSource{
				Remotes: map[string]string{"origin": sourceLocation[0]},
			},
		},
	}

	project2 := common.DevfileProject{
		ClonePath: clonePath[1],
		Name:      projectName[1],
		Git: &common.Git{
			GitLikeProjectSource: common.GitLikeProjectSource{
				Remotes: map[string]string{"origin": sourceLocation[1]},
			},
		},
	}
	return []common.DevfileProject{project1, project2}

}

// GetStarterProjects returns the fake starter projects
func (d TestDevfileData) GetStarterProjects() []common.DevfileStarterProject {
	return []common.DevfileStarterProject{}
}

// GetCommands is a mock function to get the commands from a devfile
func (d *TestDevfileData) GetCommands() map[string]common.DevfileCommand {
	commands := make(map[string]common.DevfileCommand, len(d.Commands))

	for _, command := range d.Commands {
		// we convert devfile command id to lowercase so that we can handle
		// cases efficiently without being error prone
		// we also convert the odo push commands from build-command and run-command flags
		commands[command.SetIDToLower()] = command

	}

	return commands
}

func (d TestDevfileData) AddVolume(volumeComponent common.DevfileComponent, path string) error {
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

func (d TestDevfileData) AddComponents(components []common.DevfileComponent) error { return nil }

func (d TestDevfileData) UpdateComponent(component common.DevfileComponent) {}

func (d *TestDevfileData) AddCommands(commands ...common.DevfileCommand) error {
	commandsMap := d.GetCommands()

	for _, command := range commands {
		id := command.GetID()
		if _, ok := commandsMap[id]; !ok {
			d.Commands = append(d.Commands, command)
		} else {
			return &common.AlreadyExistError{Name: id, Field: "command"}
		}
	}
	return nil
}

func (d TestDevfileData) UpdateCommand(command common.DevfileCommand) {}

func (d TestDevfileData) SetEvents(events common.DevfileEvents) {}

func (d TestDevfileData) AddProjects(projects []common.DevfileProject) error { return nil }

func (d TestDevfileData) UpdateProject(project common.DevfileProject) {}

func (d TestDevfileData) AddStarterProjects(projects []common.DevfileStarterProject) error {
	return nil
}

func (d TestDevfileData) UpdateStarterProject(project common.DevfileStarterProject) {}

func (d TestDevfileData) AddEvents(events common.DevfileEvents) error { return nil }

func (d TestDevfileData) UpdateEvents(postStart, postStop, preStart, preStop []string) {}

func (d TestDevfileData) SetParent(parent common.DevfileParent) {}

// GetFakeContainerComponent returns a fake container component for testing
func GetFakeContainerComponent(name string) common.DevfileComponent {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"

	return common.DevfileComponent{
		Name: name,
		Container: &common.Container{
			Image:       image,
			Env:         []common.Env{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []common.VolumeMount{{
				Name: volumeName,
				Path: volumePath,
			}},
			MountSources: true,
		}}

}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) common.DevfileComponent {
	return common.DevfileComponent{
		Name: name,
		Volume: &common.Volume{
			Size: size,
		}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []common.DevfileCommand {
	return []common.DevfileCommand{
		{
			Exec: &common.Exec{
				CommandLine: "ls -a",
				Component:   "alias1",
				Group: &common.Group{
					Kind: common.RunCommandGroupType,
				},
				WorkingDir: "/root",
			},
		},
	}
}

// GetFakeExecRunCommands returns a fake env for testing
func GetFakeEnv(name, value string) common.Env {
	return common.Env{
		Name:  name,
		Value: value,
	}
}

// GetFakeVolumeMount returns a fake volume mount for testing
func GetFakeVolumeMount(name, path string) common.VolumeMount {
	return common.VolumeMount{
		Name: name,
		Path: path,
	}
}
