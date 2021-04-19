package parser

import (
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	corev1 "k8s.io/api/core/v1"
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

// AddEnvVars adds environment variables to all the components in a devfile
func (d DevfileObj) AddEnvVars(otherList []v1.EnvVar) error {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Env = Merge(component.Container.Env, otherList)
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

// RemoveEnvVars removes the environment variables which have the keys from all the components in a devfile
func (d DevfileObj) RemoveEnvVars(keys []string) (err error) {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Env, err = RemoveEnvVarsFromList(component.Container.Env, keys)
			if err != nil {
				return err
			}
			d.Data.UpdateComponent(component)
		}
	}
	return d.WriteYamlDevfile()
}

// SetPorts converts ports to endpoints, adds to a devfile
func (d DevfileObj) SetPorts(ports ...string) error {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
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

// RemovePorts removes all container endpoints from a devfile
func (d DevfileObj) RemovePorts() error {
	components, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Endpoints = []v1.Endpoint{}
			d.Data.UpdateComponent(component)
		}
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

func portsToEndpoints(ports ...string) ([]v1.Endpoint, error) {
	var endpoints []v1.Endpoint
	conPorts, err := GetContainerPortsFromStrings(ports)
	if err != nil {
		return nil, err
	}
	for _, port := range conPorts {

		endpoint := v1.Endpoint{
			Name:       fmt.Sprintf("port-%d-%s", port.ContainerPort, strings.ToLower(string(port.Protocol))),
			TargetPort: int(port.ContainerPort),
			Protocol:   v1.EndpointProtocol(strings.ToLower(string(port.Protocol))),
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil

}

func addEndpoints(current []v1.Endpoint, other []v1.Endpoint) []v1.Endpoint {
	newList := make([]v1.Endpoint, len(current))
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

// GetContainerPortsFromStrings generates ContainerPort values from the array of string port values
// ports is the array containing the string port values
func GetContainerPortsFromStrings(ports []string) ([]corev1.ContainerPort, error) {
	var containerPorts []corev1.ContainerPort
	for _, port := range ports {
		splits := strings.Split(port, "/")
		if len(splits) < 1 || len(splits) > 2 {
			return nil, fmt.Errorf("unable to parse the port string %s", port)
		}

		portNumberI64, err := strconv.ParseInt(splits[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid port number %s", splits[0])
		}
		portNumber := int32(portNumberI64)

		var portProto corev1.Protocol
		if len(splits) == 2 {
			switch strings.ToUpper(splits[1]) {
			case "TCP":
				portProto = corev1.ProtocolTCP
			case "UDP":
				portProto = corev1.ProtocolUDP
			default:
				return nil, fmt.Errorf("invalid port protocol %s", splits[1])
			}
		} else {
			portProto = corev1.ProtocolTCP
		}

		port := corev1.ContainerPort{
			Name:          fmt.Sprintf("%d-%s", portNumber, strings.ToLower(string(portProto))),
			ContainerPort: portNumber,
			Protocol:      portProto,
		}
		containerPorts = append(containerPorts, port)
	}
	return containerPorts, nil
}

// RemoveEnvVarsFromList removes the env variables based on the keys provided
// and returns a new EnvVarList
func RemoveEnvVarsFromList(envVarList []v1.EnvVar, keys []string) ([]v1.EnvVar, error) {
	// convert the envVarList map to an array to easily search for env var(s)
	// to remove from the component
	envVarListArray := []string{}
	for _, env := range envVarList {
		envVarListArray = append(envVarListArray, env.Name)
	}

	// now check if the environment variable(s) requested for removal exists in
	// the env vars set for the component by odo
	for _, key := range keys {
		if !InArray(envVarListArray, key) {
			return nil, fmt.Errorf("unable to find environment variable %s in the component", key)
		}
	}

	// finally, let's remove the environment variables(s) requested by the user
	newEnvVarList := []v1.EnvVar{}
	for _, envVar := range envVarList {
		// if the env is in the keys we skip it
		if InArray(keys, envVar.Name) {
			continue
		}
		newEnvVarList = append(newEnvVarList, envVar)
	}
	return newEnvVarList, nil
}

// Merge merges the other EnvVarlist with keeping last value for duplicate EnvVars
// and returns a new EnvVarList
func Merge(original []v1.EnvVar, other []v1.EnvVar) []v1.EnvVar {

	var dedupNewEvl []v1.EnvVar
	newEvl := append(original, other...)
	uniqueMap := make(map[string]string)
	// last value will be kept in case of duplicate env vars
	for _, envVar := range newEvl {
		uniqueMap[envVar.Name] = envVar.Value
	}

	for key, value := range uniqueMap {
		dedupNewEvl = append(dedupNewEvl, v1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return dedupNewEvl

}

// In checks if the value is in the array
func InArray(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}
