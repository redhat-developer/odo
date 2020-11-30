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

	var apiVer string

	// Get "apiVersion" value from map for devfile V1
	apiVersion, okApi := r["apiVersion"]

	// Get "schemaVersion" value from map for devfile V2
	schemaVersion, okSchema := r["schemaVersion"]

	if okApi {
		apiVer = apiVersion.(string)
		// apiVersion cannot be empty
		if apiVer == "" {
			return fmt.Errorf("apiVersion in devfile cannot be empty")
		}

	} else if okSchema {
		apiVer = schemaVersion.(string)
		// SchemaVersion cannot be empty
		if schemaVersion.(string) == "" {
			return fmt.Errorf("schemaVersion in devfile cannot be empty")
		}
	} else {
		return fmt.Errorf("apiVersion or schemaVersion not present in devfile")

	}

	// Successful
	d.apiVersion = apiVer
	klog.V(4).Infof("devfile apiVersion: '%s'", d.apiVersion)
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
