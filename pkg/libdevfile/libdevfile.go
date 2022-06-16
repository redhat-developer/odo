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

type Handler interface {
	ApplyImage(image v1alpha2.Component) error
	ApplyKubernetes(kubernetes v1alpha2.Component) error
	Execute(command v1alpha2.Command) error
}

// Deploy executes the default Deploy command of the devfile
func Deploy(devfileObj parser.DevfileObj, handler Handler) error {
	return ExecuteCommandByKind(devfileObj, v1alpha2.DeployCommandGroupKind, handler, false)
}

// Build executes the default Build command of the devfile, optionally not failing if the command was not found,
// in case it is optional
func Build(devfileObj parser.DevfileObj, handler Handler, ignoreCommandNotFound bool) error {
	return ExecuteCommandByKind(devfileObj, v1alpha2.BuildCommandGroupKind, handler, ignoreCommandNotFound)
}

// ExecuteCommandByKind executes the default command of the given kind in the Devfile
func ExecuteCommandByKind(devfileObj parser.DevfileObj, kind v1alpha2.CommandGroupKind, handler Handler, ignoreCommandNotFound bool) error {
	cmd, err := GetDefaultCommand(devfileObj, kind)
	if err != nil {
		if ignoreCommandNotFound {
			if _, ok := err.(NoCommandFoundError); ok {
				return nil
			}
		}
		return err
	}

	return executeCommand(devfileObj, cmd, handler)
}

// GetDefaultCommand returns the default command of the given kind in the devfile.
// If only one command of the kind exists, it is returned, even if it is not marked as default
func GetDefaultCommand(devfileObj parser.DevfileObj, kind v1alpha2.CommandGroupKind) (v1alpha2.Command, error) {
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

// ValidateAndGetPushCommands validates the build and the run command,
// if provided through odo dev or else checks the devfile for devBuild and devRun.
// It returns the build and run commands if its validated successfully, error otherwise.
func ValidateAndGetPushCommands(
	data data.DevfileData,
	devfileBuildCmd,
	devfileRunCmd string,
) (commandMap map[v1alpha2.CommandGroupKind]v1alpha2.Command, err error) {
	var emptyCommand v1alpha2.Command
	commandMap = make(map[v1alpha2.CommandGroupKind]v1alpha2.Command)

	isBuildCommandValid, isRunCommandValid := false, false

	buildCommand, buildCmdErr := GetBuildCommand(data, devfileBuildCmd)

	isBuildCmdEmpty := reflect.DeepEqual(emptyCommand, buildCommand)
	if isBuildCmdEmpty && buildCmdErr == nil {
		// If there was no build command specified through odo dev and no default build command in the devfile, default validate to true since the build command is optional
		isBuildCommandValid = true
		klog.V(2).Infof("No build command was provided")
	} else if !reflect.DeepEqual(emptyCommand, buildCommand) && buildCmdErr == nil {
		isBuildCommandValid = true
		commandMap[v1alpha2.BuildCommandGroupKind] = buildCommand
		klog.V(2).Infof("Build command: %v", buildCommand.Id)
	}

	runCommand, runCmdErr := GetRunCommand(data, devfileRunCmd)
	if runCmdErr == nil && !reflect.DeepEqual(emptyCommand, runCommand) {
		isRunCommandValid = true
		commandMap[v1alpha2.RunCommandGroupKind] = runCommand
		klog.V(2).Infof("Run command: %v", runCommand.Id)
	}

	// If either command had a problem, return an empty list of commands and an error
	if !isBuildCommandValid || !isRunCommandValid {
		commandErrors := ""

		if buildCmdErr != nil {
			commandErrors += fmt.Sprintf("\n%s", buildCmdErr.Error())
		}
		if runCmdErr != nil {
			commandErrors += fmt.Sprintf("\n%s", runCmdErr.Error())
		}
		return commandMap, fmt.Errorf(commandErrors)
	}

	return commandMap, nil
}

// ValidateAndGetDebugCommands validates the debug command
func ValidateAndGetDebugCommands(data data.DevfileData, devfileDebugCmd string) (pushDebugCommand v1alpha2.Command, err error) {
	var emptyCommand v1alpha2.Command

	isDebugCommandValid := false
	debugCommand, debugCmdErr := GetDebugCommand(data, devfileDebugCmd)
	if debugCmdErr == nil && !reflect.DeepEqual(emptyCommand, debugCommand) {
		isDebugCommandValid = true
		klog.V(2).Infof("Debug command: %v", debugCommand.Id)
	}

	if !isDebugCommandValid {
		commandErrors := ""
		if debugCmdErr != nil {
			commandErrors += debugCmdErr.Error()
		}
		return v1alpha2.Command{}, fmt.Errorf(commandErrors)
	}

	return debugCommand, nil
}

// ValidateAndGetTestCommands validates the test command
func ValidateAndGetTestCommands(data data.DevfileData, devfileTestCmd string) (testCommand v1alpha2.Command, err error) {
	var emptyCommand v1alpha2.Command
	isTestCommandValid := false
	testCommand, testCmdErr := GetTestCommand(data, devfileTestCmd)
	if testCmdErr == nil && !reflect.DeepEqual(emptyCommand, testCommand) {
		isTestCommandValid = true
		klog.V(2).Infof("Test command: %v", testCommand.Id)
	}

	if !isTestCommandValid && testCmdErr != nil {
		return v1alpha2.Command{}, testCmdErr
	}

	return testCommand, nil
}

// GetBuildCommand iterates through the components in the devfile and returns the build command
func GetBuildCommand(data data.DevfileData, devfileBuildCmd string) (buildCommand v1alpha2.Command, err error) {
	return getCommand(data, devfileBuildCmd, v1alpha2.BuildCommandGroupKind)
}

// GetDebugCommand iterates through the components in the devfile and returns the debug command
func GetDebugCommand(data data.DevfileData, devfileDebugCmd string) (debugCommand v1alpha2.Command, err error) {
	return getCommand(data, devfileDebugCmd, v1alpha2.DebugCommandGroupKind)
}

// GetRunCommand iterates through the components in the devfile and returns the run command
func GetRunCommand(data data.DevfileData, devfileRunCmd string) (runCommand v1alpha2.Command, err error) {
	return getCommand(data, devfileRunCmd, v1alpha2.RunCommandGroupKind)
}

// GetTestCommand iterates through the components in the devfile and returns the test command
func GetTestCommand(data data.DevfileData, devfileTestCmd string) (runCommand v1alpha2.Command, err error) {
	return getCommand(data, devfileTestCmd, v1alpha2.TestCommandGroupKind)
}

// ShouldExecCommandRunOnContainer returns whether the given exec command should run on the specified containerName.
func ShouldExecCommandRunOnContainer(exec *v1alpha2.ExecCommand, containerName string) bool {
	return exec != nil && exec.Component == containerName
}

// getCommand iterates through the devfile commands and returns the devfile command associated with the group
// commands mentioned via the flags are passed via commandName, empty otherwise
func getCommand(data data.DevfileData, commandName string, groupType v1alpha2.CommandGroupKind) (supportedCommand v1alpha2.Command, err error) {

	var command v1alpha2.Command

	if commandName == "" {
		command, err = getCommandAssociatedToGroup(data, groupType)
	} else if commandName != "" {
		command, err = getCommandByName(data, groupType, commandName)
	}

	return command, err
}

// getCommandAssociatedToGroup iterates through the devfile commands and returns the command associated with the group
func getCommandAssociatedToGroup(data data.DevfileData, groupType v1alpha2.CommandGroupKind) (v1alpha2.Command, error) {
	commands, err := data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return v1alpha2.Command{}, err
	}
	var onlyCommand v1alpha2.Command

	for _, cmd := range commands {
		cmdGroup := common.GetGroup(cmd)
		if cmdGroup != nil && cmdGroup.Kind == groupType {
			if util.SafeGetBool(cmdGroup.IsDefault) {
				return cmd, nil
			}
			if reflect.DeepEqual(onlyCommand, v1alpha2.Command{}) {
				// return the only remaining command for the group if there is no default command
				// NOTE: we return outside the for loop since the next iteration can have a default command
				onlyCommand = cmd
			}
		}
	}

	// if default command is not found return the first command found for the matching type.
	if !reflect.DeepEqual(onlyCommand, v1alpha2.Command{}) {
		return onlyCommand, nil
	}

	notFoundError := NoCommandForGroup{Group: groupType}
	// if run command or test command is not found in devfile then it is an error
	if groupType == v1alpha2.RunCommandGroupKind || groupType == v1alpha2.TestCommandGroupKind {
		return onlyCommand, notFoundError
	}

	klog.V(2).Info(notFoundError)
	return onlyCommand, nil
}

