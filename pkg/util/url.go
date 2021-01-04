package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// GetValidPortNumber checks if the given port number is a valid component port or not
// if port number is not provided and the component is a single port component, the component port is returned
// port number is -1 if the user does not specify any port
func GetValidPortNumber(componentName string, portNumber int, portList []string) (int, error) {
	var componentPorts []int
	for _, p := range portList {
		port, err := strconv.Atoi(strings.Split(p, "/")[0])
		if err != nil {
			return port, err
		}
		componentPorts = append(componentPorts, port)
	}
	// port number will be -1 if the user doesn't specify any port
	if portNumber == -1 {
		switch {
		case len(componentPorts) > 1:
			return portNumber, errors.Errorf("port for the component %s is required as it exposes %d ports: %s", componentName, len(componentPorts), strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
		case len(componentPorts) == 1:
			return componentPorts[0], nil
		default:
			return portNumber, errors.Errorf("no port is exposed by the component %s", componentName)
		}
	} else {
		for _, port := range componentPorts {
			if portNumber == port {
				return portNumber, nil
			}
		}
	}

	return portNumber, fmt.Errorf("given port %d is not exposed on given component, available ports are: %s", portNumber, strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
}

// GetURLName returns a url name from the component name and the given port number
func GetURLName(componentName string, componentPort int) string {
	if componentPort == -1 {
		return componentName
	}
	return fmt.Sprintf("%v-%v", componentName, componentPort)
}
