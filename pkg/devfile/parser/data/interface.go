package data

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetMetadata() common.DevfileMetadata

	// parent related methods
	GetParent() common.DevfileParent
	SetParent(parent common.DevfileParent)

	// event related methods
	GetEvents() common.DevfileEvents
	AddEvents(events common.DevfileEvents) error
	UpdateEvents(postStart, postStop, preStart, preStop []string)

	// component related methods
	GetComponents() []common.DevfileComponent
	AddComponents(components []common.DevfileComponent) error
	UpdateComponent(Name string, component common.DevfileComponent)
	GetAliasedComponents() []common.DevfileComponent

	// project related methods
	GetProjects() []common.DevfileProject
	AddProjects(projects []common.DevfileProject) error
	UpdateProject(name string, project common.DevfileProject)

	// command related methods
	GetCommands() []common.DevfileCommand
	AddCommands(commands []common.DevfileCommand) error
	UpdateCommand(id string, command common.DevfileCommand)
}
