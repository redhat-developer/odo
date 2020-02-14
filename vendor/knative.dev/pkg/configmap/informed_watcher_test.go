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

package configmap

import (
	"sync"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
)

type counter struct {
	name string
	mu   sync.RWMutex
	cfg  []*corev1.ConfigMap
	wg   *sync.WaitGroup
}

func (c *counter) callback(cm *corev1.ConfigMap) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = append(c.cfg, cm)
	if c.wg != nil {
		c.wg.Done()
	}
}

func (c *counter) count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cfg)
}

func TestInformedWatcher(t *testing.T) {
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{"key": "val"},
	}
	barCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "bar",
		},
		Data: map[string]string{"key2": "val3"},
	}
	kc := fakekubeclientset.NewSimpleClientset(fooCM, barCM)
	cmw := NewInformedWatcher(kc, "default")

	foo1 := &counter{name: "foo1"}
	foo2 := &counter{name: "foo2"}
	bar := &counter{name: "bar"}
	cmw.Watch("foo", foo1.callback)
	cmw.Watch("foo", foo2.callback)
	cmw.Watch("bar", bar.callback)

	stopCh := make(chan struct{})
	defer close(stopCh)
	err := cmw.Start(stopCh)
	if err != nil {
		t.Fatalf("cm.Start() = %v", err)
	}

	// When Start returns the callbacks should have been called with the
	// version of the objects that is available.
	for _, obj := range []*counter{foo1, foo2, bar} {
		if got, want := obj.count(), 1; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After a "foo" event, the "foo" watchers should have 2,
	// and the "bar" watchers should still have 1
	cmw.updateConfigMapEvent(nil, fooCM)
	for _, obj := range []*counter{foo1, foo2} {
		if got, want := obj.count(), 2; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	for _, obj := range []*counter{bar} {
		if got, want := obj.count(), 1; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After a "foo" and "bar" event, the "foo" watchers should have 3,
	// and the "bar" watchers should still have 2
	cmw.updateConfigMapEvent(nil, fooCM)
	cmw.updateConfigMapEvent(nil, barCM)
	for _, obj := range []*counter{foo1, foo2} {
		if got, want := obj.count(), 3; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}
	for _, obj := range []*counter{bar} {
		if got, want := obj.count(), 2; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After a "bar" event, all watchers should have 3
	nbarCM := barCM.DeepCopy()
	nbarCM.Data["wow"] = "now!"
	cmw.updateConfigMapEvent(barCM, nbarCM)
	for _, obj := range []*counter{foo1, foo2, bar} {
		if got, want := obj.count(), 3; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After an idempotent event no changes should be recorded.
	cmw.updateConfigMapEvent(barCM, barCM)
	cmw.updateConfigMapEvent(fooCM, fooCM)
	for _, obj := range []*counter{foo1, foo2, bar} {
		if got, want := obj.count(), 3; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After an unwatched ConfigMap update, no change.

	cmw.updateConfigMapEvent(nil, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "not-watched",
		},
	})
	for _, obj := range []*counter{foo1, foo2, bar} {
		if got, want := obj.count(), 3; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}

	// After a change in an unrelated namespace, no change.
	cmw.updateConfigMapEvent(nil, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "not-default",
			Name:      "foo",
		},
	})
	for _, obj := range []*counter{foo1, foo2, bar} {
		if got, want := obj.count(), 3; got != want {
			t.Errorf("%v.count = %v, want %v", obj.name, got, want)
		}
	}
}

func TestFilterConfigByLabelExists(t *testing.T) {
	testCases := map[string]struct {
		input        string
		expectOutStr string
		expectErr    bool
	}{
		"Valid input": {
			input:        "test/label",
			expectOutStr: "test/label",
		},
		"Invalid input": {
			input:     "invalid,",
			expectErr: true,
		},
		"Empty input": {
			input:     "",
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req, err := FilterConfigByLabelExists(tc.input)

			if gotError := err != nil; gotError != tc.expectErr {
				t.Fatalf("[error] expected/got: %t/%t (msg: %v)", tc.expectErr, gotError, err)
			}
			if gotReq, wantReq := req != nil, tc.expectOutStr != ""; gotReq != wantReq {
				t.Fatalf("[output] expected/got: %t/%t (req: %q)", wantReq, gotReq, req)
			}

			if req != nil && req.String() != tc.expectOutStr {
				t.Errorf("Expected %q, got %q", tc.expectOutStr, req.String())
			}
		})
	}
}

