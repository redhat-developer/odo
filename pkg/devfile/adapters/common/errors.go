package common

import (
	"fmt"

	"github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
)

// NoDefaultForGroup indicates a error when no default command was found for the given Group
type NoDefaultForGroup struct {
	Group v1alpha2.CommandGroupKind
}

func (n NoDefaultForGroup) Error() string {
	return fmt.Sprintf("there should be exactly one default command for command group %v, currently there is no default command", n.Group)
}

// MoreDefaultForGroup indicates a error when more than one default command was found for the given Group
type MoreDefaultForGroup struct {
	Group v1alpha2.CommandGroupKind
}

func (m MoreDefaultForGroup) Error() string {
	return fmt.Sprintf("there should be exactly one default command for command group %v, currently there is more than one default command", m.Group)
}

// NoCommandForGroup indicates a error when no command was found for the given Group
type NoCommandForGroup struct {
	Group v1alpha2.CommandGroupKind
}

func (n NoCommandForGroup) Error() string {
	return fmt.Sprintf("the command group of kind \"%v\" is not found in the devfile", n.Group)
}
