/*
Copyright 2019 The Knative Authors

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

package webhook

import (
	"strconv"
	"testing"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/metrics/metricstest"
	_ "knative.dev/pkg/metrics/testing"
)

func TestWebhookStatsReporter(t *testing.T) {
	setup()
	req := &admissionv1.AdmissionRequest{
		UID:       "705ab4f5-6393-11e8-b7cc-42010a800002",
		Kind:      metav1.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "Scale"},
		Resource:  metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
		Name:      "my-deployment",
		Namespace: "my-namespace",
		Operation: admissionv1.Update,
	}

	resp := &admissionv1.AdmissionResponse{
		UID:     req.UID,
		Allowed: true,
	}

	r, _ := NewStatsReporter()

	shortTime, longTime := 1100.0, 9100.0
	expectedTags := map[string]string{
		requestOperationKey.Name():  string(req.Operation),
		kindGroupKey.Name():         req.Kind.Group,
		kindVersionKey.Name():       req.Kind.Version,
		kindKindKey.Name():          req.Kind.Kind,
		resourceGroupKey.Name():     req.Resource.Group,
		resourceVersionKey.Name():   req.Resource.Version,
		resourceResourceKey.Name():  req.Resource.Resource,
		resourceNameKey.Name():      req.Name,
		resourceNamespaceKey.Name(): req.Namespace,
		admissionAllowedKey.Name():  strconv.FormatBool(resp.Allowed),
	}

	r.ReportRequest(req, resp, time.Duration(shortTime)*time.Millisecond)
	r.ReportRequest(req, resp, time.Duration(longTime)*time.Millisecond)

	metricstest.CheckCountData(t, requestCountName, expectedTags, 2)
	metricstest.CheckDistributionData(t, requestLatenciesName, expectedTags, 2, shortTime, longTime)
}

func setup() {
	resetMetrics()
}

// opencensus metrics carry global state that need to be reset between unit tests
func resetMetrics() {
	metricstest.Unregister(requestCountName, requestLatenciesName)
	RegisterMetrics()
}
