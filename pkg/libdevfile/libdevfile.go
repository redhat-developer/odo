package libdevfile

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/util"
)

const DebugEndpointNamePrefix = "debug"

type Handler interface {
	ApplyImage(image v1alpha2.Component) error
	ApplyKubernetes(kubernetes v1alpha2.Component) error
	Execute(command v1alpha2.Command) error
}

// Deploy executes the default deploy command of the devfile.
func Deploy(devfileObj parser.DevfileObj, handler Handler) error {
	return ExecuteCommandByNameAndKind(devfileObj, "", v1alpha2.DeployCommandGroupKind, handler, false)
}

// Build executes the default Build command of the devfile.
// If buildCmd is empty, this looks for the default Build command in the Devfile. No error is returned and no operation is performed
// if the default command could not be found.
// An error is returned if buildCmd is not empty and has no corresponding command in the Devfile.
func Build(devfileObj parser.DevfileObj, buildCmd string, handler Handler) error {
	return ExecuteCommandByNameAndKind(devfileObj, buildCmd, v1alpha2.BuildCommandGroupKind, handler, buildCmd == "")
}

// ExecuteCommandByNameAndKind executes the specified command cmdName of the given kind in the Devfile.
// If cmdName is empty, it executes the default command for the given kind or returns an error if there is no default command.
// If ignoreCommandNotFound is true, nothing is executed if the command is not found and no error is returned.
func ExecuteCommandByNameAndKind(
	devfileObj parser.DevfileObj,
	cmdName string,
	kind v1alpha2.CommandGroupKind,
	handler Handler,
	ignoreCommandNotFound bool,
) error {
	cmd, hasDefaultCmd, err := GetCommand(devfileObj, cmdName, kind)
	if err != nil {
		if _, isNotFound := err.(NoCommandFoundError); isNotFound {
			if ignoreCommandNotFound {
				klog.V(3).Infof("ignoring command not found: %v", cmdName)
				return nil
			}
		}
		return err
	}
	if !hasDefaultCmd {
		if ignoreCommandNotFound {
			klog.V(3).Infof("ignoring default %v command not found", kind)
			return nil
		}
		return NewNoDefaultCommandFoundError(kind)
	}

	return executeCommand(devfileObj, cmd, handler)
}

// executeCommand executes a specific command of a devfile using handler as backend
func executeCommand(devfileObj parser.DevfileObj, command v1alpha2.Command, handler Handler) error {
	cmd, err := newCommand(devfileObj, command)
	if err != nil {
		return err
	}
	return cmd.Execute(handler)
}

// GetCommand iterates through the devfile commands and returns the devfile command with the specified name and group kind.
// If commandName is empty, it returns the default command for the group kind or returns an error if there is no default command.
func GetCommand(
	devfileObj parser.DevfileObj,
	commandName string,
	groupType v1alpha2.CommandGroupKind,
) (v1alpha2.Command, bool, error) {
	if commandName == "" {
		return getDefaultCommand(devfileObj, groupType)
	}
	cmdByName, err := getCommandByName(devfileObj, groupType, commandName)
	if err != nil {
		return v1alpha2.Command{}, false, err
	}
	return cmdByName, true, nil
}

// getDefaultCommand iterates through the devfile commands and returns the default command associated with the group kind.
// If there is no default command, the second return value is false.
func getDefaultCommand(devfileObj parser.DevfileObj, groupType v1alpha2.CommandGroupKind) (v1alpha2.Command, bool, error) {
	commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{CommandOptions: common.CommandOptions{CommandGroupKind: groupType}})
	if err != nil {
		return v1alpha2.Command{}, false, err
	}

	// if there is only one command of a given group kind, use it as default
	if len(commands) == 1 {
		return commands[0], true, nil
	}

	defaultCmds := make([]v1alpha2.Command, 0)

	for _, cmd := range commands {
		cmdGroup := common.GetGroup(cmd)
		if cmdGroup != nil {
			if cmdGroup.IsDefault != nil && *cmdGroup.IsDefault {
				defaultCmds = append(defaultCmds, cmd)
			}
		} else {
			klog.V(2).Infof("command %s has no group", cmd.Id)
		}
	}

	if len(defaultCmds) == 0 {
		return v1alpha2.Command{}, false, nil
	}
	if len(defaultCmds) > 1 {
		return v1alpha2.Command{}, false, NewMoreThanOneDefaultCommandFoundError(groupType)
	}
	return defaultCmds[0], true, nil
}

