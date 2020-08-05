package common

import "fmt"

// AlreadyExistError error returned if tried to add already exisitng field
type AlreadyExistError struct {
	// field which already exist
	Field string
	// field name
	Name string
}

func (e *AlreadyExistError) Error() string {
	return fmt.Sprintf("%s %s already exists in devfile", e.Field, e.Name)
}