func TestWatchMissingFailsOnStart(t *testing.T) {
	const (
		labelKey = "test/label"
		labelVal = "test"
	)

	cmWithLabel := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "with-label",
			Labels:    map[string]string{labelKey: labelVal},
		},
	}
	cmWithoutLabel := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "without-label",
			Labels:    map[string]string{},
		},
	}

	testCases := map[string]struct {
		initialObj []runtime.Object
		watchNames []string
		expectErr  string
		labelReq   string
	}{
		"ConfigMap does not exist": {
			initialObj: nil,
			watchNames: []string{"foo"},
			expectErr:  `configmap "foo" not found`,
		},
		"ConfigMap is missing required label": {
			initialObj: []runtime.Object{cmWithLabel, cmWithoutLabel},
			watchNames: []string{"with-label", "without-label"},
			expectErr:  `configmap "without-label" not found`,
			labelReq:   labelKey,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var cmw *InformedWatcher

			kc := fakekubeclientset.NewSimpleClientset(tc.initialObj...)

			if tc.labelReq != "" {
				req, _ := labels.NewRequirement(labelKey, selection.Equals, []string{labelVal})
				cmw = NewInformedWatcher(kc, "default", *req)
			} else {
				cmw = NewInformedWatcher(kc, "default")
			}

			for _, cmName := range tc.watchNames {
				cmw.Watch(cmName)
			}

			stopCh := make(chan struct{})
			defer close(stopCh)

			err := cmw.Start(stopCh)
			switch {
			case err == nil:
				t.Fatalf("Failed to start InformedWatcher: %s", err)
			case err.Error() != tc.expectErr:
				t.Fatalf("Unexpected error: %s", err)
			}
		})
	}
}

func TestWatchMissingOKWithDefaultOnStart(t *testing.T) {
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
	}

	kc := fakekubeclientset.NewSimpleClientset()
	cmw := NewInformedWatcher(kc, "default")

	foo1 := &counter{name: "foo1"}
	cmw.WatchWithDefault(*fooCM, foo1.callback)

	stopCh := make(chan struct{})
	defer close(stopCh)

	// This shouldn't error because we don't have a ConfigMap named "foo", but we do have a default.
	err := cmw.Start(stopCh)
	if err != nil {
		t.Fatalf("cm.Start() failed, %v", err)
	}

	if foo1.count() != 1 {
		t.Errorf("foo1.count = %v, want 1", foo1.count())
	}
}

func TestErrorOnMultipleStarts(t *testing.T) {
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
	}
	kc := fakekubeclientset.NewSimpleClientset(fooCM)
	cmw := NewInformedWatcher(kc, "default")

	foo1 := &counter{name: "foo1"}
	cmw.Watch("foo", foo1.callback)

	stopCh := make(chan struct{})
	defer close(stopCh)

	// This should succeed because the watched resource exists.
	if err := cmw.Start(stopCh); err != nil {
		t.Fatalf("cm.Start() = %v", err)
	}

	// This should error because we already called Start()
	if err := cmw.Start(stopCh); err == nil {
		t.Fatal("cm.Start() succeeded, wanted error")
	}
}

