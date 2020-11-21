package parser

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/klog"
)

// SetDevfileJSONSchema returns the JSON schema for the given devfile apiVersion
func (d *DevfileCtx) SetDevfileJSONSchema() error {

	// Check if json schema is present for the given apiVersion
	jsonSchema, err := data.GetDevfileJSONSchema(d.apiVersion)
	if err != nil {
		return err
	}
	d.jsonSchema = jsonSchema
	return nil
}

// ValidateDevfileSchema validate JSON schema of the provided devfile
func (d *DevfileCtx) ValidateDevfileSchema() error {
	var (
		schemaLoader   = gojsonschema.NewStringLoader(d.jsonSchema)
		documentLoader = gojsonschema.NewStringLoader(string(d.rawContent))
	)

	// Validate devfile with JSON schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return errors.Wrapf(err, "failed to validate devfile schema")
	}

	if !result.Valid() {
		errMsg := "invalid devfile schema. errors :\n"
		for _, desc := range result.Errors() {
			errMsg = errMsg + fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf(errMsg)
	}

	// Sucessful
	klog.V(4).Info("validated devfile schema")
	return nil
}
