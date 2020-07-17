package common

import (
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
)

// AdapterContext is a construct that is common to all adapters
type AdapterContext struct {
	ComponentName string                   // ComponentName is the odo component name, it is NOT related to any devfile components
	Context       string                   // Context is the given directory containing the source code and configs
	Devfile       devfileParser.DevfileObj // Devfile is the object returned by the Devfile parser
}

// DevfileVolume is a struct for Devfile volume that is common to all the adapters
type DevfileVolume struct {
	Name          string
	ContainerPath string
	Size          string
}

// Storage is a struct that is common to all the adapters
type Storage struct {
	Name   string
	Volume DevfileVolume
}

// BuildParameters is a struct containing the parameters to be used when building the image for a devfile component
type BuildParameters struct {
	Path            string                  // Path refers to the parent folder containing the source code to push up to a component
	DockerfileBytes []byte                  // DockerfileBytes is the contents of the project Dockerfile in bytes
	EnvSpecificInfo envinfo.EnvSpecificInfo // EnvSpecificInfo contains infomation of env.yaml file
	Tag             string                  // Tag refers to the image tag of the image being built
	IgnoredFiles    []string                // IgnoredFiles is the list of files to not push up to a component
}

// DeployParameters is a struct containing the parameters to be used when building the image for a devfile component
type DeployParameters struct {
	EnvSpecificInfo envinfo.EnvSpecificInfo // EnvSpecificInfo contains infomation of env.yaml file
	Tag             string                  // Tag refers to the image tag of the image being built
	ManifestSource  []byte                  // Source of the manifest file
	DeploymentPort  int                     // Port to be used in deployment manifest
}

// PushParameters is a struct containing the parameters to be used when pushing to a devfile component
type PushParameters struct {
	Path              string                  // Path refers to the parent folder containing the source code to push up to a component
	WatchFiles        []string                // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles []string                // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	IgnoredFiles      []string                // IgnoredFiles is the list of files to not push up to a component
	ForceBuild        bool                    // ForceBuild determines whether or not to push all of the files up to a component or just files that have changed, added or removed.
	Show              bool                    // Show tells whether the devfile command output should be shown on stdout
	DevfileInitCmd    string                  // DevfileInitCmd takes the init command through the command line and overwrites devfile init command
	DevfileBuildCmd   string                  // DevfileBuildCmd takes the build command through the command line and overwrites devfile build command
	DevfileRunCmd     string                  // DevfileRunCmd takes the run command through the command line and overwrites devfile run command
	DevfileDebugCmd   string                  // DevfileDebugCmd takes the debug command through the command line and overwrites the devfile debug command
	EnvSpecificInfo   envinfo.EnvSpecificInfo // EnvSpecificInfo contains infomation of env.yaml file
	Debug             bool                    // Runs the component in debug mode
	DebugPort         int                     // Port used for remote debugging
}

// SyncParameters is a struct containing the parameters to be used when syncing a devfile component
type SyncParameters struct {
	PushParams      PushParameters
	CompInfo        ComponentInfo
	PodChanged      bool
	ComponentExists bool
}

// ComponentInfo is a struct that holds information about a component i.e.; pod name, container name
type ComponentInfo struct {
	PodName       string
	ContainerName string
}

// PushCommandsMap stores the commands to be executed as per their types.
type PushCommandsMap map[common.DevfileCommandGroupType]common.DevfileCommand

// NewPushCommandMap returns the instance of PushCommandsMap
func NewPushCommandMap() PushCommandsMap {
	return make(map[common.DevfileCommandGroupType]common.DevfileCommand)
}
