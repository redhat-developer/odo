package devfile

import (
	"github.com/devfile/api/v2/pkg/validation/variables"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/validate"
)

// ParseFromURLAndValidate func parses the devfile data from the url
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
// Deprecated, use ParseDevfileAndValidate() instead
func ParseFromURLAndValidate(url string) (d parser.DevfileObj, err error) {

	// read and parse devfile from the given URL
	d, err = parser.ParseFromURL(url)
	if err != nil {
		return d, err
	}

	// generic validation on devfile content
	err = validate.ValidateDevfileData(d.Data)
	if err != nil {
		return d, err
	}

	return d, err
}

// ParseFromDataAndValidate func parses the devfile data
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
// Deprecated, use ParseDevfileAndValidate() instead
func ParseFromDataAndValidate(data []byte) (d parser.DevfileObj, err error) {
	// read and parse devfile from the given bytes
	d, err = parser.ParseFromData(data)
	if err != nil {
		return d, err
	}
	// generic validation on devfile content
	err = validate.ValidateDevfileData(d.Data)
	if err != nil {
		return d, err
	}

	return d, err
}

// ParseAndValidate func parses the devfile data
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
// Deprecated, use ParseDevfileAndValidate() instead
func ParseAndValidate(path string) (d parser.DevfileObj, err error) {

	// read and parse devfile from given path
	d, err = parser.Parse(path)
	if err != nil {
		return d, err
	}

	// generic validation on devfile content
	err = validate.ValidateDevfileData(d.Data)
	if err != nil {
		return d, err
	}

	return d, err
}

// ParseDevfileAndValidate func parses the devfile data, validates the devfile integrity with the schema
// replaces the top-level variable keys if present and validates the devfile data.
// It returns devfile context and runtime objects, variable substitution warning if any and an error.
func ParseDevfileAndValidate(args parser.ParserArgs) (d parser.DevfileObj, varWarning variables.VariableWarning, err error) {
	d, err = parser.ParseDevfile(args)
	if err != nil {
		return d, varWarning, err
	}

	if d.Data.GetSchemaVersion() != "2.0.0" {
		// replace the top level variable keys with their values in the devfile
		varWarning = variables.ValidateAndReplaceGlobalVariable(d.Data.GetDevfileWorkspaceSpec())
	}

	// generic validation on devfile content
	err = validate.ValidateDevfileData(d.Data)
	if err != nil {
		return d, varWarning, err
	}

	return d, varWarning, err
}
