package data

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetMetadata() common.DevfileMetadata
	GetParent() common.DevfileParent
	SetParent(parent common.DevfileParent)
	GetEvents() common.DevfileEvents
	AddEvents(events common.DevfileEvents) error
	UpdateEvents(postStart, postStop, preStart, preStop []string)
	GetComponents() []common.DevfileComponent
	AddComponents(components []common.DevfileComponent) error
	UpdateComponent(Name string, component common.DevfileComponent)
	GetAliasedComponents() []common.DevfileComponent
	GetProjects() []common.DevfileProject
	AddProjects(projects []common.DevfileProject) error
	UpdateProject(name string, project common.DevfileProject)
	GetCommands() []common.DevfileCommand
	AddCommands(commands []common.DevfileCommand) error
	UpdateCommand(id string, command common.DevfileCommand)
}
