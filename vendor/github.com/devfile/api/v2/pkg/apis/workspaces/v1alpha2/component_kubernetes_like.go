//
//
// Copyright Red Hat
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

package v1alpha2

// K8sLikeComponentLocationType describes the type of
// the location the configuration is fetched from.
// Only one of the following component type may be specified.
// +kubebuilder:validation:Enum=Uri;Inlined
type K8sLikeComponentLocationType string

const (
	UriK8sLikeComponentLocationType     K8sLikeComponentLocationType = "Uri"
	InlinedK8sLikeComponentLocationType K8sLikeComponentLocationType = "Inlined"
)

// +union
type K8sLikeComponentLocation struct {
	// Type of Kubernetes-like location
	// +
	// +unionDiscriminator
	// +optional
	LocationType K8sLikeComponentLocationType `json:"locationType,omitempty"`

	// Location in a file fetched from a uri.
	// +optional
	Uri string `json:"uri,omitempty"`

	// Inlined manifest
	// +optional
	Inlined string `json:"inlined,omitempty"`
}

// +devfile:getter:generate
type K8sLikeComponent struct {
	BaseComponent            `json:",inline"`
	K8sLikeComponentLocation `json:",inline"`

	// Defines if the component should be deployed during startup.
	//
	// Default value is `false`
	// +optional
	// +devfile:default:value=false
	DeployByDefault *bool `json:"deployByDefault,omitempty"`

	Endpoints []Endpoint `json:"endpoints,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

// Component that allows partly importing Kubernetes resources into the devworkspace POD
type KubernetesComponent struct {
	K8sLikeComponent `json:",inline"`
}

// Component that allows partly importing Openshift resources into the devworkspace POD
type OpenshiftComponent struct {
	K8sLikeComponent `json:",inline"`
}
