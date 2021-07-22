package parser

import (
	"sigs.k8s.io/yaml"

	"github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// WriteYamlDevfile creates a devfile.yaml file
func (d *DevfileObj) WriteYamlDevfile() error {

	// Encode data into YAML format
	yamlData, err := yaml.Marshal(d.Data)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal devfile object into yaml")
	}
	// Write to devfile.yaml
	fs := d.Ctx.GetFs()
	if fs == nil {
		fs = filesystem.DefaultFs{}
	}
	err = fs.WriteFile(d.Ctx.GetAbsPath(), yamlData, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to create devfile yaml file")
	}

	// Successful
	klog.V(2).Infof("devfile yaml created at: '%s'", OutputDevfileYamlPath)
	return nil
}
