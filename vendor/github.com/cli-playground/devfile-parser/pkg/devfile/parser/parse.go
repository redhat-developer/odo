package parser

import (
	"encoding/json"

	devfileCtx "github.com/cli-playground/devfile-parser/pkg/devfile/parser/context"
	"github.com/cli-playground/devfile-parser/pkg/devfile/parser/data"
	"github.com/cli-playground/devfile-parser/pkg/devfile/validate"
	"github.com/cli-playground/devfile-parser/pkg/errors"
)

// ParseDevfile func validates the devfile integrity.
// Creates devfile context and runtime objects
func parseDevfile(d DevfileObj) (DevfileObj, error) {

	// Validate devfile
	err := d.Ctx.Validate()
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

// Parse func populates the devfile data, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func parse(path string) (d DevfileObj, err error) {

	// NewDevfileCtx
	d.Ctx = devfileCtx.NewDevfileCtx(path)

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.Populate()
	if err != nil {
		return d, err
	}
	return parseDevfile(d)
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

// parseInMemory func populates the data from memory, parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func parseInMemory(bytes []byte) (d DevfileObj, err error) {

	// Fill the fields of DevfileCtx struct
	err = d.Ctx.PopulateFromBytes(bytes)
	if err != nil {
		return d, err
	}
	return parseDevfile(d)
}

// ParseInMemoryAndValidate func parses the devfile data in memory
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
func ParseInMemoryAndValidate(data []byte) (d DevfileObj, err error) {

	// read and parse devfile from given data
	d, err = parseInMemory(data)
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
