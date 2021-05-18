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
