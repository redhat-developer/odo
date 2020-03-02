package devfile

import (
	"encoding/json"

	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/versions"
	"github.com/pkg/errors"
)

// Parse func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func Parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = parser.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}

	// Validate devfile
	err = d.Ctx.Validate()
	if err != nil {
		return d, err
	}

	// Create a new devfile data object
	d.Data, err = versions.NewDevfileData(d.Ctx.GetApiVersion())
	if err != nil {
		return d, err
	}

	// Unmarshal devfile content into devfile struct
	err = json.Unmarshal(d.Ctx.GetDevfileContent(), &d.Data)
	if err != nil {
		return d, errors.Wrapf(err, "failed to decode devfile content")
	}

	// Successful
	return d, nil
}
