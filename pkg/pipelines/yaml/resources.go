package yaml

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

// WriteResources takes a prefix path, and a map of paths to values, and will
// marshal the values to the filenames as YAML resources, joining the prefix to
// the filenames before writing.
//
// It returns the list of filenames written out.
func WriteResources(fs afero.Fs, path string, files map[string]interface{}) ([]string, error) {
	filenames := make([]string, 0)
	for filename, item := range files {
		err := MarshalItemToFile(fs, filepath.Join(path, filename), item)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

// MarshalItemToFile marshals item to file
func MarshalItemToFile(fs afero.Fs, filename string, item interface{}) error {
	err := fs.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return fmt.Errorf("failed to MkDirAll for %s: %v", filename, err)
	}
	f, err := fs.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to Create file %s: %v", filename, err)
	}
	defer f.Close()
	return MarshalOutput(f, item)
}

// MarshalOutput marshal output to given writer
func MarshalOutput(out io.Writer, output interface{}) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}
	_, err = fmt.Fprintf(out, "%s", data)
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	return nil
}

// AddKustomize adds kustomization.yaml.  Name and items become map key and value, respectively
func AddKustomize(fs afero.Fs, name string, items []string, path string) error {
	content := []interface{}{}
	content = append(content, map[string]interface{}{name: items})
	return MarshalItemToFile(fs, path, content)
}
