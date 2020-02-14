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
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// Check that TriggerBinding may be validated and defaulted.
var _ apis.Validatable = (*TriggerBinding)(nil)

// TriggerBindingSpec defines the desired state of the TriggerBinding.
type TriggerBindingSpec struct {
	// Params defines the parameter mapping from the given input event.
	Params []pipelinev1.Param `json:"params,omitempty"`
}

// TriggerBindingStatus defines the observed state of TriggerBinding.
type TriggerBindingStatus struct{}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TriggerBinding defines a mapping of an input event to parameters. This is used
// to extract information from events to be passed to TriggerTemplates within a
// Trigger.
// +k8s:openapi-gen=true
type TriggerBinding struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec holds the desired state of the TriggerBinding
	// +optional
	Spec TriggerBindingSpec `json:"spec"`
	// +optional
	Status TriggerBindingStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TriggerBindingList contains a list of TriggerBindings.
// We don't use this but it's required for certain codegen features.
type TriggerBindingList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TriggerBinding `json:"items"`
}
