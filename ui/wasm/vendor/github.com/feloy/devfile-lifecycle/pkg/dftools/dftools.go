package dftools

import (
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

func GetCommandGroup(command v1alpha2.Command) v1alpha2.CommandGroup {
	if command.Exec != nil && command.Exec.Group != nil {
		return *command.Exec.Group
	}
	if command.Apply != nil && command.Apply.Group != nil {
		return *command.Apply.Group
	}
	if command.Composite != nil && command.Composite.Group != nil {
		return *command.Composite.Group
	}
	return v1alpha2.CommandGroup{}
}

func IsDebugPort(endpoint v1alpha2.Endpoint) bool {
	name := endpoint.Name
	return name == "debug" || strings.HasPrefix(name, "debug-")
}
