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

package v1alpha2

import (
	"fmt"
	"reflect"
)

func extractKeys(keyedList interface{}) []Keyed {
	value := reflect.ValueOf(keyedList)
	keys := make([]Keyed, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.CanInterface() {
			i := elem.Interface()
			if keyed, ok := i.(Keyed); ok {
				keys = append(keys, keyed)
			}
		}
	}
	return keys
}

// CheckDuplicateKeys checks if duplicate keys are present in the devfile objects
func CheckDuplicateKeys(keyedList interface{}) error {
	seen := map[string]bool{}
	value := reflect.ValueOf(keyedList)
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.CanInterface() {
			i := elem.Interface()
			if keyed, ok := i.(Keyed); ok {
				key := keyed.Key()
				if seen[key] {
					return fmt.Errorf("duplicate key: %s", key)
				}
				seen[key] = true
			}
		}
	}
	return nil
}