// getCommandByName iterates through the devfile commands and returns the command with the specified name and group.
// It returns an error if no command was found.
func getCommandByName(devfileObj parser.DevfileObj, groupType v1alpha2.CommandGroupKind, commandName string) (v1alpha2.Command, error) {
	commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{CommandOptions: common.CommandOptions{CommandGroupKind: groupType}})
	if err != nil {
		return v1alpha2.Command{}, err
	}

	for _, cmd := range commands {
		if cmd.Id == commandName {
			return cmd, nil
		}
	}

	return v1alpha2.Command{}, NewNoCommandFoundError(groupType, commandName)
}

// ValidateAndGetCommand validates and returns the command specified if it is valid.
// It works just like GetCommand, except that it returns an error if it could not find the command.
//
// If commandName is empty, it looks up the default command for the given kind.
//
// A command is "valid" here if it was found given its name (if commandName is not empty),
// or (for a default command), if there is no other default command for the same kind.
func ValidateAndGetCommand(devfileObj parser.DevfileObj, commandName string, groupType v1alpha2.CommandGroupKind) (v1alpha2.Command, error) {
	cmd, ok, err := GetCommand(devfileObj, commandName, groupType)
	if err != nil {
		return v1alpha2.Command{}, err
	}
	if !ok {
		return v1alpha2.Command{}, NewNoCommandFoundError(groupType, commandName)
	}
	return cmd, nil
}

