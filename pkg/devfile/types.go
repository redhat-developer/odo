package devfile

import (
	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/versions"
)

// Default filenames for create devfile
const (
	OutputDevfileJsonPath = "odo-devfile.json"
	OutputDevfileYamlPath = "odo-devfile.yaml"
)

// DevfileObj is the runtime devfile object
type DevfileObj struct {

	// Ctx has devfile context info
	Ctx parser.DevfileCtx

	// Data has the devfile data
	Data versions.DevfileData
}
