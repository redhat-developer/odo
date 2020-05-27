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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func makeResource(topLevelCond string) *duckv1.KResource {
	fooCond := apis.Condition{
		Type:    "Foo",
		Status:  corev1.ConditionTrue,
		Message: "Something something foo",
	}
	readyCond := apis.Condition{
		Type:    apis.ConditionType(topLevelCond),
		Status:  corev1.ConditionTrue,
		Message: "Something something bar",
	}

	return &duckv1.KResource{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 42,
		},

		Status: duckv1.Status{
			ObservedGeneration: 0,
			Conditions:         duckv1.Conditions{fooCond, readyCond},
		},
	}
}

func TestPreProcessResetsReady(t *testing.T) {
	testCases := []struct {
		name                      string
		initTopLevelCond          string
		expectedTopLevelCondition apis.ConditionType
	}{{
		name:                      "top level Ready",
		initTopLevelCond:          "Ready",
		expectedTopLevelCondition: apis.ConditionReady,
	}, {
		name:                      "top level Succeeded",
		initTopLevelCond:          "Succeeded",
		expectedTopLevelCondition: apis.ConditionSucceeded,
	}}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			resource := makeResource(test.initTopLevelCond)
			krShape := duckv1.KRShaped(resource)

			PreProcessReconcile(context.Background(), krShape)

			if rc := resource.Status.GetCondition(test.expectedTopLevelCondition); rc.Status != "Unknown" {
				t.Errorf("Expected unchanged ready status got=%s want=Unknown", rc.Status)
			}
		})
	}
}

func TestPostProcessReconcileBumpsGeneration(t *testing.T) {
	resource := makeResource("Ready")

	krShape := duckv1.KRShaped(resource)
	PostProcessReconcile(context.Background(), krShape)

	if resource.Status.ObservedGeneration != resource.Generation {
		t.Errorf("Expected observed generation bump got=%d want=%d", resource.Status.ObservedGeneration, resource.Generation)
	}

	if krShape.GetStatus().ObservedGeneration != krShape.GetGeneration() {
		t.Errorf("Expected observed generation bump got=%d want=%d", resource.Status.ObservedGeneration, resource.Generation)
	}
}
