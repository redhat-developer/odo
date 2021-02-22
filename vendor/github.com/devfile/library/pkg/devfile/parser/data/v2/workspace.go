package v2

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// GetDevfileWorkspace returns the workspace content for the devfile
func (d *DevfileV2) GetDevfileWorkspace() *v1.DevWorkspaceTemplateSpecContent {

	return &d.DevWorkspaceTemplateSpecContent
}

// SetDevfileWorkspace sets the workspace content
func (d *DevfileV2) SetDevfileWorkspace(content v1.DevWorkspaceTemplateSpecContent) {
	d.DevWorkspaceTemplateSpecContent = content
}
