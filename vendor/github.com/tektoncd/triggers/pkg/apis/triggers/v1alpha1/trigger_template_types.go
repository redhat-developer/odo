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
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/apis"
)

// Check that TriggerTemplate may be validated and defaulted.
var _ apis.Validatable = (*TriggerTemplate)(nil)
var _ apis.Defaultable = (*TriggerTemplate)(nil)

var Decoder runtime.Decoder

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(pipelinev1alpha1.AddToScheme(scheme))
	utilruntime.Must(pipelinev1beta1.AddToScheme(scheme))
	codec := serializer.NewCodecFactory(scheme)
	Decoder = codec.UniversalDecoder(
		pipelinev1alpha1.SchemeGroupVersion,
		pipelinev1beta1.SchemeGroupVersion,
	)
}

// TriggerTemplateSpec holds the desired state of TriggerTemplate
type TriggerTemplateSpec struct {
	Params            []ParamSpec               `json:"params,omitempty"`
	ResourceTemplates []TriggerResourceTemplate `json:"resourcetemplates,omitempty"`
}

// TriggerResourceTemplate describes a resource to create
type TriggerResourceTemplate struct {
	runtime.RawExtension `json:",inline"`
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
	Status TriggerTemplateStatus `json:"status,omitempty"`
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

// IsAllowedType returns true if the resourceTemplate has an apiVersion
// and kind field set to one of the allowed ones.
func (trt *TriggerResourceTemplate) IsAllowedType() error {
	_, err := runtime.Decode(Decoder, trt.RawExtension.Raw)
	return err
}
