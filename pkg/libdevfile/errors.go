package libdevfile

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// NoCommandFoundError is returned when no command of the specified kind is found in devfile
type NoCommandFoundError struct {
	kind v1alpha2.CommandGroupKind
}

func NewNoCommandFoundError(kind v1alpha2.CommandGroupKind) NoCommandFoundError {
	return NoCommandFoundError{
		kind: kind,
	}
}
func (e NoCommandFoundError) Error() string {
	return fmt.Sprintf("no %s command found in devfile", e.kind)
}

// NoDefaultCommandFoundError is returned when several commands of the specified kind exist
// but no one is the default one
type NoDefaultCommandFoundError struct {
	kind v1alpha2.CommandGroupKind
}

func NewNoDefaultCommandFoundError(kind v1alpha2.CommandGroupKind) NoDefaultCommandFoundError {
	return NoDefaultCommandFoundError{
		kind: kind,
	}
}
func (e NoDefaultCommandFoundError) Error() string {
	return fmt.Sprintf("no default %s command found in devfile", e.kind)
}

// MoreThanOneDefaultCommandFoundError is returned when several default commands of the specified kind exist
type MoreThanOneDefaultCommandFoundError struct {
	kind v1alpha2.CommandGroupKind
}

func NewMoreThanOneDefaultCommandFoundError(kind v1alpha2.CommandGroupKind) MoreThanOneDefaultCommandFoundError {
	return MoreThanOneDefaultCommandFoundError{
		kind: kind,
	}
}
func (e MoreThanOneDefaultCommandFoundError) Error() string {
	return fmt.Sprintf("more than one default %s command found in devfile, this should not happen", e.kind)
}
