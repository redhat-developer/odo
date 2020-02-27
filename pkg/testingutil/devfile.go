package testingutil

import (
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
)

// TestDevfileData is a convenience data type used to mock up a devfile configuration
type TestDevfileData struct {
	ComponentType versionsCommon.DevfileComponentType
}

// GetComponents is a mock function to get the components from a devfile
func (d TestDevfileData) GetComponents() []versionsCommon.DevfileComponent {
	return d.GetAliasedComponents()
}

// GetAliasedComponents is a mock function to get the components that have an alias from a devfile
func (d TestDevfileData) GetAliasedComponents() []versionsCommon.DevfileComponent {
	alias := "alias"
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	return []versionsCommon.DevfileComponent{
		{
			Alias: &alias,
			DevfileComponentDockerimage: versionsCommon.DevfileComponentDockerimage{
				Image:       &image,
				Command:     []string{},
				Args:        []string{},
				Env:         []versionsCommon.DockerimageEnv{},
				MemoryLimit: &memoryLimit,
			},
			Type: d.ComponentType,
		},
	}
}

// Validate is a mock validation that always validates without error
func (d TestDevfileData) Validate() error {
	return nil
}
