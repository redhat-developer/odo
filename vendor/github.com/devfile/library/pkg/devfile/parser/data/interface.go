package data

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

// Generate mock interfaces for DevfileData by executing the following cmd in pkg/devfile/parser/data
// mockgen -package=data -source=interface.go DevfileData > /tmp/mock_interface.go ; cp /tmp/mock_interface.go ./mock_interface.go

// DevfileData is an interface that defines functions for Devfile data operations
type DevfileData interface {

	// header related methods

	GetSchemaVersion() string
	SetSchemaVersion(version string)
	GetMetadata() devfilepkg.DevfileMetadata
	SetMetadata(metadata devfilepkg.DevfileMetadata)

	// top-level attributes related method

	GetAttributes() (attributes.Attributes, error)
	AddAttributes(key string, value interface{}) error
	UpdateAttributes(key string, value interface{}) error

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
	UpdateComponent(component v1.Component) error
	DeleteComponent(name string) error

	// project related methods

	GetProjects(common.DevfileOptions) ([]v1.Project, error)
	AddProjects(projects []v1.Project) error
	UpdateProject(project v1.Project) error
	DeleteProject(name string) error

	// starter projects related commands

	GetStarterProjects(common.DevfileOptions) ([]v1.StarterProject, error)
	AddStarterProjects(projects []v1.StarterProject) error
	UpdateStarterProject(project v1.StarterProject) error
	DeleteStarterProject(name string) error

	// command related methods

	GetCommands(common.DevfileOptions) ([]v1.Command, error)
	AddCommands(commands []v1.Command) error
	UpdateCommand(command v1.Command) error
	DeleteCommand(id string) error

	// volume mount related methods

	AddVolumeMounts(containerName string, volumeMounts []v1.VolumeMount) error
	DeleteVolumeMount(name string) error
	GetVolumeMountPaths(mountName, containerName string) ([]string, error)

	// workspace related methods

	GetDevfileWorkspaceSpecContent() *v1.DevWorkspaceTemplateSpecContent
	SetDevfileWorkspaceSpecContent(content v1.DevWorkspaceTemplateSpecContent)
	GetDevfileWorkspaceSpec() *v1.DevWorkspaceTemplateSpec
	SetDevfileWorkspaceSpec(spec v1.DevWorkspaceTemplateSpec)

	// utils

	GetDevfileContainerComponents(common.DevfileOptions) ([]v1.Component, error)
	GetDevfileVolumeComponents(common.DevfileOptions) ([]v1.Component, error)
}
