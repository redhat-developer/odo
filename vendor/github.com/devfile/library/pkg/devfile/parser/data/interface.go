package data

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {
	GetSchemaVersion() string
	SetSchemaVersion(version string)
	GetMetadata() devfilepkg.DevfileMetadata
	SetMetadata(name, version string)

	// parent related methods
	GetParent() *v1.Parent
	SetParent(parent *v1.Parent)

	// event related methods
	GetEvents() v1.Events
	AddEvents(events v1.Events) error
	UpdateEvents(postStart, postStop, preStart, preStop []string)

	// component related methods
	GetComponents(common.DevfileOptions) ([]v1.Component, error)
	AddComponents(components []v1.Component) error
	UpdateComponent(component v1.Component)

	// project related methods
	GetProjects(common.DevfileOptions) ([]v1.Project, error)
	AddProjects(projects []v1.Project) error
	UpdateProject(project v1.Project)

	// starter projects related commands
	GetStarterProjects(common.DevfileOptions) ([]v1.StarterProject, error)
	AddStarterProjects(projects []v1.StarterProject) error
	UpdateStarterProject(project v1.StarterProject)

	// command related methods
	GetCommands(common.DevfileOptions) ([]v1.Command, error)
	AddCommands(commands ...v1.Command) error
	UpdateCommand(command v1.Command)

	// volume related methods
	AddVolume(volume v1.Component, path string) error
	DeleteVolume(name string) error
	GetVolumeMountPath(name string) (string, error)

	// workspace related methods
	GetDevfileWorkspace() *v1.DevWorkspaceTemplateSpecContent
	SetDevfileWorkspace(content v1.DevWorkspaceTemplateSpecContent)

	//utils
	GetDevfileContainerComponents(common.DevfileOptions) ([]v1.Component, error)
	GetDevfileVolumeComponents(common.DevfileOptions) ([]v1.Component, error)
}
