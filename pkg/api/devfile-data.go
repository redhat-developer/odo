package api

import "github.com/devfile/library/v2/pkg/devfile/parser/data"

// DevfileData describes a devfile content
type DevfileData struct {
	Devfile              data.DevfileData      `json:"devfile"`
	Commands             []DevfileCommand      `json:"commands,omitempty"`
	SupportedOdoFeatures *SupportedOdoFeatures `json:"supportedOdoFeatures,omitempty"`
}

// SupportedOdoFeatures indicates the support of high-level (odo) features by a devfile component
type SupportedOdoFeatures struct {
	Dev    bool `json:"dev"`
	Deploy bool `json:"deploy"`
	Debug  bool `json:"debug"`
}

type DevfileCommand struct {
	Name          string               `json:"name,omitempty"`
	Type          DevfileCommandType   `json:"type,omitempty"`
	Group         DevfileCommandGroup  `json:"group,omitempty"`
	IsDefault     *bool                `json:"isDefault,omitempty"`
	CommandLine   string               `json:"commandLine,omitempty"`
	Component     string               `json:"component,omitempty"`
	ComponentType DevfileComponentType `json:"componentType,omitempty"`
	ImageName     string               `json:"imageName,omitempty"`
}

type DevfileCommandType string

const (
	ExecCommandType      DevfileCommandType = "exec"
	ApplyCommandType     DevfileCommandType = "apply"
	CompositeCommandType DevfileCommandType = "composite"
)

type DevfileCommandGroup string

const (
	BuildCommandGroup  DevfileCommandGroup = "build"
	RunCommandGroup    DevfileCommandGroup = "run"
	TestCommandGroup   DevfileCommandGroup = "test"
	DebugCommandGroup  DevfileCommandGroup = "debug"
	DeployCommandGroup DevfileCommandGroup = "deploy"
)

type DevfileComponentType string

const (
	ImageComponentType      DevfileComponentType = "image"
	ContainerComponentType  DevfileComponentType = "container"
	KubernetesComponentType DevfileComponentType = "kubernetes"
	OpenshiftComponentType  DevfileComponentType = "openshift"
)
