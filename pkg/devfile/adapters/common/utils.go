package common

import (
	"os"
	"strings"

	"k8s.io/klog"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
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
	defaultBootstrapperImage = "registry.access.redhat.com/ocp-tools-4/odo-init-container-rhel8:1.1.9"

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
func getCommandsByGroup(data data.DevfileData, groupType devfilev1.CommandGroupKind) []devfilev1.Command {
	var commands []devfilev1.Command

	for _, command := range data.GetCommands() {
		commandGroup := parsercommon.GetGroup(command)
		if commandGroup != nil && commandGroup.Kind == groupType {
			commands = append(commands, command)
		}
	}

	return commands
}

// GetVolumeMountPath gets the volume mount's path
func GetVolumeMountPath(volumeMount devfilev1.VolumeMount) string {
	// if there is no volume mount path, default to volume mount name as per devfile schema
	if volumeMount.Path == "" {
		volumeMount.Path = "/" + volumeMount.Name
	}

	return volumeMount.Path
}

// GetVolumes iterates through the components in the devfile and returns a map of container name to the devfile volumes
func GetVolumes(devfileObj devfileParser.DevfileObj) map[string][]DevfileVolume {
	containerComponents := devfileObj.Data.GetDevfileContainerComponents()
	volumeComponents := devfileObj.Data.GetDevfileVolumeComponents()

	// containerNameToVolumes is a map of the Devfile container name to the Devfile container Volumes
	containerNameToVolumes := make(map[string][]DevfileVolume)
	for _, containerComp := range containerComponents {
		for _, volumeMount := range containerComp.Container.VolumeMounts {
			size := DefaultVolumeSize

			// check if there is a volume component name against the container component volume mount name
			for _, volumeComp := range volumeComponents {
				if volumeComp.Name == volumeMount.Name && len(volumeComp.Volume.Size) > 0 {
					// If there is a volume size mentioned in the devfile, use it
					size = volumeComp.Volume.Size
				}
			}

			vol := DevfileVolume{
				Name:          volumeMount.Name,
				ContainerPath: GetVolumeMountPath(volumeMount),
				Size:          size,
			}
			containerNameToVolumes[containerComp.Name] = append(containerNameToVolumes[containerComp.Name], vol)
		}
	}
	return containerNameToVolumes
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
