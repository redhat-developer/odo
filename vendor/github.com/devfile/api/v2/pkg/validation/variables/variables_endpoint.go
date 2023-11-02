package variables

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// validateAndReplaceForEndpoint validates the endpoint data for global variable references and replaces them with the variable value
func validateAndReplaceForEndpoint(variables map[string]string, endpoints []v1alpha2.Endpoint) error {

	invalidKeys := make(map[string]bool)

	for i := range endpoints {
		var err error

		// Validate endpoint path
		if endpoints[i].Path, err = validateAndReplaceDataWithVariable(endpoints[i].Path, variables); err != nil {
			checkForInvalidError(invalidKeys, err)
		}
	}

	return newInvalidKeysError(invalidKeys)
}
