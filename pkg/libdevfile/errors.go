package libdevfile

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

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

type MalformedCommandError struct {
	typ  v1alpha2.CommandType
	name string
}

func NewMalformedCommandError(typ v1alpha2.CommandType, name string) MalformedCommandError {
	return MalformedCommandError{
		typ:  typ,
		name: name,
	}
}

func (e MalformedCommandError) Error() string {
	return fmt.Sprintf("%s command %q is malformed", e.typ, e.name)
}

type MalformedComponentError struct {
	typ  v1alpha2.ComponentType
	name string
}

func NewMalformedComponentError(typ v1alpha2.ComponentType, name string) MalformedComponentError {
	return MalformedComponentError{
		typ:  typ,
		name: name,
	}
}

func (e MalformedComponentError) Error() string {
	return fmt.Sprintf("%s component %q is malformed", e.typ, e.name)
}
