package devfile

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/validate"
)

// ParseFromURLAndValidate func parses the devfile data from the url
// and validates the devfile integrity with the schema
// and validates the devfile data.
// Creates devfile context and runtime objects.
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
