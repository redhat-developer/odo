/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconciler

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type TestResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status duckv1.Status `json:"status"`
}

func (t *TestResource) SetDefaults(context.Context) {
	t.Annotations = map[string]string{"default": "was set"}
}

func (t *TestResource) GetStatus() *duckv1.Status {
	return &t.Status
}

func (*TestResource) GetConditionSet() apis.ConditionSet {
	return apis.NewLivingConditionSet("Foo", "Bar")
}

func makeResource() *TestResource {
	fooCond := apis.Condition{
		Type:    "Foo",
		Status:  corev1.ConditionTrue,
		Message: "Something something foo",
	}
	readyCond := apis.Condition{
		Type:    apis.ConditionReady,
		Status:  corev1.ConditionTrue,
		Message: "Something something bar",
	}

	return &TestResource{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 42,
		},

		Status: duckv1.Status{
			ObservedGeneration: 0,
			Conditions:         duckv1.Conditions{fooCond, readyCond},
		},
	}
}

func TestPreProcess(t *testing.T) {
	resource := makeResource()
	krShape := duckv1.KRShaped(resource)

	PreProcessReconcile(context.Background(), krShape)

	if resource.Annotations["default"] != "was set" {
		t.Errorf("Expected default annotations set got=%v", resource.Annotations)
	}

	if rc := resource.Status.GetCondition("Ready"); rc.Status != "Unknown" {
		t.Errorf("Expected unchanged ready status got=%s want=Unknown", rc.Status)
	}

	// Ensure Foo is untouched
	if rc := resource.Status.GetCondition("Foo"); rc.Status != "True" {
		t.Errorf("Expected dependant condition to remain got=%s want=True", rc.Status)
	}

	// Ensure Bar is initialized
	if rc := resource.Status.GetCondition("Bar"); rc.Status != "True" {
		t.Errorf("Expected conditions to be initialized got=%s want=True", rc.Status)
	}
}

func TestPostProcessReconcileBumpsGeneration(t *testing.T) {
	resource := makeResource()

	krShape := duckv1.KRShaped(resource)
	PostProcessReconcile(context.Background(), krShape, krShape)

	if resource.Status.ObservedGeneration != resource.Generation {
		t.Errorf("Expected observed generation bump got=%d want=%d",
			resource.Status.ObservedGeneration, resource.Generation)
	}

	if krShape.GetStatus().ObservedGeneration != krShape.GetGeneration() {
		t.Errorf("Expected observed generation bump got=%d want=%d",
			resource.Status.ObservedGeneration, resource.Generation)
	}
}

func TestPostProcessReconcileUpdatesTransitionTimes(t *testing.T) {
	oldNow := apis.VolatileTime{Inner: metav1.NewTime(time.Now())}
	resource := makeResource()
	oldResource := makeResource()
	// initialize old conditions with oldNow
	oldResource.Status.Conditions[0].LastTransitionTime = oldNow
	oldResource.Status.Conditions[1].LastTransitionTime = oldNow
	// change the second condition, but keep the old timestamp.
	resource.Status.Conditions[1].LastTransitionTime = oldNow
	resource.Status.Conditions[1].Status = corev1.ConditionFalse

	new := duckv1.KRShaped(resource)
	old := duckv1.KRShaped(oldResource)
	PostProcessReconcile(context.Background(), new, old)

	unchangedCond := resource.Status.Conditions[0]
	if unchangedCond.LastTransitionTime != oldNow {
		t.Errorf("Expected unchanged condition to keep old timestamp. Got=%v Want=%v",
			unchangedCond.LastTransitionTime, oldNow)
	}

	changedCond := resource.Status.Conditions[1]
	if changedCond.LastTransitionTime == oldNow {
		t.Errorf("Expected changed condition to get a new timestamp. Got=%v Want=%v",
			changedCond.LastTransitionTime, oldNow)
	}
}
