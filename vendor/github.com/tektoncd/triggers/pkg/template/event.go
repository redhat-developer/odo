/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// ResolveParams takes a given trigger binding and produces the resulting
// resource params.
func ResolveParams(bindings []*triggersv1.TriggerBinding, body []byte, header http.Header, params []pipelinev1.ParamSpec) ([]pipelinev1.Param, error) {
	out, err := MergeBindingParams(bindings)
	if err != nil {
		return nil, fmt.Errorf("error merging trigger params: %w", err)
	}

	out, err = applyEventValuesToParams(out, body, header)
	if err != nil {
		return nil, fmt.Errorf("failed to ApplyEventValuesToParams: %w", err)
	}
	return MergeInDefaultParams(out, params), nil
}

// ResolveResources resolves a templated resource by replacing params with their values.
func ResolveResources(template *triggersv1.TriggerTemplate, params []pipelinev1.Param) []json.RawMessage {
	resources := make([]json.RawMessage, len(template.Spec.ResourceTemplates))
	uid := UID()
	for i := range template.Spec.ResourceTemplates {
		resources[i] = ApplyParamsToResourceTemplate(params, template.Spec.ResourceTemplates[i].RawMessage)
		resources[i] = ApplyUIDToResourceTemplate(resources[i], uid)
	}
	return resources
}

// event represents a HTTP event that Triggers processes
type event struct {
	Header map[string]string `json:"header"`
	Body   interface{}       `json:"body"`
}

// newEvent returns a new Event from HTTP headers and body
func newEvent(body []byte, headers http.Header) (*event, error) {
	var data interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
		}
	}
	joinedHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		joinedHeaders[k] = strings.Join(v, ",")
	}

	return &event{
		Header: joinedHeaders,
		Body:   data,
	}, nil
}

// applyEventValuesToParams returns a slice of Params with the JSONPath variables replaced
// with values from the event body and headers.
func applyEventValuesToParams(params []pipelinev1.Param, body []byte, header http.Header) ([]pipelinev1.Param, error) {
	event, err := newEvent(body, header)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	for idx, p := range params {
		pValue := p.Value.StringVal
		// Find all expressions wrapped in $() from the value
		expressions := findTektonExpressions(pValue)
		for _, expr := range expressions {
			val, err := ParseJSONPath(event, expr)
			if err != nil {
				return nil, fmt.Errorf("failed to replace JSONPath value for param %s: %s: %w", p.Name, p.Value, err)
			}
			pValue = strings.ReplaceAll(pValue, expr, val)
		}
		params[idx].Value = pipelinev1.ArrayOrString{Type: pipelinev1.ParamTypeString, StringVal: pValue}
	}
	return params, nil
}
