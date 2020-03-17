package testingutil

import (
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	ComponentType       versionsCommon.DevfileComponentType
	CommandActions      []versionsCommon.DevfileCommandAction
	MissingBuildCommand bool
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []versionsCommon.DevfileComponent {
	return d.GetAliasedComponents()
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []versionsCommon.DevfileComponent {
	alias := [...]string{"alias1", "alias2"}
	image := [...]string{"docker.io/maven:latest", "docker.io/hello-world:latest"}
	memoryLimit := "128Mi"
	volumeName := [...]string{"myvolume1", "myvolume2"}
	volumePath := [...]string{"/my/volume/mount/path1", "/my/volume/mount/path2"}
	return []versionsCommon.DevfileComponent{
		{
			Alias: &alias[0],
			DevfileComponentDockerimage: versionsCommon.DevfileComponentDockerimage{
				Image:       &image[0],
				Command:     []string{},
				Args:        []string{},
				Env:         []versionsCommon.DockerimageEnv{},
				MemoryLimit: &memoryLimit,
				Volumes: []versionsCommon.DockerimageVolume{
					{
						Name:          &volumeName[0],
						ContainerPath: &volumePath[0],
					},
				},
			},
			Type: d.ComponentType,
		},
		{
			Alias: &alias[1],
			DevfileComponentDockerimage: versionsCommon.DevfileComponentDockerimage{
				Image:       &image[1],
				Command:     []string{},
				Args:        []string{},
				Env:         []versionsCommon.DockerimageEnv{},
				MemoryLimit: &memoryLimit,
				Volumes: []versionsCommon.DockerimageVolume{
					{
						Name:          &volumeName[0],
						ContainerPath: &volumePath[0],
					},
					{
						Name:          &volumeName[1],
						ContainerPath: &volumePath[1],
					},
				},
			},
			Type: d.ComponentType,
		},
	}
}

// GetProjects is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetProjects() []versionsCommon.DevfileProject {
	projectName := [...]string{"test-project", "anotherproject"}
	clonePath := [...]string{"/test-project", "/anotherproject"}
	sourceLocation := [...]string{"https://github.com/someproject/test-project.git", "https://github.com/another/project.git"}
	return []versionsCommon.DevfileProject{
		{
			ClonePath: &clonePath[0],
			Name:      projectName[0],
			Source: versionsCommon.DevfileProjectSource{
				Type:     versionsCommon.DevfileProjectTypeGit,
				Location: sourceLocation[0],
			},
		},
		{
			ClonePath: &clonePath[1],
			Name:      projectName[1],
			Source: versionsCommon.DevfileProjectSource{
				Type:     versionsCommon.DevfileProjectTypeGit,
				Location: sourceLocation[1],
			},
		},
	}
}

// GetCommands is a mock function to get the commands from a devfile
func (d TestDevfileData) GetCommands() []versionsCommon.DevfileCommand {
	commandName := [...]string{"devbuild", "devrun", "customcommand"}

	commands := []versionsCommon.DevfileCommand{
		{
			Name:    commandName[1],
			Actions: d.CommandActions,
		},
		{
			Name:    commandName[2],
			Actions: d.CommandActions,
		},
	}

	if !d.MissingBuildCommand {
		commands = append(commands, versionsCommon.DevfileCommand{
			Name:    commandName[0],
			Actions: d.CommandActions,
		})
	}

	return commands
}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}
