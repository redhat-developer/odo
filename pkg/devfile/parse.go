package devfile

import (
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/validate"
)

// Parse func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func Parse(path string) (d devfileParser.DevfileObj, err error) {

	// read and parse devfile from given path
	d, err = devfileParser.Parse(path)
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

// Parse func parses and validates the devfile integrity.
// Creates devfile context and runtime objects
func ParseInMemory(data []byte) (d devfileParser.DevfileObj, err error) {

	// read and parse devfile from given path
	d, err = devfileParser.ParseInMemory(data)
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
