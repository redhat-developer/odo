package data

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetMetadata() common.DevfileMetadata
	GetComponents() []common.DevfileComponent
	GetAliasedComponents() []common.DevfileComponent
	GetProjects() []common.DevfileProject
	GetCommands() []common.DevfileCommand
}
