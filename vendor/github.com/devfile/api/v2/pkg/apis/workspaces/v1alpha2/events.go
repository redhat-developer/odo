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

type Events struct {
	DevWorkspaceEvents `json:",inline"`
}

type DevWorkspaceEvents struct {
	// IDs of commands that should be executed before the devworkspace start.
	// Kubernetes-wise, these commands would typically be executed in init containers of the devworkspace POD.
	// +optional
	PreStart []string `json:"preStart,omitempty"`

	// IDs of commands that should be executed after the devworkspace is completely started.
	// In the case of Che-Theia, these commands should be executed after all plugins and extensions have started, including project cloning.
	// This means that those commands are not triggered until the user opens the IDE in his browser.
	// +optional
	PostStart []string `json:"postStart,omitempty"`

	// +optional
	// IDs of commands that should be executed before stopping the devworkspace.
	PreStop []string `json:"preStop,omitempty"`

	// +optional
	// IDs of commands that should be executed after stopping the devworkspace.
	PostStop []string `json:"postStop,omitempty"`
}
