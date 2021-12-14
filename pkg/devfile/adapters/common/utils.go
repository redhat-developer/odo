package common

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
)

// PredefinedDevfileCommands encapsulates constants for predefined devfile commands
type PredefinedDevfileCommands string

// DevfileEventType encapsulates constants for devfile events
type DevfileEventType string

const (
	// DefaultDevfileInitCommand is a predefined devfile command for init
	DefaultDevfileInitCommand PredefinedDevfileCommands = "devinit"

	// DefaultDevfileBuildCommand is a predefined devfile command for build
	DefaultDevfileBuildCommand PredefinedDevfileCommands = "devbuild"

	// DefaultDevfileRunCommand is a predefined devfile command for run
	DefaultDevfileRunCommand PredefinedDevfileCommands = "devrun"

	// DefaultDevfileDebugCommand is a predefined devfile command for debug
	DefaultDevfileDebugCommand PredefinedDevfileCommands = "debugrun"

	// SupervisordInitContainerName The init container name for supervisord
	SupervisordInitContainerName = "copy-supervisord"

	// Default Image that will be used containing the supervisord binary and assembly scripts
	// use GetBootstrapperImage() function instead of this variable
	defaultBootstrapperImage = "registry.access.redhat.com/ocp-tools-4/odo-init-container-rhel8:1.1.11"

	// SupervisordControlCommand sub command which stands for control
	SupervisordControlCommand = "ctl"

	// SupervisordVolumeName Create a custom name and (hope) that users don't use the *exact* same name in their deployment (occlient.go)
	SupervisordVolumeName = "odo-supervisord-shared-data"

	// SupervisordMountPath The supervisord Mount Path for the container mounting the supervisord volume
	SupervisordMountPath = "/opt/odo/"

	// SupervisordBinaryPath The supervisord binary path inside the container volume mount
	SupervisordBinaryPath = "/opt/odo/bin/supervisord"

	// SupervisordConfFile The supervisord configuration file inside the container volume mount
	SupervisordConfFile = "/opt/odo/conf/devfile-supervisor.conf"

	// OdoInitImageContents The path to the odo init image contents
	OdoInitImageContents = "/opt/odo-init/."

	// ENV variable to overwrite image used to bootstrap SupervisorD in S2I and Devfile builder Image
	bootstrapperImageEnvName = "ODO_BOOTSTRAPPER_IMAGE"

	// BinBash The path to sh executable
	BinBash = "/bin/sh"

	// DefaultVolumeSize Default volume size for volumes defined in a devfile
	DefaultVolumeSize = "1Gi"

	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"

	// EnvProjectsSrc is the env defined for path to the project source in a component container
	EnvProjectsSrc = "PROJECT_SOURCE"

	// EnvOdoCommandRunWorkingDir is the env defined in the runtime component container which holds the work dir for the run command
	EnvOdoCommandRunWorkingDir = "ODO_COMMAND_RUN_WORKING_DIR"

	// EnvOdoCommandRun is the env defined in the runtime component container which holds the run command to be executed
	EnvOdoCommandRun = "ODO_COMMAND_RUN"

	// EnvOdoCommandDebugWorkingDir is the env defined in the runtime component container which holds the work dir for the debug command
	EnvOdoCommandDebugWorkingDir = "ODO_COMMAND_DEBUG_WORKING_DIR"

	// EnvOdoCommandDebug is the env defined in the runtime component container which holds the debug command to be executed
	EnvOdoCommandDebug = "ODO_COMMAND_DEBUG"

	// EnvDebugPort is the env defined in the runtime component container which holds the debug port for remote debugging
	EnvDebugPort = "DEBUG_PORT"

	// ShellExecutable is the shell executable
	ShellExecutable = "/bin/sh"

	// SupervisordCtlSubCommand is the supervisord sub command ctl
	SupervisordCtlSubCommand = "ctl"

	// PreStart is a devfile event
	PreStart DevfileEventType = "preStart"

	// PostStart is a devfile event
	PostStart DevfileEventType = "postStart"

	// PreStop is a devfile event
	PreStop DevfileEventType = "preStop"

	// PostStop is a devfile event
	PostStop DevfileEventType = "postStop"
)

// CommandNames is a struct to store the default and adapter names for devfile commands
type CommandNames struct {
	DefaultName string
	AdapterName string
}

// GetBootstrapperImage returns the odo-init bootstrapper image
func GetBootstrapperImage() string {
	if env, ok := os.LookupEnv(bootstrapperImageEnvName); ok {
		return env
	}
	return defaultBootstrapperImage
}

