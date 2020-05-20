package testingutil

import (
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	ComponentType       versionsCommon.DevfileComponentType
	CommandActions      []versionsCommon.Exec
	MissingInitCommand  bool
	MissingBuildCommand bool
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []versionsCommon.DevfileComponent {
	return d.GetAliasedComponents()
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []versionsCommon.DevfileComponent {
	name := [...]string{"alias1", "alias2"}
	image := [...]string{"docker.io/maven:latest", "docker.io/hello-world:latest"}
	memoryLimit := "128Mi"
	volumeName := [...]string{"myvolume1", "myvolume2"}
	volumePath := [...]string{"/my/volume/mount/path1", "/my/volume/mount/path2"}

	component1 := versionsCommon.DevfileComponent{
		Container: &versionsCommon.Container{
			Name:        name[0],
			Image:       image[0],
			Env:         []versionsCommon.Env{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []versionsCommon.VolumeMount{{
				Name: volumeName[0],
				Path: volumePath[0],
			}},
			MountSources: true,
		}}

	component2 := versionsCommon.DevfileComponent{
		Container: &versionsCommon.Container{
			Name:        name[0],
			Image:       image[0],
			Env:         []versionsCommon.Env{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []versionsCommon.VolumeMount{
				{
					Name: volumeName[0],
					Path: volumePath[0],
				},
				{
					Name: volumeName[1],
					Path: volumePath[1],
				}},
			MountSources: true,
		}}

	return []versionsCommon.DevfileComponent{component1, component2}

}

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []versionsCommon.DevfileProject {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"/test-project", "/anotherproject"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}

	project1 := versionsCommon.DevfileProject{
		ClonePath: clonePath[0],
		Name:      projectName[0],
		Git: &versionsCommon.Git{
			Location: sourceLocation[0],
		},
	}

	project2 := versionsCommon.DevfileProject{
		ClonePath: clonePath[1],
		Name:      projectName[1],
		Git: &versionsCommon.Git{
			Location: sourceLocation[1],
		},
	}
	return []versionsCommon.DevfileProject{project1, project2}

}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands() []versionsCommon.DevfileCommand {
	commandName := [...]string{"devinit", "devbuild", "devrun", "customcommand"}

	command1 := versionsCommon.DevfileCommand{
		Exec: &versionsCommon.Exec{
			Id: commandName[2],
		},
	}

	command2 := versionsCommon.DevfileCommand{
		Exec: &versionsCommon.Exec{
			Id: commandName[3],
		},
	}

	commands := []versionsCommon.DevfileCommand{command1, command2}

	if !d.MissingInitCommand {
		commands = append(commands, versionsCommon.DevfileCommand{
			Exec: &versionsCommon.Exec{
				Id: commandName[0],
			}})
	}
	if !d.MissingBuildCommand {
		commands = append(commands, versionsCommon.DevfileCommand{
			Exec: &versionsCommon.Exec{
				Id: commandName[1],
			}})
	}

	return commands
}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}
