package libdevfile

import (
	"fmt"
	"reflect"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

type Handler interface {
	ApplyImage(image v1alpha2.Component) error
	ApplyKubernetes(kubernetes v1alpha2.Component) error
	Execute(command v1alpha2.Command) error
}

// Deploy executes the default Deploy command of the devfile
func Deploy(devfileObj parser.DevfileObj, handler Handler) error {
	deployCommand, err := getDefaultCommand(devfileObj, v1alpha2.DeployCommandGroupKind)
	if err != nil {
		return err
	}

	return executeCommand(devfileObj, deployCommand, handler)
}

// getDefaultCommand returns the default command of the given kind in the devfile.
// If only one command of the kind exists, it is returned, even if it is not marked as default
func getDefaultCommand(devfileObj parser.DevfileObj, kind v1alpha2.CommandGroupKind) (v1alpha2.Command, error) {
	groupCmds, err := devfileObj.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: kind,
		},
	})
	if err != nil {
		return v1alpha2.Command{}, err
	}
	if len(groupCmds) == 0 {
		return v1alpha2.Command{}, NewNoCommandFoundError(kind)
	}
	if len(groupCmds) > 1 {
		var found bool
		var foundGroupCmd v1alpha2.Command
		for _, groupCmd := range groupCmds {
			group := common.GetGroup(groupCmd)
			if group == nil {
				continue
			}
			if group.IsDefault != nil && *group.IsDefault {
				if found {
					return v1alpha2.Command{}, NewMoreThanOneDefaultCommandFoundError(kind)
				}
				found = true
				foundGroupCmd = groupCmd
			}
		}
		if !found {
			return v1alpha2.Command{}, NewNoDefaultCommandFoundError(kind)
		}
		return foundGroupCmd, nil
	}
	return groupCmds[0], nil
}

// executeCommand executes a specific command of a devfile using handler as backend
func executeCommand(devfileObj parser.DevfileObj, command v1alpha2.Command, handler Handler) error {
	cmd, err := newCommand(devfileObj, command)
	if err != nil {
		return err
	}
	return cmd.Execute(handler)
}

func HasPostStartEvents(devfileObj parser.DevfileObj) bool {
	postStartEvents := devfileObj.Data.GetEvents().PostStart
	return len(postStartEvents) > 0
}

func HasPreStopEvents(devfileObj parser.DevfileObj) bool {
	preStopEvents := devfileObj.Data.GetEvents().PreStop
	return len(preStopEvents) > 0
}

func ExecPostStartEvents(devfileObj parser.DevfileObj, componentName string, handler Handler) error {
	postStartEvents := devfileObj.Data.GetEvents().PostStart
	return execDevfileEvent(devfileObj, postStartEvents, handler)
}

func ExecPreStopEvents(devfileObj parser.DevfileObj, componentName string, handler Handler) error {
	preStopEvents := devfileObj.Data.GetEvents().PreStop
	return execDevfileEvent(devfileObj, preStopEvents, handler)
}

// execDevfileEvent receives a Devfile Event (PostStart, PreStop etc.) and loops through them
// Each Devfile Command associated with the given event is retrieved, and executed in the container specified
// in the command
func execDevfileEvent(devfileObj parser.DevfileObj, events []string, handler Handler) error {
	if len(events) > 0 {
		commandMap, err := allCommandsMap(devfileObj)
		if err != nil {
			return err
		}
		for _, commandName := range events {
			command, ok := commandMap[commandName]
			if !ok {
				return fmt.Errorf("unable to find devfile command %q", commandName)
			}

			c, err := newCommand(devfileObj, command)
			if err != nil {
				return err
			}
			// Execute command in container
			err = c.Execute(handler)
			if err != nil {
				return fmt.Errorf("unable to execute devfile command %q: %w", commandName, err)
			}
		}
	}
	return nil
}

// GetContainerComponents returns a slice of container components in the given devfile
func GetContainerComponents(devfileObj parser.DevfileObj) ([]v1alpha2.Component, error) {
	return devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
}

// GetContainerEndpointMapping returns a map of container names and slice of its endpoints (in int) with exposure status other than none
func GetContainerEndpointMapping(containers []v1alpha2.Component) map[string][]int {
	ceMapping := make(map[string][]int)
	for _, container := range containers {
		if container.ComponentUnion.Container == nil {
			// this is not a container component; continue prevents panic when accessing Endpoints field
			continue
		}
		k := container.Name
		if _, ok := ceMapping[k]; !ok {
			ceMapping[k] = []int{}
		}

		endpoints := container.Container.Endpoints
		for _, e := range endpoints {
			if e.Exposure != v1alpha2.NoneEndpointExposure {
				ceMapping[k] = append(ceMapping[k], e.TargetPort)
			}
		}
	}
	return ceMapping
}

// GetPublicAndInternalEndpointsFromDevfile returns a slice of all endpoints in a devfile with exposure value not set to public or internal
func GetPublicAndInternalEndpointsFromDevfile(devfileObj parser.DevfileObj) ([]v1alpha2.Endpoint, error) {
	containers, err := GetContainerComponents(devfileObj)
	if err != nil {
		return nil, err
	}

	var endpoints []v1alpha2.Endpoint
	for _, c := range containers {
		for _, e := range c.Container.Endpoints {
			if e.Exposure == v1alpha2.PublicEndpointExposure || e.Exposure == v1alpha2.InternalEndpointExposure {
				endpoints = append(endpoints, e)
			}
		}

	}
	return endpoints, nil
}

// HaveEndpointsChanged returns true if the total number of public and/or internal endpoints have changed between
// the devfile objects
func HaveEndpointsChanged(oldDevfile, newDevfile parser.DevfileObj) (bool, error) {
	oldEndpoints, err := GetPublicAndInternalEndpointsFromDevfile(oldDevfile)
	if err != nil {
		return false, err
	}

	newEndpoints, err := GetPublicAndInternalEndpointsFromDevfile(newDevfile)
	if err != nil {
		return false, err
	}

	return !reflect.DeepEqual(oldEndpoints, newEndpoints), nil
}
