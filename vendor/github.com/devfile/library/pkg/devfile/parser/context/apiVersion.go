package parser

import (
	"encoding/json"
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// SetDevfileAPIVersion returns the devfile APIVersion
func (d *DevfileCtx) SetDevfileAPIVersion() error {

	// Unmarshal JSON into map
	var r map[string]interface{}
	err := json.Unmarshal(d.rawContent, &r)
	if err != nil {
		return errors.Wrapf(err, "failed to decode devfile json")
	}

	// Get "schemaVersion" value from map for devfile V2
	schemaVersion, okSchema := r["schemaVersion"]
	var devfilePath string
	if d.GetAbsPath() != "" {
		devfilePath = d.GetAbsPath()
	} else if d.GetURL() != "" {
		devfilePath = d.GetURL()
	}

	if okSchema {
		// SchemaVersion cannot be empty
		if schemaVersion.(string) == "" {
			return fmt.Errorf("schemaVersion in devfile: %s cannot be empty", devfilePath)
		}
	} else {
		return fmt.Errorf("schemaVersion not present in devfile: %s", devfilePath)
	}

	// Successful
	d.apiVersion = schemaVersion.(string)
	klog.V(4).Infof("devfile schemaVersion: '%s'", d.apiVersion)
	return nil
}

// GetApiVersion returns apiVersion stored in devfile context
func (d *DevfileCtx) GetApiVersion() string {
	return d.apiVersion
}

// IsApiVersionSupported return true if the apiVersion in DevfileCtx is supported
func (d *DevfileCtx) IsApiVersionSupported() bool {
	return data.IsApiVersionSupported(d.apiVersion)
}
