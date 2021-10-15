package config

import (
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

const (
	// Name is the name of the setting controlling the component name
	Name = "Name"
	// NameDescription is human-readable description of the Name setting
	NameDescription = "The name of the component"
	// Memory is the name of the setting controlling the memory a component consumes
	Memory = "Memory"
	// MemoryDescription is the description of the setting controlling the min and max memory to same value
	MemoryDescription = "The minimum and maximum memory a component can consume"
	// Ports is the space separated list of user specified ports to be opened in the component
	Ports = "Ports"
	// PortsDescription is the description of the ports component setting
	PortsDescription = "Ports to be opened in the component"
)

var (
	supportedDevfileParameterDescriptions = map[string]string{
		Name:   NameDescription,
		Ports:  PortsDescription,
		Memory: MemoryDescription,
	}
	lowerCaseDevfileParameters = util.GetLowerCaseParameters(GetDevfileSupportedParameters())
)

// FormatDevfileSupportedParameters outputs supported parameters and their description
func FormatDevfileSupportedParameters() (result string) {
	for _, v := range GetDevfileSupportedParameters() {
		result = result + " " + v + " - " + supportedDevfileParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Parameters for Devfile Components:\n" + result
}

func GetDevfileSupportedParameters() []string {
	return util.GetSortedKeys(supportedDevfileParameterDescriptions)
}

// AsDevfileSupportedParameter returns the parameter in lower case and a boolean indicating if it is a supported parameter
func AsDevfileSupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseDevfileParameters[lower]
}

// SetDevfileConfiguration allows setting all the parameters that are configurable in a devfile
func SetDevfileConfiguration(d parser.DevfileObj, parameter string, value interface{}) error {

	// we are ignoring this error becase a developer is usually aware of the type of value that is
	// being passed. So consider this a shortcut, if you know its a string value use this strValue
	// else parse it inside the switch case.
	strValue, _ := value.(string)
	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.SetMetadataName(strValue)
		case "ports":
			arrValue := strings.Split(strValue, ",")
			return d.SetPorts(arrValue...)
		case "memory":
			return d.SetMemory(strValue)
		}

	}
	return errors.Errorf("unknown parameter :'%s' is not a configurable parameter in the devfile", parameter)

}

// DeleteConfiguration allows deleting  the parameters that are configurable in a devfile
func DeleteDevfileConfiguration(d parser.DevfileObj, parameter string) error {
	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.SetMetadataName("")
		case "ports":
			return d.RemovePorts()
		case "memory":
			return d.SetMemory("")
		}
	}
	return errors.Errorf("unknown parameter :'%s' is not a configurable parameter in the devfile", parameter)
}

// IsSet checks if a parameter is set in the devfile
func IsSetInDevfile(d parser.DevfileObj, parameter string) bool {

	if parameter, ok := AsDevfileSupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			return d.GetMetadataName() != ""
		case "ports":
			return d.HasPorts()
		case "memory":
			return d.GetMemory() != ""
		}
	}
	return false

}