// getCommandByName iterates through the devfile commands and returns the command specified associated with the group
func getCommandByName(data data.DevfileData, groupType v1alpha2.CommandGroupKind, commandName string) (v1alpha2.Command, error) {
	commands, err := data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return v1alpha2.Command{}, err
	}

	for _, cmd := range commands {
		if cmd.Id == commandName {

			// Update Group only custom commands (specified by odo flags)
			cmd = updateCommandGroupIfNeeded(groupType, cmd)

			// we have found the command with name, its groupType Should match to the flag
			// e.g --build-command "mybuild"
			// exec:
			//   id: mybuild
			//   group:
			//     kind: build
			cmdGroup := common.GetGroup(cmd)
			if cmdGroup != nil && cmdGroup.Kind != groupType {
				return cmd, fmt.Errorf("command group mismatched, command %s is of group %v in devfile.yaml", commandName, cmd.Exec.Group.Kind)
			}

			return cmd, nil
		}
	}

	// if any command specified via flag is not found in devfile then it is an error.
	return v1alpha2.Command{}, fmt.Errorf("the command \"%v\" is not found in the devfile", commandName)
}

// updateCommandGroupIfNeeded updates the Group of the command specified if it is an Exec command with no Group.
func updateCommandGroupIfNeeded(groupType v1alpha2.CommandGroupKind, command v1alpha2.Command) v1alpha2.Command {
	// Update Group only for exec commands
	// Update Group only when Group is not nil, devfile v2 might contain group for custom commands.
	if command.Exec != nil && command.Exec.Group == nil {
		command.Exec.Group = &v1alpha2.CommandGroup{Kind: groupType}
		return command
	}
	return command
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

// GetK8sManifestWithVariablesSubstituted returns the full content of either a Kubernetes or an Openshift
// Devfile component, either Inlined or referenced via a URI.
// No matter how the component is defined, it returns the content with all variables substituted
// using the global variables map defined in `devfileObj`.
// An error is returned if the content references an invalid variable key not defined in the Devfile object.
func GetK8sManifestWithVariablesSubstituted(devfileObj parser.DevfileObj, devfileCmpName string,
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
