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

package errors

import "fmt"

// NonCompliantDevfile returns an error if devfile parsing failed due to Non-Compliant Devfile
type NonCompliantDevfile struct {
	Err string
}

func (e *NonCompliantDevfile) Error() string {
	errMsg := "error parsing devfile because of non-compliant data"
	if e.Err != "" {
		errMsg = fmt.Sprintf("%s due to %v", errMsg, e.Err)
	}
	return errMsg
}
