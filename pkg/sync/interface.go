package sync

import "io"

// ComponentInfo is a struct that holds information about a component i.e.; component name, pod name, container name, and source mount (if applicable)
type ComponentInfo struct {
	ComponentName string
	PodName       string
	ContainerName string
	SyncFolder    string
}

type SyncExtracter func(ComponentInfo, string, io.Reader) error

// SyncParameters is a struct containing the parameters to be used when syncing a devfile component
type SyncParameters struct {
	Path                     string   // Path refers to the parent folder containing the source code to push up to a component
	WatchFiles               []string // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles        []string // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	IgnoredFiles             []string // IgnoredFiles is the list of files to not push up to a component
	DevfileScanIndexForWatch bool     // DevfileScanIndexForWatch is true if watch's push should regenerate the index file during SyncFiles, false otherwise. See 'pkg/sync/adapter.go' for details
	ForcePush                bool
	CompInfo                 ComponentInfo
	Files                    map[string]string
}

type Client interface {
	SyncFiles(syncParameters SyncParameters) (bool, error)
}
