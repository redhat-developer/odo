package kclient

import "fmt"

// PodNotFoundError returns an error if no pod is found with the selector
type PodNotFoundError struct {
	Selector string
}

func (e *PodNotFoundError) Error() string {
	return fmt.Sprintf("pod not found for the selector: %s", e.Selector)
}

// DeploymentNotFoundError returns an error if no deployment is found with the selector
type DeploymentNotFoundError struct {
	Selector string
}

func (e *DeploymentNotFoundError) Error() string {
	return fmt.Sprintf("deployment not found for the selector: %s", e.Selector)
}

// ServiceNotFoundError returns an error if no service is found with the selector
type ServiceNotFoundError struct {
	Selector string
}

func (e *ServiceNotFoundError) Error() string {
	return fmt.Sprintf("service not found for the selector: %s", e.Selector)
}
