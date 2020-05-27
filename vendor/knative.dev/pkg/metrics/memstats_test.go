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
	"context"
	"runtime"
	"testing"
	"time"

	"go.opencensus.io/stats/view"

	"knative.dev/pkg/metrics/metricstest"
)

func TestMemStatsMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	period := 200 * time.Millisecond

	msp := NewMemStatsAll()
	msp.Start(ctx, period)

	// Reset the metrics configuration to avoid leaked state from other tests.
	InitForTesting()

	views := msp.DefaultViews()
	if got, want := len(views), 27; got != want {
		t.Errorf("len(DefaultViews()) = %d, want %d", got, want)
	}
	if err := view.Register(views...); err != nil {
		t.Errorf("view.Register() = %v", err)
	}
	defer view.Unregister(views...)

	metricstest.CheckStatsNotReported(t,
		"go_alloc",
		"go_total_alloc",
		"go_sys",
		"go_lookups",
		"go_mallocs",
		"go_frees",
		"go_heap_alloc",
		"go_heap_sys",
		"go_heap_idle",
		"go_heap_in_use",
		"go_heap_released",
		"go_heap_objects",
		"go_stack_in_use",
		"go_stack_sys",
		"go_mspan_in_use",
		"go_mspan_sys",
		"go_mcache_in_use",
		"go_mcache_sys",
		"go_bucket_hash_sys",
		"go_gc_sys",
		"go_other_sys",
		"go_next_gc",
		"go_last_gc",
		"go_total_gc_pause_ns",
		"go_num_gc",
		"go_num_forced_gc",
		"go_gc_cpu_fraction",
	)

	time.Sleep(period + 100*time.Millisecond)

	metricstest.CheckStatsReported(t,
		"go_alloc",
		"go_total_alloc",
		"go_sys",
		"go_lookups",
		"go_mallocs",
		"go_frees",
		"go_heap_alloc",
		"go_heap_sys",
		"go_heap_idle",
		"go_heap_in_use",
		"go_heap_released",
		"go_heap_objects",
		"go_stack_in_use",
		"go_stack_sys",
		"go_mspan_in_use",
		"go_mspan_sys",
		"go_mcache_in_use",
		"go_mcache_sys",
		"go_bucket_hash_sys",
		"go_gc_sys",
		"go_other_sys",
		"go_next_gc",
		"go_last_gc",
		"go_total_gc_pause_ns",
		"go_num_gc",
		"go_num_forced_gc",
		"go_gc_cpu_fraction",
	)

	// We have seen zero forced GCs.
	metricstest.CheckLastValueData(t, "go_num_forced_gc", map[string]string{}, 0)

	// Force a GC, and wait for the reporting period.
	runtime.GC()
	time.Sleep(period + 100*time.Millisecond)

	// Now we should report a single forced GC.
	metricstest.CheckLastValueData(t, "go_num_forced_gc", map[string]string{}, 1)

	// Repeat, and we should see two.
	runtime.GC()
	time.Sleep(period + 100*time.Millisecond)
	metricstest.CheckLastValueData(t, "go_num_forced_gc", map[string]string{}, 2)

	// After we cancel the context, it should kill the go routine and any additional GCs
	// should not be reported.
	cancel()
	runtime.GC()
	time.Sleep(period + 100*time.Millisecond)
	metricstest.CheckLastValueData(t, "go_num_forced_gc", map[string]string{}, 2)
}
