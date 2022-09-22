package adapters

import (
	"io"

	"github.com/redhat-developer/odo/pkg/envinfo"
)

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
	EnvSpecificInfo          envinfo.EnvSpecificInfo // EnvSpecificInfo contains information of devfile
	Debug                    bool                    // Runs the component in debug mode
	RandomPorts              bool                    // True to forward containers ports on local random ports
	ErrOut                   io.Writer               // Writer to output forwarded port information
}

// SyncParameters is a struct containing the parameters to be used when syncing a devfile component
type SyncParameters struct {
	Path                     string   // Path refers to the parent folder containing the source code to push up to a component
	WatchFiles               []string // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles        []string // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	IgnoredFiles             []string // IgnoredFiles is the list of files to not push up to a component
	ForceBuild               bool     // ForceBuild determines whether or not to push all of the files up to a component or just files that have changed, added or removed.
	DevfileScanIndexForWatch bool     // DevfileScanIndexForWatch is true if watch's push should regenerate the index file during SyncFiles, false otherwise. See 'pkg/sync/adapter.go' for details

	CompInfo        ComponentInfo
	PodChanged      bool
	ComponentExists bool
	Files           map[string]string
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
