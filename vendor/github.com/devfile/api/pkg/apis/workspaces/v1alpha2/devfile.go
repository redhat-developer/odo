package v1alpha2

import (
	"github.com/devfile/api/pkg/devfile"
)

// Devfile describes the structure of a cloud-native workspace and development environment.
// +devfile:jsonschema:generate:omitCustomUnionMembers=true
type Devfile struct {
	devfile.DevfileHeader `json:",inline"`

	DevWorkspaceTemplateSpec `json:",inline"`
}
