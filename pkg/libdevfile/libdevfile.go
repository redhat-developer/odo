package libdevfile

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/util"
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

func ExecPostStartEvents(devfileObj parser.DevfileObj, handler Handler) error {
	postStartEvents := devfileObj.Data.GetEvents().PostStart
	return execDevfileEvent(devfileObj, postStartEvents, handler)
}

func ExecPreStopEvents(devfileObj parser.DevfileObj, handler Handler) error {
	preStopEvents := devfileObj.Data.GetEvents().PreStop
	return execDevfileEvent(devfileObj, preStopEvents, handler)
}

func hasCommand(devfileData data.DevfileData, kind v1alpha2.CommandGroupKind) bool {
	commands, err := devfileData.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: kind,
		},
	})
	return err == nil && len(commands) > 0
}

func HasRunCommand(devfileData data.DevfileData) bool {
	return hasCommand(devfileData, v1alpha2.RunCommandGroupKind)
}

func HasDeployCommand(devfileData data.DevfileData) bool {
	return hasCommand(devfileData, v1alpha2.DeployCommandGroupKind)
}

func HasDebugCommand(devfileData data.DevfileData) bool {
	return hasCommand(devfileData, v1alpha2.DebugCommandGroupKind)
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

// GetEndpointsFromDevfile returns a slice of all endpoints in a devfile and ignores the endpoints with exposure values in ignoreExposures
func GetEndpointsFromDevfile(devfileObj parser.DevfileObj, ignoreExposures []v1alpha2.EndpointExposure) ([]v1alpha2.Endpoint, error) {
	containers, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: v1alpha2.ContainerComponentType},
	})
	if err != nil {
		return nil, err
	}

	var allEndpoints []v1alpha2.Endpoint
	for _, c := range containers {
		allEndpoints = append(allEndpoints, c.Container.Endpoints...)
	}

	var endpoints []v1alpha2.Endpoint
	for _, e := range allEndpoints {
		ignore := false
		for _, i := range ignoreExposures {
			if e.Exposure == i {
				ignore = true
			}
		}
		if !ignore {
			endpoints = append(endpoints, e)
		}
	}
	return endpoints, nil
}

// GetComponentResourceManifestContentWithVariablesResolved returns the full content of either a Kubernetes or an Openshift
// Devfile component, either Inlined or referenced via a URI.
// No matter how the component is defined, it returns
// the content with all variables substituted using the global variables map defined in `devfileObj`.
// An error is returned if the content references an invalid variable key not defined in the Devfile object.
func GetComponentResourceManifestContentWithVariablesResolved(devfileObj parser.DevfileObj, devfileCmp interface{},
	context string, fs devfilefs.Filesystem) (string, error) {

	var content, uri string
	switch devfileCmp := devfileCmp.(type) {
	case v1alpha2.Component:
		componentType, err := common.GetComponentType(devfileCmp)
		if err != nil {
			return "", err
		}
		switch componentType {
		case v1alpha2.KubernetesComponentType:
			return GetComponentResourceManifestContentWithVariablesResolved(devfileObj, devfileCmp.Kubernetes, context, fs)

		case v1alpha2.OpenshiftComponentType:
			return GetComponentResourceManifestContentWithVariablesResolved(devfileObj, devfileCmp.Openshift, context, fs)

		default:
			return "", fmt.Errorf("unexpected component type %s", componentType)
		}
	case *v1alpha2.KubernetesComponent:
		content = devfileCmp.Inlined
		if devfileCmp.Uri != "" {
			uri = devfileCmp.Uri
		}

	case *v1alpha2.OpenshiftComponent:
		content = devfileCmp.Inlined
		if devfileCmp.Uri != "" {
			uri = devfileCmp.Uri
		}
	default:
		return "", fmt.Errorf("unexpected type for %v", devfileCmp)
	}

	if uri == "" {
		return substituteVariables(devfileObj.Data.GetDevfileWorkspaceSpec().Variables, content)
	}

	return loadResourceManifestFromUriAndResolveVariables(devfileObj, uri, context, fs)
}

func loadResourceManifestFromUriAndResolveVariables(devfileObj parser.DevfileObj, uri string,
	context string, fs devfilefs.Filesystem) (string, error) {
	content, err := util.GetDataFromURI(uri, context, fs)
	if err != nil {
		return content, err
	}
	return substituteVariables(devfileObj.Data.GetDevfileWorkspaceSpec().Variables, content)
}

// substituteVariables validates the string for a global variable in the given `devfileObj` and replaces it.
// An error is returned if the string references an invalid variable key not defined in the Devfile object.
//
//Inspired from variables.validateAndReplaceDataWithVariable, which is unfortunately not exported
func substituteVariables(devfileVars map[string]string, val string) (string, error) {
	// example of the regex: {{variable}} / {{ variable }}
	matches := regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`).FindAllStringSubmatch(val, -1)
	var invalidKeys []string
	for _, match := range matches {
		varValue, ok := devfileVars[match[1]]
		if !ok {
			invalidKeys = append(invalidKeys, match[1])
		} else {
			val = strings.Replace(val, match[0], varValue, -1)
		}
	}

	if len(invalidKeys) > 0 {
		return val, &variables.InvalidKeysError{Keys: invalidKeys}
	}

	return val, nil
}
