package common

import "fmt"

// AlreadyExistError error returned if tried to add already existing field
type AlreadyExistError struct {
	// field which already exist
	Field string
	// field name
	Name string
}

func (e *AlreadyExistError) Error() string {
	return fmt.Sprintf("%s %s already exists in the devfile", e.Field, e.Name)
}

// NotFoundError error returned if the field with the name is not found
type NotFoundError struct {
	// field which doesn't exist
	Field string
	// field name
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s is not found in the devfile", e.Field, e.Name)
}
