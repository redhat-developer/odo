package common

import (
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/dev"
)

// PushParameters is a struct containing the parameters to be used when pushing to a devfile component
type PushParameters struct {
	StartOptions dev.StartOptions

	Devfile                  parser.DevfileObj
	WatchFiles               []string // Optional: WatchFiles is the list of changed files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine changed files
	WatchDeletedFiles        []string // Optional: WatchDeletedFiles is the list of deleted files detected by odo watch. If empty or nil, odo will check .odo/odo-file-index.json to determine deleted files
	Show                     bool     // Show tells whether the devfile command output should be shown on stdout
	DevfileScanIndexForWatch bool     // DevfileScanIndexForWatch is true if watch's push should regenerate the index file during SyncFiles, false otherwise. See 'pkg/sync/adapter.go' for details
}
