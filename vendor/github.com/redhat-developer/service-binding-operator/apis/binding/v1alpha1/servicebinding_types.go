/*
Copyright 2021.

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
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var templates = map[string]string{
	"none":      "{{ .name }}",
	"uppercase": "{{ .service.kind | upper }}_{{ .name | upper }}",
	"lowercase": "{{ .name | lower }}",
}

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	// MountPath is the path inside app container where bindings will be mounted
	// If `SERVICE_BINDING_ROOT` env var is present, mountPath is ignored.
	// If `SERVICE_BINDING_ROOT` is absent and mountPath is present, set `SERVICE_BINDING_ROOT` as mountPath value
	// If `SERVICE_BINDING_ROOT` is absent but mounthPath is absent, set   SERVICE_BINDING_ROOT as `/bindings`
	// When mountPath is used, the file will be mounted directly under that directory
	// Otherwise it will be under `SERVICE_BINDING_ROOT`/<SERVICE-BINDING-NAME>
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// NamingStrategy defines custom string template for preparing binding names.
	// It can be pre-defined strategies(i.e none,uppercase), in case strategy provided in this field isn't defined
	// we are going to treat the value as a custom template and prepare binding names accordingly.
	// +optional
	NamingStrategy string `json:"namingStrategy,omitempty"`

	// Custom mappings
	// +optional
	Mappings []Mapping `json:"mappings,omitempty"`

	// Services is used to identify multiple backing services.
	// +kubebuilder:validation:MinItems:=1
	Services []Service `json:"services"`

	// Application is used to identify the application connecting to the
	// backing service operator.
	Application Application `json:"application"`

	// DetectBindingResources is flag used to bind all non-bindable variables from
	// different subresources owned by backing operator CR.
	// +optional
	DetectBindingResources bool `json:"detectBindingResources,omitempty"`

	// BindAsFiles makes available the binding values as files in the application's container
	// See MountPath attribute description for more details.
	// +optional
	// +kubebuilder:default:=true
	BindAsFiles bool `json:"bindAsFiles"`
}

// ServiceBindingMapping defines a new binding from set of existing bindings
type Mapping struct {
	// Name is the name of new binding
	Name string `json:"name"`
	// Value is a template which will be rendered and ibjected into the application
	Value string `json:"value"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
// +k8s:openapi-gen=true
type ServiceBindingStatus struct {
	// Conditions describes the state of the operator's reconciliation functionality.
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Secret is the name of the intermediate secret
	Secret string `json:"secret"`

	// Application defines the application workloads to which the binding secret has
	// injected
	Applications []Ref `json:"applications,omitempty"`
}

// Object reference in the same namespace
// +mapType=atomic
type Ref struct {

	// Group of the referent.
	Group string `json:"group"`

	// Version of the referent.
	Version string `json:"version"`

	// Kind of the referent.
	// +optional
	Kind string `json:"kind,omitempty"`

	// Resource of the referent.
	// +optional
	Resource string `json:"resource,omitempty"`

	// Name of the referent.
	Name string `json:"name,omitempty"`
}

// Object reference in some namespace
type NamespacedRef struct {
	Ref `json:",inline"`

	// Namespace of the referent.
	// if empty assumes the same namespace as ServiceBinding
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// Service defines the selector based on resource name, version, and resource kind
type Service struct {
	NamespacedRef `json:",inline"`

	Id *string `json:"id,omitempty"`
}

// Application defines the selector based on labels and GVR
type Application struct {
	Ref `json:",inline"`
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`

	// BindingPath refers to the paths in the application workload's schema
	// where the binding workload would be referenced.
	// If BindingPath is not specified the default path locations is going to
	// be used.  The default location for ContainersPath is
	// going to be: "spec.template.spec.containers" and if SecretPath
	// is not specified, the name of the secret object is not going
	// to be specified.
	// +optional
	BindingPath *BindingPath `json:"bindingPath,omitempty"`
}

// BindingPath defines the path to the field where the binding would be
// embedded in the workload
type BindingPath struct {
	// ContainersPath defines the path to the corev1.Containers reference
	// If BindingPath is not specified, the default location is
	// going to be: "spec.template.spec.containers"
	// +optional
	ContainersPath string `json:"containersPath"`

	// SecretPath defines the path to a string field where
	// the name of the secret object is going to be assigned.
	// Note: The name of the secret object is same as that of the name of SBR CR (metadata.name)
	// +optional
	SecretPath string `json:"secretPath"`
}

// ServiceBinding expresses intent to bind an operator-backed service with
// an application workload.
// +kubebuilder:subresource:status
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Service Binding"
// +kubebuilder:resource:path=servicebindings,shortName=sbr;sbrs
// +kubebuilder:object:root=true
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}

func (sbr *ServiceBinding) AsOwnerReference() metav1.OwnerReference {
	var ownerRefController bool = true
	return metav1.OwnerReference{
		Name:       sbr.Name,
		UID:        sbr.UID,
		Kind:       sbr.Kind,
		APIVersion: sbr.APIVersion,
		Controller: &ownerRefController,
	}
}

func (spec *ServiceBindingSpec) NamingTemplate() string {
	if spec.NamingStrategy != "" {
		if v, ok := templates[spec.NamingStrategy]; ok {
			return v
		} else {
			return spec.NamingStrategy
		}
	}
	if spec.BindAsFiles {
		return templates["none"]
	} else {
		return templates["uppercase"]
	}
}

func (sb *ServiceBinding) HasDeletionTimestamp() bool {
	return !sb.DeletionTimestamp.IsZero()
}

// Returns GVR of reference if available, otherwise error
func (ref *Ref) GroupVersionResource() (*schema.GroupVersionResource, error) {
	if ref.Resource == "" {
		return nil, errors.New("Resource undefined")
	}
	return &schema.GroupVersionResource{
		Group:    ref.Group,
		Version:  ref.Version,
		Resource: ref.Resource,
	}, nil
}

// Returns GVK of reference if available, otherwise error
func (ref *Ref) GroupVersionKind() (*schema.GroupVersionKind, error) {
	if ref.Kind == "" {
		return nil, errors.New("Kind undefined")
	}
	return &schema.GroupVersionKind{
		Group:   ref.Group,
		Version: ref.Version,
		Kind:    ref.Kind,
	}, nil
}

func (r *ServiceBinding) StatusConditions() []metav1.Condition {
	return r.Status.Conditions
}
