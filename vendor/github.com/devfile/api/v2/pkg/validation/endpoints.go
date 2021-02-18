package validation

import "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

// validateEndpoints checks if
// 1. all the endpoint names are unique across components
// 2. endpoint port are unique across component containers
// ie; two component containers cannot have the same target port but two endpoints
// in a single component container can have the same target port
func validateEndpoints(endpoints []v1alpha2.Endpoint, processedEndPointPort map[int]bool, processedEndPointName map[string]bool) error {
	currentComponentEndPointPort := make(map[int]bool)

	for _, endPoint := range endpoints {
		if _, ok := processedEndPointName[endPoint.Name]; ok {
			return &InvalidEndpointError{name: endPoint.Name}
		}
		processedEndPointName[endPoint.Name] = true
		currentComponentEndPointPort[endPoint.TargetPort] = true
	}

	for targetPort := range currentComponentEndPointPort {
		if _, ok := processedEndPointPort[targetPort]; ok {
			return &InvalidEndpointError{port: targetPort}
		}
		processedEndPointPort[targetPort] = true
	}
	return nil
}
