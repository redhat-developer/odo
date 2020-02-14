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

package metrics

import (
	"testing"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"k8s.io/client-go/util/workqueue"

	"knative.dev/pkg/metrics/metricstest"
)

func newInt64(name string) *stats.Int64Measure {
	return stats.Int64(name, "bar", "wtfs/s")
}

func newFloat64(name string) *stats.Float64Measure {
	return stats.Float64(name, "bar", "wtfs/s")
}

func TestWorkqueueMetrics(t *testing.T) {
	wp := &WorkqueueProvider{
		Adds:                           newInt64("adds"),
		Depth:                          newInt64("depth"),
		Latency:                        newFloat64("latency"),
		Retries:                        newInt64("retries"),
		WorkDuration:                   newFloat64("work_duration"),
		UnfinishedWorkSeconds:          newFloat64("unfinished_work_seconds"),
		LongestRunningProcessorSeconds: newFloat64("longest_running_processor_seconds"),
	}
	workqueue.SetProvider(wp)

	// Reset the metrics configuration to avoid leaked state from other tests.
	setCurMetricsConfig(nil)

	views := wp.DefaultViews()
	if got, want := len(views), 7; got != want {
		t.Errorf("len(DefaultViews()) = %d, want %d", got, want)
	}
	if err := view.Register(views...); err != nil {
		t.Errorf("view.Register() = %v", err)
	}
	defer view.Unregister(views...)

	queueName := t.Name()
	wq := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		queueName,
	)

	metricstest.CheckStatsNotReported(t, "adds", "depth", "latency", "retries", "work_duration",
		"unfinished_work_seconds", "longest_running_processor_seconds")

	wq.Add("foo")

	metricstest.CheckStatsReported(t, "adds", "depth")
	metricstest.CheckStatsNotReported(t, "latency", "retries", "work_duration",
		"unfinished_work_seconds", "longest_running_processor_seconds")
	metricstest.CheckCountData(t, "adds", map[string]string{"name": queueName}, 1)
	metricstest.CheckLastValueData(t, "depth", map[string]string{"name": queueName}, 1)

	wq.Add("bar")

	metricstest.CheckStatsNotReported(t, "latency", "retries", "work_duration",
		"unfinished_work_seconds", "longest_running_processor_seconds")
	metricstest.CheckCountData(t, "adds", map[string]string{"name": queueName}, 2)
	metricstest.CheckLastValueData(t, "depth", map[string]string{"name": queueName}, 2)

	if got, shutdown := wq.Get(); shutdown {
		t.Errorf("Get() = %v, true; want false", got)
	} else if want := "foo"; got != want {
		t.Errorf("Get() = %s, false; want %s", got, want)
	} else {
		wq.Forget(got)
		wq.Done(got)
	}

	metricstest.CheckStatsReported(t, "latency", "work_duration")
	metricstest.CheckStatsNotReported(t, "retries",
		"unfinished_work_seconds", "longest_running_processor_seconds")
	metricstest.CheckCountData(t, "adds", map[string]string{"name": queueName}, 2)

	if got, shutdown := wq.Get(); shutdown {
		t.Errorf("Get() = %v, true; want false", got)
	} else if want := "bar"; got != want {
		t.Errorf("Get() = %s, false; want %s", got, want)
	} else {
		wq.AddRateLimited(got)
		wq.Done(got)
	}

	// It should show up as a retry now.
	metricstest.CheckStatsReported(t, "retries")
	metricstest.CheckStatsNotReported(t, "unfinished_work_seconds", "longest_running_processor_seconds")
	metricstest.CheckCountData(t, "retries", map[string]string{"name": queueName}, 1)
	// It is not added right away.
	metricstest.CheckCountData(t, "adds", map[string]string{"name": queueName}, 2)

	// It doesn't show up as an "add" until the rate limit has elapsed.
	time.Sleep(1 * time.Second)
	metricstest.CheckCountData(t, "adds", map[string]string{"name": queueName}, 3)

	wq.ShutDown()

	if got, shutdown := wq.Get(); shutdown {
		t.Errorf("Get() = %v, true; want false", got)
	} else if want := "bar"; got != want {
		t.Errorf("Get() = %s, true; want %s", got, want)
	}

	if got, shutdown := wq.Get(); !shutdown {
		t.Errorf("Get() = %v, false; want true", got)
	}
}
