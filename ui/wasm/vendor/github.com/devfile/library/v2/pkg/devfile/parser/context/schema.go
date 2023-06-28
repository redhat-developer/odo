//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"fmt"

	"github.com/devfile/library/v2/pkg/devfile/parser/data"
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
