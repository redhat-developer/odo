package kclient

import "fmt"

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
	return fmt.Sprintf("service not found for the selector %q", e.Selector)
}

type NoConnectionError struct{}

func NewNoConnectionError() NoConnectionError {
	return NoConnectionError{}
}

func (e NoConnectionError) Error() string {
	// could also be "cluster is non accessible"
	return "unable to access the cluster"
}
