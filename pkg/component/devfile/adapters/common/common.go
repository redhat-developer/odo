package common

import (
	"github.com/openshift/odo/pkg/devfile"
)

// DevfileAdapter is a construct that is common to all adapters
type DevfileAdapter struct {
	Devfile     devfile.DevfileObj
	DevfileComp DevfileComponent
}

// DevfileComponent contains properties of the specific component related to the Devfile
type DevfileComponent struct {
	Name string
}
