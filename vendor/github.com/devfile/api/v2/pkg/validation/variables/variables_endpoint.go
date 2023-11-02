//
//
// Copyright Red Hat
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
