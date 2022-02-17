package parser

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

const (
	Name              = "Name"
	Ports             = "Ports"
	Memory            = "Memory"
	PortsDescription  = "Ports to be opened in all component containers"
	MemoryDescription = "The Maximum memory all the component containers can consume"
	NameDescription   = "The name of the component"
)

// SetMetadataName set metadata name in a devfile
func (d DevfileObj) SetMetadataName(name string) error {
	metadata := d.Data.GetMetadata()
	metadata.Name = name
	d.Data.SetMetadata(metadata)
	return d.WriteYamlDevfile()
}

// AddEnvVars accepts a map of container name mapped to an array of the env vars to be set;
// it adds the envirnoment variables to a given container name, and writes to the devfile
// Example of containerEnvMap : {"runtime": {{Name: "Foo", Value: "Bar"}}}
func (d DevfileObj) AddEnvVars(containerEnvMap map[string][]v1.EnvVar) error {
	err := d.Data.AddEnvVars(containerEnvMap)
	if err != nil {
		return err
	}
	return d.WriteYamlDevfile()
}

// RemoveEnvVars accepts a map of container name mapped to an array of environment variables to be removed;
// it removes the env vars from the specified container name and writes it to the devfile
func (d DevfileObj) RemoveEnvVars(containerEnvMap map[string][]string) (err error) {
	err = d.Data.RemoveEnvVars(containerEnvMap)
	if err != nil {
		return err
	}
	return d.WriteYamlDevfile()
}

// SetPorts accepts a map of container name mapped to an array of port numbers to be set;
// it converts ports to endpoints, sets the endpoint to a given container name, and writes to the devfile
// Example of containerPortsMap: {"runtime": {"8080", "9000"}, "wildfly": {"12956"}}
func (d DevfileObj) SetPorts(containerPortsMap map[string][]string) error {
	err := d.Data.SetPorts(containerPortsMap)
	if err != nil {
		return err
	}
	return d.WriteYamlDevfile()
}

// RemovePorts accepts a map of container name mapped to an array of port numbers to be removed;
// it removes the container endpoints with the specified port numbers of the specified container, and writes to the devfile
// Example of containerPortsMap: {"runtime": {"8080", "9000"}, "wildfly": {"12956"}}
func (d DevfileObj) RemovePorts(containerPortsMap map[string][]string) error {
	err := d.Data.RemovePorts(containerPortsMap)
	if err != nil {
		return err
	}
	return d.WriteYamlDevfile()
}

// HasPorts checks if a devfile contains container endpoints
func (d DevfileObj) HasPorts() bool {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return false
	}
	for _, component := range components {
		if component.Container != nil {
			if len(component.Container.Endpoints) > 0 {
				return true
			}
		}
	}
	return false
}

// SetMemory sets memoryLimit in devfile container
func (d DevfileObj) SetMemory(memory string) error {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.MemoryLimit = memory
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

// GetMemory gets memoryLimit from devfile container
func (d DevfileObj) GetMemory() string {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return ""
	}
	for _, component := range components {
		if component.Container != nil {
			if component.Container.MemoryLimit != "" {
				return component.Container.MemoryLimit
			}
		}

	}
	return ""
}

// GetMetadataName gets metadata name from a devfile
func (d DevfileObj) GetMetadataName() string {
	return d.Data.GetMetadata().Name
}
