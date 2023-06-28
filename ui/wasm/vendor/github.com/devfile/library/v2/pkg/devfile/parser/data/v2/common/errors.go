//
// Copyright 2022 Red Hat, Inc.
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

package common

import "fmt"

// FieldAlreadyExistError error returned if tried to add already exisitng field
type FieldAlreadyExistError struct {
	// field which already exist
	Field string
	// field name
	Name string
}

func (e *FieldAlreadyExistError) Error() string {
	return fmt.Sprintf("%s %s already exists in devfile", e.Field, e.Name)
}

// FieldNotFoundError error returned if the field with the name is not found
type FieldNotFoundError struct {
	// field which doesn't exist
	Field string
	// field name
	Name string
}

func (e *FieldNotFoundError) Error() string {
	return fmt.Sprintf("%s %s is not found in the devfile", e.Field, e.Name)
}
