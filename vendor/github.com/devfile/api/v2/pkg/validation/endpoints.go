package validation

import "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

// validateEndpoints checks if
// 1. all the endpoint names are unique across components
// 2. endpoint port are unique across component containers
func validateEndpoints(endpoints []v1alpha2.Endpoint, processedEndPointPort map[int]bool, processedEndPointName map[string]bool) (errList []error) {

	for _, endPoint := range endpoints {
		if _, ok := processedEndPointName[endPoint.Name]; ok {
			errList = append(errList, &InvalidEndpointError{name: endPoint.Name})
		}
		if _, ok := processedEndPointPort[endPoint.TargetPort]; ok {
			errList = append(errList, &InvalidEndpointError{port: endPoint.TargetPort})
		}
		processedEndPointName[endPoint.Name] = true
		processedEndPointPort[endPoint.TargetPort] = true
	}

	return errList
}