func TestDefaultObserved(t *testing.T) {
	defaultFooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"default": "from code",
		},
	}
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"from": "k8s",
		},
	}

	kc := fakekubeclientset.NewSimpleClientset(fooCM)
	cmw := NewInformedWatcher(kc, "default")

	foo1 := &counter{name: "foo1"}
	cmw.WatchWithDefault(*defaultFooCM, foo1.callback)

	stopCh := make(chan struct{})
	defer close(stopCh)
	err := cmw.Start(stopCh)
	if err != nil {
		t.Fatalf("cm.Start() = %v", err)
	}
	// We expect:
	// 1. The default to be seen once during startup.
	// 2. The real K8s version during the initial pass.
	expected := []*corev1.ConfigMap{defaultFooCM, fooCM}
	if got, want := foo1.count(), len(expected); got != want {
		t.Fatalf("foo1.count = %v, want %d", got, want)
	}
	for i, cfg := range expected {
		if got, want := foo1.cfg[i].Data, cfg.Data; !equality.Semantic.DeepEqual(want, got) {
			t.Errorf("%d config seen should have been '%v', actually '%v'", i, want, got)
		}
	}
}

func TestDefaultConfigMapDeleted(t *testing.T) {
	defaultFooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"default": "from code",
		},
	}
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"from": "k8s",
		},
	}

	kc := fakekubeclientset.NewSimpleClientset(fooCM)
	cmw := NewInformedWatcher(kc, "default")

	foo1 := &counter{name: "foo1"}
	cmw.WatchWithDefault(*defaultFooCM, foo1.callback)

	stopCh := make(chan struct{})
	defer close(stopCh)
	err := cmw.Start(stopCh)
	if err != nil {
		t.Fatalf("cm.Start() = %v", err)
	}

	// Delete the real ConfigMap in K8s, which should cause the default to be processed again.
	// Because this happens asynchronously via a watcher, use a sync.WaitGroup to wait until it has
	// occurred.
	foo1.mu.Lock()
	foo1.wg = &sync.WaitGroup{}
	foo1.mu.Unlock()
	foo1.wg.Add(1)
	err = kc.CoreV1().ConfigMaps(fooCM.Namespace).Delete(fooCM.Name, nil)
	if err != nil {
		t.Fatalf("Error deleting fooCM: %v", err)
	}
	foo1.wg.Wait()

	// We expect:
	// 1. The default to be seen once during startup.
	// 2. The real K8s version during the initial pass.
	// 3. The default again, when the real K8s version is deleted.
	expected := []*corev1.ConfigMap{defaultFooCM, fooCM, defaultFooCM}
	if got, want := foo1.count(), len(expected); got != want {
		t.Fatalf("foo1.count = %v, want %d", got, want)
	}
	for i, cfg := range expected {
		if got, want := foo1.cfg[i].Data, cfg.Data; !equality.Semantic.DeepEqual(want, got) {
			t.Errorf("%d config seen should have been '%v', actually '%v'", i, want, got)
		}
	}
}

func TestWatchWithDefaultAfterStart(t *testing.T) {
	defaultFooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"default": "from code",
		},
	}
	fooCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "foo",
		},
		Data: map[string]string{
			"from": "k8s",
		},
	}

	kc := fakekubeclientset.NewSimpleClientset(fooCM)
	cmw := NewInformedWatcher(kc, "default")

	stopCh := make(chan struct{})
	defer close(stopCh)
	// Start before adding the WatchWithDefault.
	err := cmw.Start(stopCh)
	if err != nil {
		t.Fatalf("cm.Start() = %v", err)
	}

	foo1 := &counter{name: "foo1"}

	// Add the WatchWithDefault. This should panic because the InformedWatcher has already started.
	func() {
		defer func() {
			recover()
		}()
		cmw.WatchWithDefault(*defaultFooCM, foo1.callback)
		t.Fatal("WatchWithDefault should have panicked")
	}()

	// We expect nothing.
	var expected []*corev1.ConfigMap
	if got, want := foo1.count(), len(expected); got != want {
		t.Fatalf("foo1.count = %v, want %d", got, want)
	}
}
