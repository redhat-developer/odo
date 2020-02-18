package common

import (
	"github.com/openshift/odo/pkg/devfile"
)

// AdapterMetadata is a construct that is common to all adapters
type AdapterMetadata struct {
	Name    string
	Devfile devfile.DevfileObj
}
