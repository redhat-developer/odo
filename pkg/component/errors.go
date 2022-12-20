package component

import (
	"fmt"
)

// NoComponentFoundError is returned when no component of the specified name was found.
type NoComponentFoundError struct {
	name      string
	namespace string
}

func NewNoComponentFoundError(name string, namespace string) NoComponentFoundError {
	return NoComponentFoundError{
		name:      name,
		namespace: namespace,
	}
}
func (e NoComponentFoundError) Error() string {
	if e.namespace != "" {
		return fmt.Sprintf("no component found with name %q in the namespace %q", e.name, e.namespace)
	}
	return fmt.Sprintf("no component found with name %q", e.name)
}
