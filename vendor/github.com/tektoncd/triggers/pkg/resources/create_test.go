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

package resources

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	resourcev1 "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	dynamicclientset "github.com/tektoncd/triggers/pkg/client/dynamic/clientset"
	"github.com/tektoncd/triggers/pkg/client/dynamic/clientset/tekton"
	"github.com/tektoncd/triggers/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"knative.dev/pkg/logging"
)

const (
	resourceLabel = triggersv1.GroupName + triggersv1.EventListenerLabelKey
	triggerLabel  = triggersv1.GroupName + triggersv1.TriggerLabelKey
	eventIDLabel  = triggersv1.GroupName + triggersv1.EventIDLabelKey

	triggerName = "trigger"
	eventID     = "12345"
)

func Test_FindAPIResource_error(t *testing.T) {
	dc := fakekubeclientset.NewSimpleClientset().Discovery()
	if _, err := FindAPIResource("v1", "Pod", dc); err == nil {
		t.Error("findAPIResource() did not return error when expected")
	}
}

func TestFindAPIResource(t *testing.T) {
	// Create fake kubeclient with list of resources
	kubeClient := fakekubeclientset.NewSimpleClientset()
	kubeClient.Resources = []*metav1.APIResourceList{{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{{
			Name:       "pods",
			Namespaced: true,
			Kind:       "Pod",
		}, {
			Name:       "namespaces",
			Namespaced: false,
			Kind:       "Namespace",
		}},
	}}
	test.AddTektonResources(kubeClient)
	dc := kubeClient.Discovery()

	tests := []struct {
		apiVersion string
		kind       string
		want       *metav1.APIResource
	}{{
		apiVersion: "v1",
		kind:       "Pod",
		want: &metav1.APIResource{
			Name:       "pods",
			Namespaced: true,
			Version:    "v1",
			Kind:       "Pod",
		},
	}, {
		apiVersion: "v1",
		kind:       "Namespace",
		want: &metav1.APIResource{
			Name:       "namespaces",
			Namespaced: false,
			Version:    "v1",
			Kind:       "Namespace",
		},
	}, {
		apiVersion: "tekton.dev/v1alpha1",
		kind:       "TriggerTemplate",
		want: &metav1.APIResource{
			Group:      "tekton.dev",
			Version:    "v1alpha1",
			Name:       "triggertemplates",
			Namespaced: true,
			Kind:       "TriggerTemplate",
		},
	}, {
		apiVersion: "tekton.dev/v1alpha1",
		kind:       "PipelineRun",
		want: &metav1.APIResource{
			Group:      "tekton.dev",
			Version:    "v1alpha1",
			Name:       "pipelineruns",
			Namespaced: true,
			Kind:       "PipelineRun",
		},
	},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.apiVersion, tt.kind), func(t *testing.T) {
			got, err := FindAPIResource(tt.apiVersion, tt.kind, dc)
			if err != nil {
				t.Errorf("findAPIResource() returned error: %s", err)
			} else if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("findAPIResource() Diff: -want +got: %s", diff)
			}
		})
	}
}

