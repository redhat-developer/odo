package parser

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

// This file contains all the parameters that can be change via odo config

const (
	Name              = "Name"
	Ports             = "Ports"
	Memory            = "Memory"
	PortsDescription  = "Ports to be opened in all component containers"
	MemoryDescription = "The Maximum memory all the component containers can consume"
	NameDescription   = "The name of the component"
)

var (
	supportedDevfileParameterDescriptions = map[string]string{
		Name:   NameDescription,
		Ports:  PortsDescription,
		Memory: MemoryDescription,
	}

	lowerCaseDevfileParameters = util.GetLowerCaseParameters(GetDevfileSupportedParameters())
)

// SetConfiguration allows setting all the parameters that are configurable in a devfile
func (d DevfileObj) SetConfiguration(parameter string, value interface{}) error {

	// we are ignoring this error becase a developer is usually aware of the type of value that is
	// being passed. So consider this a shortcut, if you know its a string value use this strValue
	// else parse it inside the switch case.
	strValue, _ := value.(string)
	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.setMetadataName(strValue)
		case "ports":
			arrValue := strings.Split(strValue, ",")
			return d.setPorts(arrValue...)
		case "memory":
			return d.setMemory(strValue)
		}

	}
	return errors.Errorf("unknown parameter :'%s' is not a configurable parameter in the devfile", parameter)

}

// DeleteConfiguration allows deleting  the parameters that are configurable in a devfile
func (d DevfileObj) DeleteConfiguration(parameter string) error {
	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.setMetadataName("")
		case "ports":
			return d.removePorts()
		case "memory":
			return d.setMemory("")
		}
	}
	return errors.Errorf("unknown parameter :'%s' is not a configurable parameter in the devfile", parameter)
}

// IsSet checks if a parameter is set in the devfile
func (d DevfileObj) IsSet(parameter string) bool {

	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.getMetadataName() != ""
		case "ports":
			return d.hasPorts()
		case "memory":
			return d.getMemory() != ""
		}
	}
	return false

}

func (d DevfileObj) setMetadataName(name string) error {
	metadata := d.Data.GetMetadata()
	d.Data.SetMetadata(name, metadata.Version)
	return d.WriteYamlDevfile()
}

// AddEnvVars adds environment variables to all the components in a devfile
func (d DevfileObj) AddEnvVars(otherList config.EnvVarList) error {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {
			currentlist := config.NewEnvVarListFromDevfileEnv(component.Container.Env)
			component.Container.Env = currentlist.Merge(otherList).ToDevfileEnv()
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

// RemoveEnvVars removes the environment variables which have the keys from all the components in a devfile
func (d DevfileObj) RemoveEnvVars(keys []string) error {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {

			currentlist := config.NewEnvVarListFromDevfileEnv(component.Container.Env)
			envList, err := config.RemoveEnvVarsFromList(currentlist, keys)
			if err != nil {
				return err
			}
			component.Container.Env = envList.ToDevfileEnv()
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

func (d DevfileObj) setPorts(ports ...string) error {
	components := d.Data.GetComponents()
	endpoints, err := portsToEndpoints(ports...)
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Endpoints = addEndpoints(component.Container.Endpoints, endpoints)
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

func (d DevfileObj) removePorts() error {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {
			component.Container.Endpoints = []common.Endpoint{}
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

func (d DevfileObj) hasPorts() bool {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {
			if len(component.Container.Endpoints) > 0 {
				return true
			}
		}
	}
	return false
}

func (d DevfileObj) setMemory(memory string) error {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {
			component.Container.MemoryLimit = memory
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}
func (d DevfileObj) getMemory() string {
	components := d.Data.GetComponents()
	for _, component := range components {
		if component.Container != nil {
			if component.Container.MemoryLimit != "" {
				return component.Container.MemoryLimit
			}
		}

	}
	return ""
}

func (d DevfileObj) getMetadataName() string {
	return d.Data.GetMetadata().Name
}

// FormatDevfileSupportedParameters outputs supported parameters and their description
func FormatDevfileSupportedParameters() (result string) {
	for _, v := range GetDevfileSupportedParameters() {
		result = result + " " + v + " - " + supportedDevfileParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Parameters for Devfile Components:\n" + result
}

// AsDevfileSupportedParameter returns the parameter in lower case and a boolean indicating if it is a supported parameter
func AsDevfileSupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseDevfileParameters[lower]
}

// GetDevfileSupportedParameters returns the name of the supported global parameters
func GetDevfileSupportedParameters() []string {
	return util.GetSortedKeys(supportedDevfileParameterDescriptions)
}

func portsToEndpoints(ports ...string) ([]common.Endpoint, error) {
	var endpoints []common.Endpoint
	conPorts, err := util.GetContainerPortsFromStrings(ports)
	if err != nil {
		return nil, err
	}
	for _, port := range conPorts {

		endpoint := common.Endpoint{
			Name:       fmt.Sprintf("port-%d-%s", port.ContainerPort, strings.ToLower(string(port.Protocol))),
			TargetPort: port.ContainerPort,
			Protocol:   common.ProtocolType(strings.ToLower(string(port.Protocol))),
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil

}

func addEndpoints(current []common.Endpoint, other []common.Endpoint) []common.Endpoint {
	newList := make([]common.Endpoint, len(current))
	copy(newList, current)
	for _, ep := range other {
		present := false

		for _, presentep := range newList {

			protocol := presentep.Protocol
			if protocol == "" {
				// endpoint protocol default value is http
				protocol = "http"
			}
			// if the target port and protocol match, we add a case where the protocol is not provided and hence we assume that to be "tcp"
			if presentep.TargetPort == ep.TargetPort && (ep.Protocol == protocol) {
				present = true
				break
			}
		}
		if !present {
			newList = append(newList, ep)
		}
	}

	return newList
}
