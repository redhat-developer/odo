package component

import (
	"fmt"
)

// NoCommandFoundError is returned when no command of the specified kind is found in devfile
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
	return fmt.Sprintf("no component found with name %q in the namespace %q", e.name, e.namespace)
}
