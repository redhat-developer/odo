package common

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// NoCommandForGroup indicates a error when no command was found for the given Group
type NoCommandForGroup struct {
	Group v1alpha2.CommandGroupKind
}

func (n NoCommandForGroup) Error() string {
	return fmt.Sprintf("the command group of kind \"%v\" is not found in the devfile", n.Group)
}
