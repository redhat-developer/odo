package testingutil

import (
	"github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/common"
	versionsCommon "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	Components   []versionsCommon.DevfileComponent
	ExecCommands []versionsCommon.Exec
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []versionsCommon.DevfileComponent {
	return d.GetAliasedComponents()
}

// GetEvents is a mock function to get events from devfile
func (d TestDevfileData) GetEvents() versionsCommon.DevfileEvents {
	return versionsCommon.DevfileEvents{}
}

// GetMetadata is a mock function to get metadata from devfile
func (d TestDevfileData) GetMetadata() versionsCommon.DevfileMetadata {
	return versionsCommon.DevfileMetadata{}
}

// GetParent is a mock function to get parent from devfile
func (d TestDevfileData) GetParent() versionsCommon.DevfileParent {
	return versionsCommon.DevfileParent{}
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []versionsCommon.DevfileComponent {
	var aliasedComponents = []common.DevfileComponent{}

	for _, comp := range d.Components {
		if comp.Container != nil {
			if comp.Container.Name != "" {
				aliasedComponents = append(aliasedComponents, comp)
			}
		}
	}
	return aliasedComponents

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

	var commands []versionsCommon.DevfileCommand

	for i := range d.ExecCommands {
		commands = append(commands, versionsCommon.DevfileCommand{Exec: &d.ExecCommands[i]})
	}

	return commands

}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}

// GetFakeComponent returns fake component for testing
func GetFakeComponent(name string) versionsCommon.DevfileComponent {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"

	return versionsCommon.DevfileComponent{
		Container: &versionsCommon.Container{
			Name:        name,
			Image:       image,
			Env:         []versionsCommon.Env{},
			MemoryLimit: memoryLimit,
			VolumeMounts: []versionsCommon.VolumeMount{{
				Name: volumeName,
				Path: volumePath,
			}},
			MountSources: true,
		}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []versionsCommon.Exec {
	return []versionsCommon.Exec{
		{
			CommandLine: "ls -a",
			Component:   "alias1",
			Group: &versionsCommon.Group{
				Kind: versionsCommon.RunCommandGroupType,
			},
			WorkingDir: "/root",
		},
	}
}