// getCommandsByGroup gets commands by the group kind
func getCommandsByGroup(commands []devfilev1.Command, groupType devfilev1.CommandGroupKind) []devfilev1.Command {
	var filteredCommands []devfilev1.Command
	for _, command := range commands {
		commandGroup := parsercommon.GetGroup(command)
		if commandGroup != nil && commandGroup.Kind == groupType {
			filteredCommands = append(filteredCommands, command)
		}
	}

	return filteredCommands
}

// IsRestartRequired checks if restart required for run command
func IsRestartRequired(hotReload bool, runModeChanged bool) bool {
	if runModeChanged || !hotReload {
		return true
	}

	return false
}

// IsEnvPresent checks if the env variable is present in an array of env variables
func IsEnvPresent(envVars []devfilev1.EnvVar, envVarName string) bool {
	for _, envVar := range envVars {
		if envVar.Name == envVarName {
			return true
		}
	}

	return false
}

// IsPortPresent checks if the port is present in the endpoints array
func IsPortPresent(endpoints []devfilev1.Endpoint, port int) bool {
	for _, endpoint := range endpoints {
		if endpoint.TargetPort == port {
			return true
		}
	}

	return false
}

// GetComponentEnvVar returns true if a list of env vars contains the specified env var
// If the env exists, it returns the value of it
func GetComponentEnvVar(env string, envs []devfilev1.EnvVar) string {
	for _, envVar := range envs {
		if envVar.Name == env {
			return envVar.Value
		}
	}
	return ""
}

// GetCommandsFromEvent returns the list of commands from the event name.
// If the event is a composite command, it returns the sub-commands from the tree
func GetCommandsFromEvent(commandsMap map[string]devfilev1.Command, eventName string) []string {
	var commands []string

	if command, ok := commandsMap[eventName]; ok {
		if command.Composite != nil {
			klog.V(4).Infof("%s is a composite command", command.Id)
			for _, compositeSubCmd := range command.Composite.Commands {
				klog.V(4).Infof("checking if sub-command %s is either an exec or a composite command ", compositeSubCmd)
				subCommands := GetCommandsFromEvent(commandsMap, strings.ToLower(compositeSubCmd))
				commands = append(commands, subCommands...)
			}
		} else {
			klog.V(4).Infof("%s is an exec command", command.Id)
			commands = append(commands, command.Id)
		}
	}

	return commands
}

// GetCommandsMap returns a map of the command Id to the command
func GetCommandsMap(commands []devfilev1.Command) map[string]devfilev1.Command {
	commandMap := make(map[string]devfilev1.Command, len(commands))
	for _, command := range commands {
		command.Id = strings.ToLower(command.Id)
		commandMap[command.Id] = command
	}
	return commandMap
}

// GetSyncFilesFromAttributes gets the target files and folders along with their respective remote destination from the devfile
// it uses the "dev.odo.push.path" attribute in the run command
func GetSyncFilesFromAttributes(commandsMap PushCommandsMap) map[string]string {
	syncMap := make(map[string]string)
	if value, ok := commandsMap[devfilev1.RunCommandGroupKind]; ok {
		for key, value := range value.Attributes.Strings(nil) {
			if strings.HasPrefix(key, "dev.odo.push.path:") {
				localValue := strings.ReplaceAll(key, "dev.odo.push.path:", "")
				syncMap[filepath.Clean(localValue)] = filepath.ToSlash(filepath.Clean(value))
			}
		}
	}
	return syncMap
}

// RemoveDevfileURIContents removes contents
// which are used via a URI in the devfile
func RemoveDevfileURIContents(devfile devfileParser.DevfileObj, componentContext string) error {
	return removeDevfileURIContents(devfile, componentContext, devfilefs.DefaultFs{})
}

func removeDevfileURIContents(devfile devfileParser.DevfileObj, componentContext string, fs devfilefs.Filesystem) error {
	components, err := devfile.Data.GetComponents(parsercommon.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		var uri string
		if component.Kubernetes != nil && component.Kubernetes.Uri != "" {
			uri = component.Kubernetes.Uri
		}

		if component.Openshift != nil && component.Openshift.Uri != "" {
			uri = component.Openshift.Uri
		}

		if uri == "" {
			continue
		}

		parsedURL, err := url.Parse(uri)
		if err != nil {
			continue
		}
		if len(parsedURL.Host) != 0 && len(parsedURL.Scheme) != 0 {
			continue
		}

		completePath := filepath.Join(componentContext, uri)
		err = fs.Remove(completePath)
		if err != nil {
			return err
		}
	}
	return nil
}
