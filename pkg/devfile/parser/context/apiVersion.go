package parser

import (
	"encoding/json"
	"fmt"

	"github.com/openshift/odo/pkg/devfile/parser/data"
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

	// Get "apiVersion" value from the map
	apiVersion, ok := r["apiVersion"]
	if !ok {
		return fmt.Errorf("apiVersion not present in devfile")
	}

	// apiVersion cannot be empty
	if apiVersion.(string) == "" {
		return fmt.Errorf("apiVersion in devfile cannot be empty")
	}

	// Successful
	d.apiVersion = apiVersion.(string)
	klog.V(4).Infof("devfile apiVersion: '%s'", d.apiVersion)
	return nil
}

// GetApiVersion returns apiVersion stored in devfile context
func (d *DevfileCtx) GetApiVersion() string {
	return d.apiVersion
}

// IsApiVersionSupported return true if the apiVersion in DevfileCtx is supported in odo
func (d *DevfileCtx) IsApiVersionSupported() bool {
	return data.IsApiVersionSupported(d.apiVersion)
}
