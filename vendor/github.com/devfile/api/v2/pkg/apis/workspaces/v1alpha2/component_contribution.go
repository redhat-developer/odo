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

import attributes "github.com/devfile/api/v2/pkg/attributes"

type ComponentContribution struct {
	// Mandatory name that allows referencing the component
	// from other elements (such as commands) or from an external
	// devfile that may reference this component through a parent or a plugin.
	// +kubebuilder:validation:Pattern=^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name"`
	// Map of implementation-dependant free-form YAML attributes.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Attributes attributes.Attributes `json:"attributes,omitempty"`
	// Reference to a remote Devfile or DevWorkspace resource to be included
	// in the DevWorkspace.
	PluginComponent `json:",inline"`
}
