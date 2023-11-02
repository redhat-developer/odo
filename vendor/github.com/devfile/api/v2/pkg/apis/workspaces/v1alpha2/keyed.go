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

// Keyed is expected to be implemented by the elements of the devfile top-level lists
// (such as Command, Component, Project, ...).
//
// The Keys of list objects will typically be used to merge the top-level lists
// according to strategic merge patch rules, during parent or plugin overriding.
// +k8s:deepcopy-gen=false
type Keyed interface {
	// Key is a string that allows uniquely identifying the object,
	// especially in the Devfile top-level lists that are map-like K8S-compatible lists.
	Key() string
}

// KeyedList is a list of object that are uniquely identified by a Key
// The devfile top-level list (such as Commands, Components, Projects, ...)
// are examples of such lists of Keyed objects
// +k8s:deepcopy-gen=false
type KeyedList []Keyed

// GetKeys converts a KeyedList into a slice of string by calling Key() on each
// element in the list.
func (l KeyedList) GetKeys() []string {
	var res []string
	for _, keyed := range l {
		res = append(res, keyed.Key())
	}
	return res
}

// TopLevelLists is a map that contains several Devfile top-level lists
// (such as `Commands`, `Components`, `Projects`, ...), available as `KeyedList`s.
//
// Each key of this map is the name of the field that contains the given top-level list:
// `Commands`, `Components`, etc...
// +k8s:deepcopy-gen=false
type TopLevelLists map[string]KeyedList

// TopLevelListContainer is an interface that allows retrieving the devfile top-level lists
// from an object.
// Main implementor of this interface will be the `DevWorkspaceTemplateSpecContent`, which
// will return all its devfile top-level lists.
//
// But this will also be implemented by `Overrides` which may return less top-level lists
// the `DevWorkspaceTemplateSpecContent`, according to the top-level lists they can override.
// `PluginOverride` will not return the `Projects` and `StarterProjects` list, since plugins are
// not expected to override `projects` or `starterProjects`
// +k8s:deepcopy-gen=false
type TopLevelListContainer interface {
	GetToplevelLists() TopLevelLists
}
