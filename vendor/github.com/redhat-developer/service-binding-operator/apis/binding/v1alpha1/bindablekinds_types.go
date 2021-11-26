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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BindableKindsStatus defines the observed state of BindableKinds
type BindableKindsStatus struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// BindableKinds is the Schema for the bindablekinds API
type BindableKinds struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status []BindableKindsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BindableKindsList contains a list of BindableKinds
type BindableKindsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BindableKinds `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BindableKinds{}, &BindableKindsList{})
}
