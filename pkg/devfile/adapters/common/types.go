package common

import (
	"github.com/openshift/odo/pkg/devfile"
)

// AdapterMetadata is a construct that is common to all adapters
type AdapterMetadata struct {
	Devfile devfile.DevfileObj
	OdoComp OdoComponent
}

// OdoComponent contains properties of the specific odo component based on the associated Devfile
type OdoComponent struct {
	Name string
}
