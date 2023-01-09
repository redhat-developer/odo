package platform

import "fmt"

// PodNotFoundError returns an error if no pod is found with the selector
type PodNotFoundError struct {
	Selector string
}

func (e *PodNotFoundError) Error() string {
	return fmt.Sprintf("pod not found for the selector: %s", e.Selector)
}
