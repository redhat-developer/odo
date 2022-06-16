package common

import (
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

const (
	// EnvProjectsRoot is the env defined for project mount in a component container when component's mountSources=true
	EnvProjectsRoot = "PROJECTS_ROOT"

	// EnvDebugPort is the env defined in the runtime component container which holds the debug port for remote debugging
	EnvDebugPort = "DEBUG_PORT"
)

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
