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

// ComponentNotExistError is returned when a component referenced in a command or component does not exist
type ComponentNotExistError struct {
	name string
}

func NewComponentNotExistError(name string) ComponentNotExistError {
	return ComponentNotExistError{
		name: name,
	}
}

func (e ComponentNotExistError) Error() string {
	return fmt.Sprintf("component %q does not exists", e.name)
}

type ComponentsWithSameNameError struct {
	name string
}

func NewComponentsWithSameNameError(name string) ComponentsWithSameNameError {
	return ComponentsWithSameNameError{
		name: name,
	}
}

func (e ComponentsWithSameNameError) Error() string {
	return fmt.Sprintf("more than one component with the same name %q, should not happen", e.name)
}

// ComponentTypeNotFoundError is returned when no component with the specified type has been found in Devfile
type ComponentTypeNotFoundError struct {
	componentType v1alpha2.ComponentType
}

func NewComponentTypeNotFoundError(componentType v1alpha2.ComponentType) ComponentTypeNotFoundError {
	return ComponentTypeNotFoundError{
		componentType: componentType,
	}
}

func (e ComponentTypeNotFoundError) Error() string {
	return fmt.Sprintf("no component with type %q found in Devfile", e.componentType)
}
