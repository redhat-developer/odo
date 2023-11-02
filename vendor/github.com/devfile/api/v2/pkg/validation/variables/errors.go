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
	"fmt"
	"sort"
	"strings"
)

// InvalidKeysError returns an error for the invalid keys
type InvalidKeysError struct {
	Keys []string
}

func (e *InvalidKeysError) Error() string {
	return fmt.Sprintf("invalid variable references - %s", strings.Join(e.Keys, ","))
}

// newInvalidKeysError processes the invalid key set and returns an InvalidKeysError if present
func newInvalidKeysError(keySet map[string]bool) error {
	var invalidKeysArr []string
	for key := range keySet {
		invalidKeysArr = append(invalidKeysArr, key)
	}

	if len(invalidKeysArr) > 0 {
		sort.Strings(invalidKeysArr)
		return &InvalidKeysError{Keys: invalidKeysArr}
	}

	return nil
}
