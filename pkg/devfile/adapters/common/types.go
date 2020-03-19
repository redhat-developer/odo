package common

import (
	"github.com/openshift/odo/pkg/devfile"
)

// AdapterContext is a construct that is common to all adapters
type AdapterContext struct {
	ComponentName string             // ComponentName is the odo component name, it is NOT related to any devfile components
	Devfile       devfile.DevfileObj // Devfile is the object returned by the Devfile parser
}

// DevfileVolume is a struct for Devfile volume that is common to all the adapters
type DevfileVolume struct {
	Name          *string
	ContainerPath *string
	Size          *string
}

// Storage is a struct that is common to all the adapters
type Storage struct {
	Name   string
	Volume DevfileVolume
}

// PushParameters is a struct containing the parameters to be used when pushing to a devfile component
type PushParameters struct {
	Path         string   // Path refers to the parent folder containing the source code to push up to a component
	Files        []string // Files is the list of changed files to push up to a component. If empty, odo will look at the index-json file under. odo
	IgnoredFiles []string // IgnoredFiles is the list of files to not push up to a component
	ForceBuild   bool     // ForceBuild determines whether or not to push all of the files up to a component or just some.
	GlobExps     []string
}
