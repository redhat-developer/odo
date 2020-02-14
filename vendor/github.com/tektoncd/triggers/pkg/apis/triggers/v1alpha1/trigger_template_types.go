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

package v1alpha1

import (
	"bytes"
	"encoding/json"
	"strings"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// apiVersion for Tekton core types
const tektonAPIVersion = GroupName + "/v1alpha1"

// allowedTemplate types are the resource types identified by apiVersion + kind
// that can be templated using TriggerResourceTemplates
// TODO: Replace static restrictions here with SubjectAccessReview create checks
var allowedTemplateTypes = map[string][]string{
	tektonAPIVersion: {"pipelineresource", "pipelinerun", "taskrun", "pipeline", "clustertask", "task", "condition"},
}

// Check that TriggerTemplate may be validated and defaulted.
var _ apis.Validatable = (*TriggerTemplate)(nil)
var _ apis.Defaultable = (*TriggerTemplate)(nil)

// TriggerTemplateSpec holds the desired state of TriggerTemplate
type TriggerTemplateSpec struct {
	Params            []pipelinev1.ParamSpec    `json:"params,omitempty"`
	ResourceTemplates []TriggerResourceTemplate `json:"resourcetemplates,omitempty"`
}

// TriggerResourceTemplate describes a resource to create
type TriggerResourceTemplate struct {
	json.RawMessage `json:",inline"`
}

// TriggerTemplateStatus describes the desired state of TriggerTemplate
type TriggerTemplateStatus struct{}

// TriggerTemplate takes parameters and uses them to create CRDs
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type TriggerTemplate struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec holds the desired state of the TriggerTemplate from the client
	// +optional
	Spec TriggerTemplateSpec `json:"spec"`
	// +optional
	Status TriggerTemplateStatus `json:"status"`
}

// TriggerTemplateList contains a list of TriggerTemplate
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TriggerTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TriggerTemplate `json:"items"`
}

// getAPIVersionAndKind returns the apiVersion and Kind for the resourceTemplate
// Missing fields are represented by empty strings
func (trt *TriggerResourceTemplate) getAPIVersionAndKind() (string, string) {
	var tm metav1.TypeMeta
	if err := json.NewDecoder(bytes.NewReader(trt.RawMessage)).Decode(&tm); err != nil {
		return "", ""
	}
	return tm.APIVersion, tm.Kind
}

// IsAllowedType returns true if the resourceTemplate has an apiVersion
// and kind field set to one of the allowed ones.
func (trt *TriggerResourceTemplate) IsAllowedType() bool {
	apiVersion, kind := trt.getAPIVersionAndKind()

	if kinds, ok := allowedTemplateTypes[apiVersion]; ok {
		for _, allowedKind := range kinds {
			if strings.ToLower(kind) == allowedKind {
				return true
			}
		}
	}
	return false
}
