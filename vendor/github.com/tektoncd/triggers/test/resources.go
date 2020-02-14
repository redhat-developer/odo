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

package test

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
)

// ToUnstructured returns an Unstructured object from interface in.
func ToUnstructured(t *testing.T, in interface{}) *unstructured.Unstructured {
	t.Helper()

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("error encoding to JSON: %v", err)
	}

	out := new(unstructured.Unstructured)
	if err := out.UnmarshalJSON(b); err != nil {
		t.Fatalf("error encoding to unstructured: %v", err)
	}
	return out
}

// AddTektonResources will update clientset to know it knows about the types it is
// expected to be able to interact with.
func AddTektonResources(clientset *fakekubeclientset.Clientset) {
	nameKind := map[string]string{
		"triggertemplates":  "TriggerTemplate",
		"pipelineruns":      "PipelineRun",
		"taskruns":          "TaskRun",
		"pipelineresources": "PipelineResource",
	}
	resources := make([]metav1.APIResource, 0, len(nameKind))
	for name, kind := range nameKind {
		resources = append(resources, metav1.APIResource{
			Group:      "tekton.dev",
			Version:    "v1alpha1",
			Namespaced: true,
			Name:       name,
			Kind:       kind,
		})
	}

	clientset.Resources = append(clientset.Resources, &metav1.APIResourceList{
		GroupVersion: "tekton.dev/v1alpha1",
		APIResources: resources,
	})
}
