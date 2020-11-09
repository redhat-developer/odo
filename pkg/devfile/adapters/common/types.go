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
	AppName       string                   // the application name associated to a component
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

// PushParameters is a struct containing the parameters to be used when pushing to a devfile component
type PushParameters struct {
	Path                     string                  // Path refers to the parent folder containing the source code to push up to a component
	WatchFiles               []string                // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles        []string                // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	IgnoredFiles             []string                // IgnoredFiles is the list of files to not push up to a component
	ForceBuild               bool                    // ForceBuild determines whether or not to push all of the files up to a component or just files that have changed, added or removed.
	Show                     bool                    // Show tells whether the devfile command output should be shown on stdout
	DevfileBuildCmd          string                  // DevfileBuildCmd takes the build command through the command line and overwrites devfile build command
	DevfileRunCmd            string                  // DevfileRunCmd takes the run command through the command line and overwrites devfile run command
	DevfileDebugCmd          string                  // DevfileDebugCmd takes the debug command through the command line and overwrites the devfile debug command
	DevfileScanIndexForWatch bool                    // DevfileScanIndexForWatch is true if watch's push should regenerate the index file during SyncFiles, false otherwise. See 'pkg/sync/adapter.go' for details
	EnvSpecificInfo          envinfo.EnvSpecificInfo // EnvSpecificInfo contains information of env.yaml file
	Debug                    bool                    // Runs the component in debug mode
	DebugPort                int                     // Port used for remote debugging
	RunModeChanged           bool                    // It determines if run mode is changed from run to debug or vice versa
}

// SyncParameters is a struct containing the parameters to be used when syncing a devfile component
type SyncParameters struct {
	PushParams      PushParameters
	CompInfo        ComponentInfo
	PodChanged      bool
	ComponentExists bool
}

// ComponentInfo is a struct that holds information about a component i.e.; pod name, container name, and source mount (if applicable)
type ComponentInfo struct {
	PodName       string
	ContainerName string
	SyncFolder    string
}

func (ci ComponentInfo) IsEmpty() bool {
	return len(ci.ContainerName) == 0
}

// PushCommandsMap stores the commands to be executed as per their types.
type PushCommandsMap map[common.DevfileCommandGroupType]common.DevfileCommand

// NewPushCommandMap returns the instance of PushCommandsMap
func NewPushCommandMap() PushCommandsMap {
	return make(map[common.DevfileCommandGroupType]common.DevfileCommand)
}
