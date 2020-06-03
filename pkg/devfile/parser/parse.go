package parser

import (
	"encoding/json"

	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/pkg/errors"
)

// parse func parses and validates the devfile integrity.
// Creates devfile context and runtime objects.
func parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}

	// Validate devfile schema
	err = d.Ctx.Validate()
	if err != nil {
		return d, err
	}

	// Create a new devfile data object
	d.Data, err = data.NewDevfileData(d.Ctx.GetApiVersion())
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

// ParseAndValidate func parses the devfile data
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
func ParseAndValidate(path string) (d DevfileObj, err error) {

	// read and parse devfile from given path
	d, err = parse(path)
	if err != nil {
		return d, err
	}

	// odo specific validation on devfile content
	err = validate.ValidateDevfileData(d.Data)
	if err != nil {
		return d, err
	}

	// Successful
	return d, nil
}
