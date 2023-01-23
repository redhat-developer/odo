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
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetDevfileWorkspaceSpecContent returns the workspace spec content for the devfile
func (d *DevfileV2) GetDevfileWorkspaceSpecContent() *v1.DevWorkspaceTemplateSpecContent {

	return &d.DevWorkspaceTemplateSpecContent
}

// SetDevfileWorkspaceSpecContent sets the workspace spec content
func (d *DevfileV2) SetDevfileWorkspaceSpecContent(content v1.DevWorkspaceTemplateSpecContent) {
	d.DevWorkspaceTemplateSpecContent = content
}

func (d *DevfileV2) GetDevfileWorkspaceSpec() *v1.DevWorkspaceTemplateSpec {
	return &d.DevWorkspaceTemplateSpec
}

// SetDevfileWorkspaceSpec sets the workspace spec
func (d *DevfileV2) SetDevfileWorkspaceSpec(spec v1.DevWorkspaceTemplateSpec) {
	d.DevWorkspaceTemplateSpec = spec
}
