package yaml

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// WriteResources takes a prefix path, and a map of paths to values, and will
// marshal the values to the filenames as YAML resources, joining the prefix to
// the filenames before writing.
//
// It returns the list of filenames written out.
func WriteResources(path string, files map[string]interface{}) ([]string, error) {
	filenames := make([]string, 0)
	for filename, item := range files {
		err := marshalItemsToFile(filepath.Join(path, filename), list(item))
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

func marshalItemsToFile(filename string, items []interface{}) error {
	err := os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to MkDirAll for %s: %v", filename, err)
	}
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to Create file %s: %v", filename, err)
	}
	defer f.Close()
	return marshalOutputs(f, items)
}

func list(i interface{}) []interface{} {
	return []interface{}{i}
}

// marshalOutputs marshal outputs to given writer
func marshalOutputs(out io.Writer, outputs []interface{}) error {
	for _, r := range outputs {
		data, err := yaml.Marshal(r)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		_, err = fmt.Fprintf(out, "%s---\n", data)
		if err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}
	return nil
}
