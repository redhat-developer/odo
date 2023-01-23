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

package v2

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/attributes"
)

// GetAttributes gets the devfile top level attributes
func (d *DevfileV2) GetAttributes() (attributes.Attributes, error) {
	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return attributes.Attributes{}, fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		return d.Attributes, nil
	}
}

// UpdateAttributes updates the devfile top level attribute for the specific key, err out if key is absent
func (d *DevfileV2) UpdateAttributes(key string, value interface{}) error {
	var err error

	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		if d.Attributes.Exists(key) {
			d.Attributes.Put(key, value, &err)
		} else {
			return fmt.Errorf("cannot update top-level attribute, key %s is not present", key)
		}
	}

	return err
}

// AddAttributes adds to the devfile top level attributes, value will be overwritten if key is already present
func (d *DevfileV2) AddAttributes(key string, value interface{}) error {
	var err error

	// This feature was introduced in 2.1.0; so any version 2.1.0 and up should use the 2.1.0 implementation
	switch d.SchemaVersion {
	case "2.0.0":
		return fmt.Errorf("top-level attributes is not supported in devfile schema version 2.0.0")
	default:
		d.Attributes.Put(key, value, &err)
	}

	return err
}
