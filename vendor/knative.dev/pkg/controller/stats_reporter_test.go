/*
Copyright 2018 The Knative Authors

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

package controller

import (
	"strings"
	"testing"
	"time"

	"go.opencensus.io/stats/view"
	"knative.dev/pkg/metrics/metricstest"
	_ "knative.dev/pkg/metrics/testing"
)

func TestNewStatsReporterErrors(t *testing.T) {
	// These are invalid as defined by the current OpenCensus library.
	invalidTagValues := []string{
		"naïve",                  // Includes non-ASCII character.
		strings.Repeat("a", 256), // Longer than 255 characters.
	}

	for _, v := range invalidTagValues {
		_, err := NewStatsReporter(v)
		if err == nil {
			t.Errorf("Expected err to not be nil for value %q, got nil", v)
		}

	}
}

func TestReportQueueDepth(t *testing.T) {
	r1 := &reporter{}
	if err := r1.ReportQueueDepth(10); err == nil {
		t.Error("Reporter.Report() expected an error for Report call before init. Got success.")
	}

	r, _ := NewStatsReporter("testreconciler")
	wantTags := map[string]string{
		"reconciler": "testreconciler",
	}

	// Send statistics only once and observe the results
	expectSuccess(t, func() error { return r.ReportQueueDepth(10) })
	metricstest.CheckLastValueData(t, "work_queue_depth", wantTags, 10)

	// Queue depth stats is a gauge - record multiple entries - last one should stick
	expectSuccess(t, func() error { return r.ReportQueueDepth(1) })
	expectSuccess(t, func() error { return r.ReportQueueDepth(2) })
	expectSuccess(t, func() error { return r.ReportQueueDepth(3) })
	metricstest.CheckLastValueData(t, "work_queue_depth", wantTags, 3)
}

func TestReportReconcile(t *testing.T) {
	r, _ := NewStatsReporter("testreconciler")
	wantTags := map[string]string{
		"reconciler": "testreconciler",
		"key":        "test/key",
		"success":    "true",
	}

	initialReconcileCount := int64(0)
	if d, err := view.RetrieveData("reconcile_count"); err == nil && len(d) == 1 {
		initialReconcileCount = d[0].Data.(*view.CountData).Value
	}
	initialReconcileLatency := float64(0)
	if d, err := view.RetrieveData("reconcile_latency"); err == nil && len(d) == 1 {
		initialReconcileLatency = d[0].Data.(*view.DistributionData).Sum()
	}

	expectSuccess(t, func() error { return r.ReportReconcile(10*time.Millisecond, "test/key", "true") })
	metricstest.CheckCountData(t, "reconcile_count", wantTags, initialReconcileCount+1)
	metricstest.CheckDistributionData(t, "reconcile_latency", wantTags, 1, initialReconcileLatency+10, initialReconcileLatency+10)

	expectSuccess(t, func() error { return r.ReportReconcile(15*time.Millisecond, "test/key", "true") })
	metricstest.CheckCountData(t, "reconcile_count", wantTags, initialReconcileCount+2)
	metricstest.CheckDistributionData(t, "reconcile_latency", wantTags, 2, initialReconcileLatency+10, initialReconcileLatency+15)
}

func expectSuccess(t *testing.T, f func() error) {
	t.Helper()
	if err := f(); err != nil {
		t.Errorf("Reporter.Report() expected success but got error %v", err)
	}
}
