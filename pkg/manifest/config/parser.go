package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// Parse decodes YAML describing an environment manifest.
func Parse(in io.Reader) (*Manifest, error) {
	dec := yaml.NewDecoder(in)
	m := &Manifest{}
	err := dec.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ParseFile is a wrapper around Parse that accepts a filename, it opens and
// parses the file, and closes it.
func ParseFile(filename string) (*Manifest, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
