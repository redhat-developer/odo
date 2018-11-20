package models

import "github.com/redhat-developer/odo/pkg/util"

// CreateType is an enum to indicate the type of source of component -- local source/binary or git for the generation of app/component names
type CreateType string

const (
	// GIT as source of component
	GIT CreateType = "git"
	// LOCAL Local source path as source of component
	LOCAL CreateType = "local"
	// BINARY Local Binary as source of component
	BINARY CreateType = "binary"
	// NONE indicates there's no information about the type of source of the component
	NONE CreateType = ""
)

// CreateArgs is a container of attributes of component create action
type CreateArgs struct {
	Name            string
	SourcePath      string
	SourceType      CreateType
	ImageName       string
	EnvVars         []string
	Ports           []string
	Resources       []util.ResourceRequirementInfo
	ApplicationName string
}
