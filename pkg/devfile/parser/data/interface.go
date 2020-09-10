package data

import (
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	SetSchemaVersion(version string)

	GetMetadata() common.DevfileMetadata
	SetMetadata(name, version string)

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
	UpdateComponent(component common.DevfileComponent)
	GetAliasedComponents() []common.DevfileComponent

	// project related methods
	GetProjects() []common.DevfileProject
	AddProjects(projects []common.DevfileProject) error
	UpdateProject(project common.DevfileProject)

	// starter projects related commands
	GetStarterProjects() []common.DevfileStarterProject
	AddStarterProjects(projects []common.DevfileStarterProject) error
	UpdateStarterProject(project common.DevfileStarterProject)

	// command related methods
	GetCommands() map[string]common.DevfileCommand
	AddCommands(commands ...common.DevfileCommand) error
	UpdateCommand(command common.DevfileCommand)

	AddVolume(volume common.DevfileComponent, path string) error
	DeleteVolume(name string) error
	GetVolumeMountPath(name string) (string, error)
}