// ValidateAndGetPushCommands validates the build and the run commands, if provided through odo dev or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if validated successfully, or an error otherwise.
func ValidateAndGetPushCommands(
	devfileObj parser.DevfileObj,
	devfileBuildCmd,
	devfileRunCmd string,
) (map[v1alpha2.CommandGroupKind]v1alpha2.Command, error) {
	var buildCmd v1alpha2.Command
	var present bool
	var err error

	if devfileBuildCmd != "" {
		buildCmd, err = ValidateAndGetCommand(devfileObj, devfileBuildCmd, v1alpha2.BuildCommandGroupKind)
		present = true
	} else {
		buildCmd, present, err = GetCommand(devfileObj, devfileBuildCmd, v1alpha2.BuildCommandGroupKind)
	}
	if err != nil {
		return nil, err
	}

	commandMap := make(map[v1alpha2.CommandGroupKind]v1alpha2.Command)
	if present {
		klog.V(2).Infof("Build command: %v", buildCmd.Id)
		commandMap[v1alpha2.BuildCommandGroupKind] = buildCmd
	} else {
		// Build command is optional, unless it was explicitly specified by the caller (at which point it would have been validated via ValidateAndGetCommand).
		klog.V(2).Infof("No build command was provided")
	}

	var runCmd v1alpha2.Command
	runCmd, err = ValidateAndGetCommand(devfileObj, devfileRunCmd, v1alpha2.RunCommandGroupKind)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("Run command: %v", runCmd.Id)
	commandMap[v1alpha2.RunCommandGroupKind] = runCmd

	return commandMap, nil
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

// GetContainerEndpointMapping returns a map of container names and slice of its endpoints (in int)
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
			ceMapping[k] = append(ceMapping[k], e.TargetPort)
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

// GetDebugEndpointsForComponent returns all Debug endpoints for the specified component.
// It returns an error if the component specified is not a container component.
func GetDebugEndpointsForComponent(cmp v1alpha2.Component) ([]v1alpha2.Endpoint, error) {
	if cmp.Container == nil {
		return nil, fmt.Errorf("component %q is not a container component", cmp.Name)
	}

	var result []v1alpha2.Endpoint
	for _, ep := range cmp.Container.Endpoints {
		if IsDebugEndpoint(ep) {
			result = append(result, ep)
		}
	}
	return result, nil
}

// IsDebugEndpoint returns whether the specified endpoint represents a Debug endpoint,
// based on the following naming convention: it is considered a Debug endpoint if it's named "debug" or if its name starts with "debug-".
func IsDebugEndpoint(ep v1alpha2.Endpoint) bool {
	return IsDebugPort(ep.Name)
}

// IsDebugPort returns whether the specified string represents a Debug endpoint,
// based on the following naming convention: it is considered a Debug endpoint if it's named "debug" or if its name starts with "debug-".
func IsDebugPort(name string) bool {
	return name == DebugEndpointNamePrefix || strings.HasPrefix(name, DebugEndpointNamePrefix+"-")
}

// GetContainerComponentsForCommand returns the list of container components that would get used if the specified command runs.
func GetContainerComponentsForCommand(devfileObj parser.DevfileObj, cmd v1alpha2.Command) ([]string, error) {
	// No error if cmd is empty
	if reflect.DeepEqual(cmd, v1alpha2.Command{}) {
		return nil, nil
	}

	commandType, err := common.GetCommandType(cmd)
	if err != nil {
		return nil, err
	}

	hasComponent := func(n string) bool {
		_, ok, _ := findComponentByNameAndType(devfileObj, n, v1alpha2.ContainerComponentType)
		return ok
	}

	switch commandType {
	case v1alpha2.ExecCommandType:
		if hasComponent(cmd.Exec.Component) {
			return []string{cmd.Exec.Component}, nil
		}
		return nil, nil
	case v1alpha2.ApplyCommandType:
		if hasComponent(cmd.Apply.Component) {
			return []string{cmd.Apply.Component}, nil
		}
		return nil, nil
	case v1alpha2.CompositeCommandType:
		var commandsMap map[string]v1alpha2.Command
		commandsMap, err = allCommandsMap(devfileObj)
		if err != nil {
			return nil, err
		}

		var res []string
		set := make(map[string]bool)
		var componentsForCommand []string
		for _, c := range cmd.Composite.Commands {
			fromCommandMap, present := commandsMap[strings.ToLower(c)]
			if !present {
				return nil, fmt.Errorf("command %q not found in all commands map", c)
			}
			componentsForCommand, err = GetContainerComponentsForCommand(devfileObj, fromCommandMap)
			if err != nil {
				return nil, err
			}
			for _, s := range componentsForCommand {
				if _, ok := set[s]; !ok && hasComponent(s) {
					set[s] = true
					res = append(res, s)
				}
			}
		}

		return res, nil

	default:
		return nil, fmt.Errorf("type not handled for command %q: %v", cmd.Id, commandType)
	}
}

// GetK8sManifestsWithVariablesSubstituted returns the full content of either a Kubernetes or an Openshift
// Devfile component, either Inlined or referenced via a URI.
// No matter how the component is defined, it returns the content with all variables substituted
// using the global variables map defined in `devfileObj`.
// An error is returned if the content references an invalid variable key not defined in the Devfile object.
func GetK8sManifestsWithVariablesSubstituted(devfileObj parser.DevfileObj, devfileCmpName string,
	context string, fs devfilefs.Filesystem) (string, error) {

	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{FilterByName: devfileCmpName})
	if err != nil {
		return "", err
	}

	if len(components) == 0 {
		return "", NewComponentNotExistError(devfileCmpName)
	}

	if len(components) != 1 {
		return "", NewComponentsWithSameNameError(devfileCmpName)
	}

	devfileCmp := components[0]
	componentType, err := common.GetComponentType(devfileCmp)
	if err != nil {
		return "", err
	}

	var content, uri string
	switch componentType {
	case v1alpha2.KubernetesComponentType:
		content = devfileCmp.Kubernetes.Inlined
		if devfileCmp.Kubernetes.Uri != "" {
			uri = devfileCmp.Kubernetes.Uri
		}

	case v1alpha2.OpenshiftComponentType:
		content = devfileCmp.Openshift.Inlined
		if devfileCmp.Openshift.Uri != "" {
			uri = devfileCmp.Openshift.Uri
		}

	default:
		return "", fmt.Errorf("unexpected component type %s", componentType)
	}

	if uri != "" {
		return loadResourceManifestFromUriAndResolveVariables(devfileObj, uri, context, fs)
	}
	return substituteVariables(devfileObj.Data.GetDevfileWorkspaceSpec().Variables, content)
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
// Inspired from variables.validateAndReplaceDataWithVariable, which is unfortunately not exported
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

// findComponentByNameAndType returns the Devfile component that matches the specified name and type.
func findComponentByNameAndType(d parser.DevfileObj, n string, t v1alpha2.ComponentType) (v1alpha2.Component, bool, error) {
	comps, err := d.Data.GetComponents(common.DevfileOptions{ComponentOptions: common.ComponentOptions{ComponentType: t}})
	if err != nil {
		return v1alpha2.Component{}, false, err
	}
	for _, c := range comps {
		if c.Name == n {
			return c, true, nil
		}
	}
	return v1alpha2.Component{}, false, nil
}
