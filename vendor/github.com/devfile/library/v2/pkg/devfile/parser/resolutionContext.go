//
// Copyright 2022 Red Hat, Inc.
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

package parser

import (
	"fmt"
	"reflect"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// resolutionContextTree is a recursive structure representing information about the devfile that is
// lost when flattening (e.g. plugins, parents)
type resolutionContextTree struct {
	importReference v1.ImportReference
	parentNode      *resolutionContextTree
}

// appendNode adds a new node to the resolution context.
func (t *resolutionContextTree) appendNode(importReference v1.ImportReference) *resolutionContextTree {
	newNode := &resolutionContextTree{
		importReference: importReference,
		parentNode:      t,
	}
	return newNode
}

// hasCycle checks if the current resolutionContextTree has a cycle
func (t *resolutionContextTree) hasCycle() error {
	var seenRefs []v1.ImportReference
	currNode := t
	hasCycle := false
	cycle := resolveImportReference(t.importReference)

	for currNode.parentNode != nil {
		for _, seenRef := range seenRefs {
			if reflect.DeepEqual(seenRef, currNode.importReference) {
				hasCycle = true
			}
		}
		seenRefs = append(seenRefs, currNode.importReference)
		currNode = currNode.parentNode
		cycle = fmt.Sprintf("%s -> %s", resolveImportReference(currNode.importReference), cycle)
	}

	if hasCycle {
		return fmt.Errorf("devfile has an cycle in references: %v", cycle)
	}
	return nil
}

func resolveImportReference(importReference v1.ImportReference) string {
	if !reflect.DeepEqual(importReference, v1.ImportReference{}) {
		switch {
		case importReference.Uri != "":
			return fmt.Sprintf("uri: %s", importReference.Uri)
		case importReference.Id != "":
			return fmt.Sprintf("id: %s, registryURL: %s", importReference.Id, importReference.RegistryUrl)
		case importReference.Kubernetes != nil:
			return fmt.Sprintf("name: %s, namespace: %s", importReference.Kubernetes.Name, importReference.Kubernetes.Namespace)
		}

	}
	// the first node
	return "main devfile"
}
