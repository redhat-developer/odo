package common

import "fmt"

// FieldAlreadyExistError error returned if tried to add already exisitng field
type FieldAlreadyExistError struct {
	// field which already exist
	Field string
	// field name
	Name string
}

func (e *FieldAlreadyExistError) Error() string {
	return fmt.Sprintf("%s %s already exists in devfile", e.Field, e.Name)
}

// FieldNotFoundError error returned if the field with the name is not found
type FieldNotFoundError struct {
	// field which doesn't exist
	Field string
	// field name
	Name string
}

func (e *FieldNotFoundError) Error() string {
	return fmt.Sprintf("%s %s is not found in the devfile", e.Field, e.Name)
}
