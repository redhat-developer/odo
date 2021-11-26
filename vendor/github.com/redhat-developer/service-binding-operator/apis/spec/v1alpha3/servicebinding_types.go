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

package v1alpha3

import (
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ServiceBindingWorkloadReference defines a subset of corev1.ObjectReference with extensions
type ServiceBindingWorkloadReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name,omitempty"`
	// Selector is a query that selects the workload or workloads to bind the service to
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	// Containers describes which containers in a Pod should be bound to
	Containers []string `json:"containers,omitempty"`
}

// ServiceBindingServiceReference defines a subset of corev1.ObjectReference
type ServiceBindingServiceReference struct {
	// API version of the referent.
	APIVersion string `json:"apiVersion"`
	// Kind of the referent.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`
}

// ServiceBindingSecretReference defines a mirror of corev1.LocalObjectReference
type ServiceBindingSecretReference struct {
	// Name of the referent secret.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`
}

// EnvMapping defines a mapping from the value of a Secret entry to an environment variable
type EnvMapping struct {
	// Name is the name of the environment variable
	Name string `json:"name"`
	// Key is the key in the Secret that will be exposed
	Key string `json:"key"`
}

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	// Name is the name of the service as projected into the workload container.  Defaults to .metadata.name.
	// +kubebuilder:validation:Pattern=`^[a-z0-9\-\.]*$`
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`
	// Type is the type of the service as projected into the workload container
	Type string `json:"type,omitempty"`
	// Provider is the provider of the service as projected into the workload container
	Provider string `json:"provider,omitempty"`
	// Workload is a reference to an object
	Workload ServiceBindingWorkloadReference `json:"workload"`
	// Service is a reference to an object that fulfills the ProvisionedService duck type
	Service ServiceBindingServiceReference `json:"service"`
	// Env is the collection of mappings from Secret entries to environment variables
	Env []EnvMapping `json:"env,omitempty"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
type ServiceBindingStatus struct {
	// ObservedGeneration is the 'Generation' of the ServiceBinding that
	// was last processed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions are the conditions of this ServiceBinding
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Binding exposes the projected secret for this ServiceBinding
	Binding *ServiceBindingSecretReference `json:"binding,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ServiceBinding is the Schema for the servicebindings API
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}

func (ref *ServiceBindingServiceReference) GroupVersionResource() (*schema.GroupVersionResource, error) {
	return nil, errors.New("Resource undefined")
}

func (ref *ServiceBindingServiceReference) GroupVersionKind() (*schema.GroupVersionKind, error) {
	typeMeta := &metav1.TypeMeta{Kind: ref.Kind, APIVersion: ref.APIVersion}
	gvk := typeMeta.GroupVersionKind()
	return &gvk, nil
}

func (ref *ServiceBindingWorkloadReference) GroupVersionResource() (*schema.GroupVersionResource, error) {
	return nil, errors.New("Resource undefined")
}

func (ref *ServiceBindingWorkloadReference) GroupVersionKind() (*schema.GroupVersionKind, error) {
	typeMeta := &metav1.TypeMeta{Kind: ref.Kind, APIVersion: ref.APIVersion}
	gvk := typeMeta.GroupVersionKind()
	return &gvk, nil
}

func (sb *ServiceBinding) AsOwnerReference() metav1.OwnerReference {
	var ownerRefController bool = true
	return metav1.OwnerReference{
		Name:       sb.Name,
		UID:        sb.UID,
		Kind:       sb.Kind,
		APIVersion: sb.APIVersion,
		Controller: &ownerRefController,
	}
}

func (sb *ServiceBinding) HasDeletionTimestamp() bool {
	return !sb.DeletionTimestamp.IsZero()
}

func (r *ServiceBinding) StatusConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (sb *ServiceBinding) GetSpec() interface{} {
	return &sb.Spec
}
