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
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
)

//GetSchemaVersion gets devfile schema version
func (d *DevfileV2) GetSchemaVersion() string {
	return d.SchemaVersion
}

//SetSchemaVersion sets devfile schema version
func (d *DevfileV2) SetSchemaVersion(version string) {
	d.SchemaVersion = version
}

// GetMetadata returns the DevfileMetadata Object parsed from devfile
func (d *DevfileV2) GetMetadata() devfilepkg.DevfileMetadata {
	return d.Metadata
}

// SetMetadata sets the metadata for devfile
func (d *DevfileV2) SetMetadata(metadata devfilepkg.DevfileMetadata) {
	d.Metadata = metadata
}
