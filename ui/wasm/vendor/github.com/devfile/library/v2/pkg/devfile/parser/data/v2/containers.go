//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"fmt"
	"strconv"
	"strings"

	v1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	corev1 "k8s.io/api/core/v1"
)

// AddEnvVars accepts a map of container name mapped to an array of the env vars to be set;
// it adds the envirnoment variables to a given container name of the DevfileV2 object
// Example of containerEnvMap : {"runtime": {{Name: "Foo", Value: "Bar"}}}
func (d *DevfileV2) AddEnvVars(containerEnvMap map[string][]v1alpha2.EnvVar) error {
	components, err := d.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Env = merge(component.Container.Env, containerEnvMap[component.Name])
			_ = d.UpdateComponent(component)
		}
	}
	return nil
}

// RemoveEnvVars accepts a map of container name mapped to an array of environment variables to be removed;
// it removes the env vars from the specified container name of the DevfileV2 object
func (d *DevfileV2) RemoveEnvVars(containerEnvMap map[string][]string) error {
	components, err := d.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Env, err = removeEnvVarsFromList(component.Container.Env, containerEnvMap[component.Name])
			if err != nil {
				return err
			}
			_ = d.UpdateComponent(component)
		}
	}
	return nil
}

// SetPorts accepts a map of container name mapped to an array of port numbers to be set;
// it converts ports to endpoints, sets the endpoint to a given container name of the DevfileV2 object
// Example of containerPortsMap: {"runtime": {"8080", "9000"}, "wildfly": {"12956"}}
func (d *DevfileV2) SetPorts(containerPortsMap map[string][]string) error {
	components, err := d.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		endpoints, err := portsToEndpoints(containerPortsMap[component.Name]...)
		if err != nil {
			return err
		}
		if component.Container != nil {
			component.Container.Endpoints = addEndpoints(component.Container.Endpoints, endpoints)
			_ = d.UpdateComponent(component)
		}
	}

	return nil
}

// RemovePorts accepts a map of container name mapped to an array of port numbers to be removed;
// it removes the container endpoints with the specified port numbers of the specified container of the DevfileV2 object
// Example of containerPortsMap: {"runtime": {"8080", "9000"}, "wildfly": {"12956"}}
func (d *DevfileV2) RemovePorts(containerPortsMap map[string][]string) error {
	components, err := d.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			component.Container.Endpoints, err = removePortsFromList(component.Container.Endpoints, containerPortsMap[component.Name])
			if err != nil {
				return err
			}
			_ = d.UpdateComponent(component)
		}
	}

	return nil
}

// removeEnvVarsFromList removes the env variables based on the keys provided
// and returns a new EnvVarList
func removeEnvVarsFromList(envVarList []v1alpha2.EnvVar, keys []string) ([]v1alpha2.EnvVar, error) {
	// convert the array of envVarList to a map such that it can easily search for env var(s)
	// to remove from the component
	envVarListMap := map[string]bool{}
	for _, env := range envVarList {
		if !envVarListMap[env.Name] {
			envVarListMap[env.Name] = true
		}
	}

	// convert the array of keys to a map so that it can do a fast search for environment variable(s)
	// to remove from the component
	envVarToBeRemoved := map[string]bool{}
	// now check if the environment variable(s) requested for removal exists in
	// the env vars currently set in the component
	// if an env var requested for removal is not currently set, then raise an error
	// else add the env var to the envVarToBeRemoved map
	for _, key := range keys {
		if !envVarListMap[key] {
			return envVarList, fmt.Errorf("unable to find environment variable %s in the component", key)
		}
		envVarToBeRemoved[key] = true
	}

	// finally, let's remove the environment variables(s) requested by the user
	newEnvVarList := []v1alpha2.EnvVar{}
	for _, envVar := range envVarList {
		// if the env is in the keys(env var(s) to be removed), we skip it
		if envVarToBeRemoved[envVar.Name] {
			continue
		}
		newEnvVarList = append(newEnvVarList, envVar)
	}
	return newEnvVarList, nil
}

// removePortsFromList removes the ports from a given Endpoint list based on the provided port numbers
// and returns a new list of Endpoint
func removePortsFromList(endpoints []v1alpha2.Endpoint, ports []string) ([]v1alpha2.Endpoint, error) {
	// convert the array of Endpoint to a map such that it can easily search for port(s)
	// to remove from the component
	portInEndpoint := map[string]bool{}
	for _, ep := range endpoints {
		port := strconv.Itoa(ep.TargetPort)
		if !portInEndpoint[port] {
			portInEndpoint[port] = true
		}
	}

	// convert the array of ports to a map so that it can do a fast search for port(s)
	// to remove from the component
	portsToBeRemoved := map[string]bool{}

	// now check if the port(s) requested for removal exists in
	// the ports currently present in the component;
	// if a port requested for removal is not currently present, then raise an error
	// else add the port to the portsToBeRemoved map
	for _, port := range ports {
		if !portInEndpoint[port] {
			return endpoints, fmt.Errorf("unable to find port %q in the component", port)
		}
		portsToBeRemoved[port] = true
	}

	// finally, let's remove the port(s) requested by the user
	newEndpointsList := []v1alpha2.Endpoint{}
	for _, ep := range endpoints {
		// if the port is in the port(s)(to be removed), we skip it
		if portsToBeRemoved[strconv.Itoa(ep.TargetPort)] {
			continue
		}
		newEndpointsList = append(newEndpointsList, ep)
	}
	return newEndpointsList, nil
}

// merge merges the other EnvVarlist with keeping last value for duplicate EnvVars
// and returns a new EnvVarList
func merge(original []v1alpha2.EnvVar, other []v1alpha2.EnvVar) []v1alpha2.EnvVar {

	var dedupNewEvl []v1alpha2.EnvVar
	newEvl := append(original, other...)
	uniqueMap := make(map[string]string)
	// last value will be kept in case of duplicate env vars
	for _, envVar := range newEvl {
		uniqueMap[envVar.Name] = envVar.Value
	}

	for key, value := range uniqueMap {
		dedupNewEvl = append(dedupNewEvl, v1alpha2.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return dedupNewEvl

}

// portsToEndpoints converts an array of ports to an array of v1alpha2.Endpoint
func portsToEndpoints(ports ...string) ([]v1alpha2.Endpoint, error) {
	var endpoints []v1alpha2.Endpoint
	conPorts, err := getContainerPortsFromStrings(ports)
	if err != nil {
		return nil, err
	}
	for _, port := range conPorts {

		endpoint := v1alpha2.Endpoint{
			Name:       fmt.Sprintf("port-%d-%s", port.ContainerPort, strings.ToLower(string(port.Protocol))),
			TargetPort: int(port.ContainerPort),
			Protocol:   v1alpha2.EndpointProtocol(strings.ToLower(string(port.Protocol))),
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil

}

// addEndpoints appends two arrays of v1alpha2.Endpoint objects
func addEndpoints(current []v1alpha2.Endpoint, other []v1alpha2.Endpoint) []v1alpha2.Endpoint {
	newList := make([]v1alpha2.Endpoint, len(current))
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

// getContainerPortsFromStrings generates ContainerPort values from the array of string port values
// ports is the array containing the string port values
func getContainerPortsFromStrings(ports []string) ([]corev1.ContainerPort, error) {
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
