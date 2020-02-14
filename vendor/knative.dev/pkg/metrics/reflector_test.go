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

	"go.opencensus.io/stats/view"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/metrics/metricstest"
)

func TestReflectorMetrics(t *testing.T) {
	rp := &ReflectorProvider{
		ItemsInList:         newFloat64("items_in_list"),
		ItemsInMatch:        newFloat64("items_in_match"),
		ItemsInWatch:        newFloat64("items_in_watch"),
		LastResourceVersion: newFloat64("last_resource_version"),
		ListDuration:        newFloat64("list_duration"),
		Lists:               newInt64("lists"),
		ShortWatches:        newInt64("short_watches"),
		WatchDuration:       newFloat64("watch_duration"),
		Watches:             newInt64("watches"),
	}
	cache.SetReflectorMetricsProvider(rp)

	// Reset the metrics configuration to avoid leaked state from other tests.
	setCurMetricsConfig(nil)

	views := rp.DefaultViews()
	if got, want := len(views), 9; got != want {
		t.Errorf("len(DefaultViews()) = %d, want %d", got, want)
	}
	if err := view.Register(views...); err != nil {
		t.Errorf("view.Register() = %v", err)
	}
	defer view.Unregister(views...)

	metricstest.CheckStatsNotReported(t, "items_in_list", "items_in_match", "items_in_watch",
		"last_resource_version", "list_duration", "lists", "short_watches", "watch_duration", "watches")

	stopCh := make(chan struct{})
	defer close(stopCh)

	fake := kubefake.NewSimpleClientset()
	factory := kubeinformers.NewSharedInformerFactory(fake, 0)
	endpoints := factory.Core().V1().Endpoints()

	informer := endpoints.Informer()
	go informer.Run(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, informer.HasSynced); !ok {
		t.Error("failed to wait for endpoints cache to sync")
	}

	// TODO(mattmoor): The reflector metrics don't seem to be hooked up to anything,
	// so nothing is reported...  (－‸ლ)
	metricstest.CheckStatsNotReported(t, "items_in_list", "items_in_match", "items_in_watch",
		"last_resource_version", "list_duration", "lists", "short_watches", "watch_duration", "watches")
}
