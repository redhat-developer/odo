package parser

import (
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
)

// Default filenames for create devfile
const (
	OutputDevfileYamlPath = "devfile.yaml"
)

// DevfileObj is the runtime devfile object
type DevfileObj struct {

	// Ctx has devfile context info
	Ctx devfileCtx.DevfileCtx

	// Data has the devfile data
	Data data.DevfileData
}
