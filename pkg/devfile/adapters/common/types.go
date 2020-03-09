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
