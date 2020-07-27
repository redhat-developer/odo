package kclient

import "fmt"

// NoPodFoundError returns an error if no pod is found with the selector
type NoPodFoundError struct {
	Selector string
}

func (e *NoPodFoundError) Error() string {
	return fmt.Sprintf("no Pod was found for the selector: %s", e.Selector)
}
