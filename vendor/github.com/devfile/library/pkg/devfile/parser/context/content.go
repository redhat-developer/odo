package parser

import (
	"bytes"
	"unicode"

	"github.com/devfile/library/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

// Every JSON document starts with "{"
var jsonPrefix = []byte("{")

// YAMLToJSON converts a single YAML document into a JSON document
// or returns an error. If the document appears to be JSON the
// YAML decoding path is not used.
func YAMLToJSON(data []byte) ([]byte, error) {

	// Is already JSON
	if hasJSONPrefix(data) {
		return data, nil
	}

	// Is YAML, convert to JSON
	data, err := yaml.YAMLToJSON(data)
	if err != nil {
		return data, errors.Wrapf(err, "failed to convert devfile yaml to json")
	}

	// Successful
	klog.V(4).Infof("converted devfile YAML to JSON")
	return data, nil
}

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// hasPrefix returns true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

// SetDevfileContent reads devfile and if devfile is in YAML format converts it to JSON
func (d *DevfileCtx) SetDevfileContent() error {

	var err error
	var data []byte
	if d.url != "" {
		data, err = util.DownloadFileInMemory(d.url)
		if err != nil {
			return errors.Wrap(err, "error getting devfile info from url")
		}
	} else if d.absPath != "" {
		// Read devfile
		fs := d.GetFs()
		data, err = fs.ReadFile(d.absPath)
		if err != nil {
			return errors.Wrapf(err, "failed to read devfile from path '%s'", d.absPath)
		}
	}

	// set devfile content
	return d.SetDevfileContentFromBytes(data)
}

// SetDevfileContentFromBytes sets devfile content from byte input
func (d *DevfileCtx) SetDevfileContentFromBytes(data []byte) error {
	// If YAML file convert it to JSON
	var err error
	d.rawContent, err = YAMLToJSON(data)
	if err != nil {
		return err
	}

	// Successful
	return nil
}

// GetDevfileContent returns the devfile content
func (d *DevfileCtx) GetDevfileContent() []byte {
	return d.rawContent
}