func TestCreateResource(t *testing.T) {
	elName := "foo-el"
	elNamespace := "bar"

	kubeClient := fakekubeclientset.NewSimpleClientset()
	test.AddTektonResources(kubeClient)

	dynamicClient := fakedynamic.NewSimpleDynamicClient(runtime.NewScheme())
	dynamicSet := dynamicclientset.New(tekton.WithClient(dynamicClient))

	logger, _ := logging.NewLogger("", "")

	tests := []struct {
		name string
		json []byte
		want resourcev1.PipelineResource
	}{{
		name: "PipelineResource without namespace",
		json: json.RawMessage(`{"kind":"PipelineResource","apiVersion":"tekton.dev/v1alpha1","metadata":{"name":"my-pipelineresource","creationTimestamp":null,"labels":{"woriginal-label-1":"label-1"}},"spec":{"type":"","params":[{"name":"foo","value":"bar\r\nbaz"}]},"status":{}}`),
		want: resourcev1.PipelineResource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "tekton.dev/v1alpha1",
				Kind:       "PipelineResource",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-pipelineresource",
				Labels: map[string]string{
					"woriginal-label-1": "label-1",
					resourceLabel:       elName,
					triggerLabel:        triggerName,
					eventIDLabel:        eventID,
				},
			},
			Spec: resourcev1.PipelineResourceSpec{
				Params: []resourcev1.ResourceParam{{
					Name:  "foo",
					Value: "bar\r\nbaz",
				}},
			},
			Status: &resourcev1.PipelineResourceStatus{},
		},
	}, {
		name: "PipelineResource with namespace",
		json: json.RawMessage(`{"kind":"PipelineResource","apiVersion":"tekton.dev/v1alpha1","metadata":{"name":"my-pipelineresource","namespace":"foo","creationTimestamp":null,"labels":{"woriginal-label-1":"label-1"}},"spec":{"type":"","params":null},"status":{}}`),
		want: resourcev1.PipelineResource{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "tekton.dev/v1alpha1",
				Kind:       "PipelineResource",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "foo",
				Name:      "my-pipelineresource",
				Labels: map[string]string{
					"woriginal-label-1": "label-1",
					resourceLabel:       elName,
					triggerLabel:        triggerName,
					eventIDLabel:        eventID,
				},
			},
			Spec:   resourcev1.PipelineResourceSpec{},
			Status: &resourcev1.PipelineResourceStatus{},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dynamicClient.ClearActions()
			if err := Create(logger, tt.json, triggerName, eventID, elName, elNamespace, kubeClient.Discovery(), dynamicSet); err != nil {
				t.Errorf("createResource() returned error: %s", err)
			}

			gvr := schema.GroupVersionResource{
				Group:    "tekton.dev",
				Version:  "v1alpha1",
				Resource: "pipelineresources",
			}
			namespace := tt.want.Namespace
			if namespace == "" {
				namespace = elNamespace
			}
			want := []ktesting.Action{ktesting.NewCreateAction(gvr, namespace, test.ToUnstructured(t, tt.want))}
			if diff := cmp.Diff(want, dynamicClient.Actions()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_AddLabels(t *testing.T) {
	tests := []struct {
		name        string
		us          *unstructured.Unstructured
		labelsToAdd map[string]string
		want        *unstructured.Unstructured
	}{
		{
			name: "add to empty labels",
			us: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{},
				}},
			labelsToAdd: map[string]string{"foo": "bar"},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"triggers.tekton.dev/foo": "bar",
						},
					},
				},
			},
		},
		{
			name: "overwrite label",
			us: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"triggers.tekton.dev/foo": "bar",
						},
					},
				}},
			labelsToAdd: map[string]string{"foo": "foo"},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"triggers.tekton.dev/foo": "foo",
						},
					},
				},
			},
		},
		{
			name: "add and overwrite labels",
			us: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							// should be overwritten
							"triggers.tekton.dev/foo":   "bar",
							"triggers.tekton.dev/hello": "world",
							// should be preserved
							"triggers.tekton.dev/z": "0",
							"best-palindrome":       "tacocat",
						},
					},
				}},
			labelsToAdd: map[string]string{
				"foo":   "foo",
				"hello": "there",
				"a":     "a",
				"b":     "b",
			},
			want: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"triggers.tekton.dev/foo":   "foo",
							"triggers.tekton.dev/hello": "there",
							"triggers.tekton.dev/z":     "0",
							"best-palindrome":           "tacocat",
							"triggers.tekton.dev/a":     "a",
							"triggers.tekton.dev/b":     "b",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddLabels(tt.us, tt.labelsToAdd)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AddLabels(): -want +got: %s", diff)
			}
		})
	}
}
