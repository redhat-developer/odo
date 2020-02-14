/*
Copyright 2017 The Knative Authors

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
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	. "knative.dev/pkg/controller/testing"
	. "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/testing"
)

func TestPassNew(t *testing.T) {
	old := "foo"
	new := "bar"

	PassNew(func(got interface{}) {
		if new != got.(string) {
			t.Errorf("PassNew() = %v, wanted %v", got, new)
		}
	})(old, new)
}

func TestHandleAll(t *testing.T) {
	old := "foo"
	new := "bar"

	ha := HandleAll(func(got interface{}) {
		if new != got.(string) {
			t.Errorf("HandleAll() = %v, wanted %v", got, new)
		}
	})

	ha.OnAdd(new)
	ha.OnUpdate(old, new)
	ha.OnDelete(new)
}

var (
	boolTrue  = true
	boolFalse = false
	gvk       = schema.GroupVersionKind{
		Group:   "pkg.knative.dev",
		Version: "v1meta1",
		Kind:    "Parent",
	}
)

func TestFilterWithNameAndNamespace(t *testing.T) {
	filter := FilterWithNameAndNamespace("test-namespace", "test-name")

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{{
		name:  "not a metav1.Object",
		input: "foo",
		want:  false,
	}, {
		name:  "nil",
		input: nil,
		want:  false,
	}, {
		name: "name matches, namespace does not",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "wrong-namespace",
			},
		},
		want: false,
	}, {
		name: "namespace matches, name does not",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "wrong-name",
				Namespace: "test-namespace",
			},
		},
		want: false,
	}, {
		name: "neither matches",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "wrong-name",
				Namespace: "wrong-namespace",
			},
		},
		want: false,
	}, {
		name: "matches",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		},
		want: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter(test.input)
			if test.want != got {
				t.Errorf("FilterWithNameAndNamespace() = %v, wanted %v", got, test.want)
			}
		})
	}
}

func TestFilterWithName(t *testing.T) {
	filter := FilterWithName("test-name")

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{{
		name:  "not a metav1.Object",
		input: "foo",
		want:  false,
	}, {
		name:  "nil",
		input: nil,
		want:  false,
	}, {
		name: "name matches, namespace does not",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "wrong-namespace",
			},
		},
		want: true, // Unlike FilterWithNameAndNamespace this passes
	}, {
		name: "namespace matches, name does not",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "wrong-name",
				Namespace: "test-namespace",
			},
		},
		want: false,
	}, {
		name: "neither matches",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "wrong-name",
				Namespace: "wrong-namespace",
			},
		},
		want: false,
	}, {
		name: "matches",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		},
		want: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter(test.input)
			if test.want != got {
				t.Errorf("FilterWithNameAndNamespace() = %v, wanted %v", got, test.want)
			}
		})
	}
}

func TestFilterGroupKind(t *testing.T) {
	filter := FilterGroupKind(gvk.GroupKind())

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{{
		name:  "not a metav1.Object",
		input: "foo",
		want:  false,
	}, {
		name:  "nil",
		input: nil,
		want:  false,
	}, {
		name: "no owner reference",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		},
		want: false,
	}, {
		name: "wrong owner reference, not controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "another.knative.dev/v1beta3",
					Kind:       "Parent",
					Controller: &boolFalse,
				}},
			},
		},
		want: false,
	}, {
		name: "right owner reference, not controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Controller: &boolFalse,
				}},
			},
		},
		want: false,
	}, {
		name: "wrong owner reference, but controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "another.knative.dev/v1beta3",
					Kind:       "Parent",
					Controller: &boolTrue,
				}},
			},
		},
		want: false,
	}, {
		name: "right owner reference, is controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Controller: &boolTrue,
				}},
			},
		},
		want: true,
	}, {
		name: "right owner reference, is controller, different version",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: schema.GroupVersion{Group: gvk.Group, Version: "other"}.String(),
					Kind:       gvk.Kind,
					Controller: &boolTrue,
				}},
			},
		},
		want: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter(test.input)
			if test.want != got {
				t.Errorf("Filter() = %v, wanted %v", got, test.want)
			}
		})
	}
}

func TestFilterGroupVersionKind(t *testing.T) {
	filter := FilterGroupVersionKind(gvk)

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{{
		name:  "not a metav1.Object",
		input: "foo",
		want:  false,
	}, {
		name:  "nil",
		input: nil,
		want:  false,
	}, {
		name: "no owner reference",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		},
		want: false,
	}, {
		name: "wrong owner reference, not controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "another.knative.dev/v1beta3",
					Kind:       "Parent",
					Controller: &boolFalse,
				}},
			},
		},
		want: false,
	}, {
		name: "right owner reference, not controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Controller: &boolFalse,
				}},
			},
		},
		want: false,
	}, {
		name: "wrong owner reference, but controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "another.knative.dev/v1beta3",
					Kind:       "Parent",
					Controller: &boolTrue,
				}},
			},
		},
		want: false,
	}, {
		name: "right owner reference, is controller",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Controller: &boolTrue,
				}},
			},
		},
		want: true,
	}, {
		name: "right owner reference, is controller, wrong version",
		input: &Resource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: schema.GroupVersion{Group: gvk.Group, Version: "other"}.String(),
					Kind:       gvk.Kind,
					Controller: &boolTrue,
				}},
			},
		},
		want: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filter(test.input)
			if test.want != got {
				t.Errorf("Filter() = %v, wanted %v", got, test.want)
			}
		})
	}
}

type NopReconciler struct{}

func (nr *NopReconciler) Reconcile(context.Context, string) error {
	return nil
}

func TestEnqueues(t *testing.T) {
	tests := []struct {
		name      string
		work      func(*Impl)
		wantQueue []types.NamespacedName
	}{{
		name: "do nothing",
		work: func(*Impl) {},
	}, {
		name: "enqueue key",
		work: func(impl *Impl) {
			impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})
		},
		wantQueue: []types.NamespacedName{{Namespace: "foo", Name: "bar"}},
	}, {
		name: "enqueue duplicate key",
		work: func(impl *Impl) {
			impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})
			impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})
		},
		// The queue deduplicates.
		wantQueue: []types.NamespacedName{{Namespace: "foo", Name: "bar"}},
	}, {
		name: "enqueue different keys",
		work: func(impl *Impl) {
			impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})
			impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "baz"})
		},
		wantQueue: []types.NamespacedName{{Namespace: "foo", Name: "bar"}, {Namespace: "foo", Name: "baz"}},
	}, {
		name: "enqueue resource",
		work: func(impl *Impl) {
			impl.Enqueue(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "bar", Name: "foo"}},
	}, {
		name: "enqueue sentinel resource",
		work: func(impl *Impl) {
			e := impl.EnqueueSentinel(types.NamespacedName{Namespace: "foo", Name: "bar"})
			e(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "baz",
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "foo", Name: "bar"}},
	}, {
		name: "enqueue duplicate sentinel resource",
		work: func(impl *Impl) {
			e := impl.EnqueueSentinel(types.NamespacedName{Namespace: "foo", Name: "bar"})
			e(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "baz-1",
				},
			})
			e(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "baz-2",
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "foo", Name: "bar"}},
	}, {
		name: "enqueue bad resource",
		work: func(impl *Impl) {
			impl.Enqueue("baz/blah")
		},
	}, {
		name: "enqueue controller of bad resource",
		work: func(impl *Impl) {
			impl.EnqueueControllerOf("baz/blah")
		},
	}, {
		name: "enqueue controller of resource without owner",
		work: func(impl *Impl) {
			impl.EnqueueControllerOf(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			})
		},
	}, {
		name: "enqueue controller of resource with owner",
		work: func(impl *Impl) {
			impl.EnqueueControllerOf(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: gvk.GroupVersion().String(),
						Kind:       gvk.Kind,
						Name:       "baz",
						Controller: &boolTrue,
					}},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "bar", Name: "baz"}},
	}, {
		name: "enqueue controller of deleted resource with owner",
		work: func(impl *Impl) {
			impl.EnqueueControllerOf(cache.DeletedFinalStateUnknown{
				Key: "foo/bar",
				Obj: &Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: gvk.GroupVersion().String(),
							Kind:       gvk.Kind,
							Name:       "baz",
							Controller: &boolTrue,
						}},
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "bar", Name: "baz"}},
	}, {
		name: "enqueue controller of deleted bad resource",
		work: func(impl *Impl) {
			impl.EnqueueControllerOf(cache.DeletedFinalStateUnknown{
				Key: "foo/bar",
				Obj: "bad-resource",
			})
		},
	}, {
		name: "enqueue label of namespaced resource bad resource",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("test-ns", "test-name")("baz/blah")
		},
	}, {
		name: "enqueue label of namespaced resource without label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("ns-key", "name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"ns-key": "bar",
					},
				},
			})
		},
	}, {
		name: "enqueue label of namespaced resource without namespace label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("ns-key", "name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"name-key": "baz",
					},
				},
			})
		},
	}, {
		name: "enqueue label of namespaced resource with labels",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("ns-key", "name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"ns-key":   "qux",
						"name-key": "baz",
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "qux", Name: "baz"}},
	}, {
		name: "enqueue label of namespaced resource with empty namespace label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("", "name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"name-key": "baz",
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "bar", Name: "baz"}},
	}, {
		name: "enqueue label of deleted namespaced resource with label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("ns-key", "name-key")(cache.DeletedFinalStateUnknown{
				Key: "foo/bar",
				Obj: &Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						Labels: map[string]string{
							"ns-key":   "qux",
							"name-key": "baz",
						},
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "qux", Name: "baz"}},
	}, {
		name: "enqueue label of deleted bad namespaced resource",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfNamespaceScopedResource("ns-key", "name-key")(cache.DeletedFinalStateUnknown{
				Key: "foo/bar",
				Obj: "bad-resource",
			})
		},
	}, {
		name: "enqueue label of cluster scoped resource bad resource",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfClusterScopedResource("name-key")("baz")
		},
	}, {
		name: "enqueue label of cluster scoped resource without label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfClusterScopedResource("name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels:    map[string]string{},
				},
			})
		},
	}, {
		name: "enqueue label of cluster scoped resource with label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfClusterScopedResource("name-key")(&Resource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Labels: map[string]string{
						"name-key": "baz",
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "", Name: "baz"}},
	}, {
		name: "enqueue label of deleted cluster scoped resource with label",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfClusterScopedResource("name-key")(cache.DeletedFinalStateUnknown{
				Key: "foo/bar",
				Obj: &Resource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						Labels: map[string]string{
							"name-key": "baz",
						},
					},
				},
			})
		},
		wantQueue: []types.NamespacedName{{Namespace: "", Name: "baz"}},
	}, {
		name: "enqueue label of deleted bad cluster scoped resource",
		work: func(impl *Impl) {
			impl.EnqueueLabelOfClusterScopedResource("name-key")(cache.DeletedFinalStateUnknown{
				Key: "bar",
				Obj: "bad-resource",
			})
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer ClearAll()
			impl := NewImplWithStats(&NopReconciler{}, TestLogger(t), "Testing", &FakeStatsReporter{})
			test.work(impl)

			// The rate limit on our queue delays when things are added to the queue.
			time.Sleep(50 * time.Millisecond)
			impl.WorkQueue.ShutDown()
			gotQueue := drainWorkQueue(impl.WorkQueue)

			if diff := cmp.Diff(test.wantQueue, gotQueue); diff != "" {
				t.Errorf("unexpected queue (-want +got): %s", diff)
			}
		})
	}
}

func TestEnqeueAfter(t *testing.T) {
	defer ClearAll()
	impl := NewImplWithStats(&NopReconciler{}, TestLogger(t), "Testing", &FakeStatsReporter{})
	impl.EnqueueAfter(&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "for",
			Namespace: "waiting",
		},
	}, 5*time.Second)
	impl.EnqueueAfter(&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waterfall",
			Namespace: "the",
		},
	}, 500*time.Millisecond)
	impl.EnqueueAfter(&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "to",
			Namespace: "fall",
		},
	}, 20*time.Second)
	time.Sleep(10 * time.Millisecond)
	if got, want := impl.WorkQueue.Len(), 0; got != want {
		t.Errorf("|Queue| = %d, want: %d", got, want)
	}
	// Sleep the remaining time.
	time.Sleep(time.Second)
	if got, want := impl.WorkQueue.Len(), 1; got != want {
		t.Errorf("|Queue| = %d, want: %d", got, want)
	}
	impl.WorkQueue.ShutDown()
	if got, want := drainWorkQueue(impl.WorkQueue), []types.NamespacedName{{Namespace: "the", Name: "waterfall"}}; !cmp.Equal(got, want) {
		t.Errorf("Queue = %v, want: %v, diff: %s", got, want, cmp.Diff(got, want))
	}
}

func TestEnqeueKeyAfter(t *testing.T) {
	defer ClearAll()
	impl := NewImplWithStats(&NopReconciler{}, TestLogger(t), "Testing", &FakeStatsReporter{})
	impl.EnqueueKeyAfter(types.NamespacedName{Namespace: "waiting", Name: "for"}, 5*time.Second)
	impl.EnqueueKeyAfter(types.NamespacedName{Namespace: "the", Name: "waterfall"}, 500*time.Millisecond)
	impl.EnqueueKeyAfter(types.NamespacedName{Namespace: "to", Name: "fall"}, 20*time.Second)
	time.Sleep(10 * time.Millisecond)
	if got, want := impl.WorkQueue.Len(), 0; got != want {
		t.Errorf("|Queue| = %d, want: %d", got, want)
	}
	// Sleep the remaining time.
	time.Sleep(time.Second)
	if got, want := impl.WorkQueue.Len(), 1; got != want {
		t.Errorf("|Queue| = %d, want: %d", got, want)
	}
	impl.WorkQueue.ShutDown()
	if got, want := drainWorkQueue(impl.WorkQueue), []types.NamespacedName{{Namespace: "the", Name: "waterfall"}}; !cmp.Equal(got, want) {
		t.Errorf("Queue = %v, want: %v, diff: %s", got, want, cmp.Diff(got, want))
	}
}

type CountingReconciler struct {
	m     sync.Mutex
	Count int
}

func (cr *CountingReconciler) Reconcile(context.Context, string) error {
	cr.m.Lock()
	defer cr.m.Unlock()
	cr.Count++
	return nil
}

func TestStartAndShutdown(t *testing.T) {
	defer ClearAll()
	r := &CountingReconciler{}
	impl := NewImplWithStats(r, TestLogger(t), "Testing", &FakeStatsReporter{})

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		StartAll(stopCh, impl)
	}()

	select {
	case <-time.After(10 * time.Millisecond):
		// We don't expect completion before the stopCh closes.
	case <-doneCh:
		t.Error("StartAll finished early.")
	}
	close(stopCh)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for controller to finish.")
	case <-doneCh:
		// We expect the work to complete.
	}

	if got, want := r.Count, 0; got != want {
		t.Errorf("Count = %v, wanted %v", got, want)
	}
}

func TestStartAndShutdownWithWork(t *testing.T) {
	defer ClearAll()
	r := &CountingReconciler{}
	reporter := &FakeStatsReporter{}
	impl := NewImplWithStats(r, TestLogger(t), "Testing", reporter)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})

	go func() {
		defer close(doneCh)
		StartAll(stopCh, impl)
	}()

	select {
	case <-time.After(10 * time.Millisecond):
		// We don't expect completion before the stopCh closes.
	case <-doneCh:
		t.Error("StartAll finished early.")
	}
	close(stopCh)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for controller to finish.")
	case <-doneCh:
		// We expect the work to complete.
	}

	if got, want := r.Count, 1; got != want {
		t.Errorf("Count = %v, wanted %v", got, want)
	}
	if got, want := impl.WorkQueue.NumRequeues(types.NamespacedName{Namespace: "foo", Name: "bar"}), 0; got != want {
		t.Errorf("Count = %v, wanted %v", got, want)
	}

	checkStats(t, reporter, 1, 0, 1, trueString)
}

type ErrorReconciler struct{}

func (er *ErrorReconciler) Reconcile(context.Context, string) error {
	return errors.New("I always error")
}

func TestStartAndShutdownWithErroringWork(t *testing.T) {
	defer ClearAll()
	r := &ErrorReconciler{}
	reporter := &FakeStatsReporter{}
	impl := NewImplWithStats(r, TestLogger(t), "Testing", reporter)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	impl.EnqueueKey(types.NamespacedName{Namespace: "", Name: "bar"})

	go func() {
		defer close(doneCh)
		// StartAll blocks until all the worker threads finish, which shouldn't
		// be until we close stopCh.
		StartAll(stopCh, impl)
	}()

	select {
	case <-time.After(1 * time.Second):
		// We don't expect completion before the stopCh closes,
		// but the workers should spin on the erroring work.

	case <-doneCh:
		t.Error("StartAll finished early.")
	}

	// By closing the stopCh all the workers should complete and
	// we should close the doneCh.
	close(stopCh)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for controller to finish.")
	case <-doneCh:
		// We expect any outstanding work to complete, for the worker
		// threads to complete and for doneCh to close in a timely manner.
	}

	// Check that the work was requeued in RateLimiter.
	// As NumRequeues can't fully reflect the real state of queue length.
	// Here we need to wait for NumRequeues to be more than 1, to ensure
	// the key get re-queued and reprocessed as expect.
	if got, wantAtLeast := impl.WorkQueue.NumRequeues(types.NamespacedName{Namespace: "", Name: "bar"}), 2; got < wantAtLeast {
		t.Errorf("Requeue count = %v, wanted at least %v", got, wantAtLeast)
	}
}

type PermanentErrorReconciler struct{}

func (er *PermanentErrorReconciler) Reconcile(context.Context, string) error {
	err := errors.New("I always error")
	return NewPermanentError(err)
}

func TestStartAndShutdownWithPermanentErroringWork(t *testing.T) {
	defer ClearAll()
	r := &PermanentErrorReconciler{}
	reporter := &FakeStatsReporter{}
	impl := NewImplWithStats(r, TestLogger(t), "Testing", reporter)

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	impl.EnqueueKey(types.NamespacedName{Namespace: "foo", Name: "bar"})

	go func() {
		defer close(doneCh)
		StartAll(stopCh, impl)
	}()

	select {
	case <-time.After(20 * time.Millisecond):
		// We don't expect completion before the stopCh closes.
	case <-doneCh:
		t.Error("StartAll finished early.")
	}
	close(stopCh)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for controller to finish.")
	case <-doneCh:
		// We expect the work to complete.
	}

	// Check that the work was not requeued in RateLimiter.
	if got, want := impl.WorkQueue.NumRequeues(types.NamespacedName{Namespace: "foo", Name: "bar"}), 0; got != want {
		t.Errorf("Requeue count = %v, wanted %v", got, want)
	}

	checkStats(t, reporter, 1, 0, 1, falseString)
}

func drainWorkQueue(wq workqueue.RateLimitingInterface) (hasQueue []types.NamespacedName) {
	for {
		key, shutdown := wq.Get()
		if key == nil && shutdown {
			break
		}
		hasQueue = append(hasQueue, key.(types.NamespacedName))
	}
	return
}

type dummyInformer struct {
	cache.SharedInformer
}

type dummyStore struct {
	cache.Store
}

func (*dummyInformer) GetStore() cache.Store {
	return &dummyStore{}
}

var dummyKeys = []string{"foo/bar", "bar/foo", "fizz/buzz"}
var dummyObjs = []interface{}{
	&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar",
			Namespace: "foo",
		},
	},
	&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
	},
	&Resource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buzz",
			Namespace: "fizz",
		},
	},
}

func (*dummyStore) ListKeys() []string {
	return dummyKeys
}

func (*dummyStore) List() []interface{} {
	return dummyObjs
}

func TestImplGlobalResync(t *testing.T) {
	defer ClearAll()
	r := &CountingReconciler{}
	impl := NewImplWithStats(r, TestLogger(t), "Testing", &FakeStatsReporter{})

	stopCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		StartAll(stopCh, impl)
	}()

	impl.GlobalResync(&dummyInformer{})

	// The global resync delays enqueuing things by a second with a jitter that
	// goes up to len(dummyObjs) times a second.
	select {
	case <-time.After((1 + 3) * time.Second):
		// We don't expect completion before the stopCh closes.
	case <-doneCh:
		t.Error("StartAll finished early.")
	}
	close(stopCh)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for controller to finish.")
	case <-doneCh:
		// We expect the work to complete.
	}

	if want, got := 3, r.Count; want != got {
		t.Errorf("GlobalResync: want = %v, got = %v", want, got)
	}
}

func checkStats(t *testing.T, r *FakeStatsReporter, reportCount, lastQueueDepth, reconcileCount int, lastReconcileSuccess string) {
	qd := r.GetQueueDepths()
	if got, want := len(qd), reportCount; got != want {
		t.Errorf("Queue depth reports = %v, wanted %v", got, want)
	}
	if got, want := qd[len(qd)-1], int64(lastQueueDepth); got != want {
		t.Errorf("Queue depth report = %v, wanted %v", got, want)
	}
	rd := r.GetReconcileData()
	if got, want := len(rd), reconcileCount; got != want {
		t.Errorf("Reconcile reports = %v, wanted %v", got, want)
	}
	if got, want := rd[len(rd)-1].Success, lastReconcileSuccess; got != want {
		t.Errorf("Reconcile success = %v, wanted %v", got, want)
	}
}

type fixedInformer struct {
	m    sync.Mutex
	sunk bool
	done bool
}

var _ Informer = (*fixedInformer)(nil)

func (fi *fixedInformer) Run(stopCh <-chan struct{}) {
	<-stopCh

	fi.m.Lock()
	defer fi.m.Unlock()
	fi.done = true
}

func (fi *fixedInformer) HasSynced() bool {
	fi.m.Lock()
	defer fi.m.Unlock()
	return fi.sunk
}

func (fi *fixedInformer) ToggleSynced(b bool) {
	fi.m.Lock()
	defer fi.m.Unlock()
	fi.sunk = b
}

func (fi *fixedInformer) Done() bool {
	fi.m.Lock()
	defer fi.m.Unlock()
	return fi.done
}

func TestStartInformersSuccess(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: true}

	stopCh := make(chan struct{})
	defer close(stopCh)
	go func() {
		errCh <- StartInformers(stopCh, fi)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for informers to sync.")
	}
}

func TestStartInformersEventualSuccess(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: false}

	stopCh := make(chan struct{})
	defer close(stopCh)
	go func() {
		errCh <- StartInformers(stopCh, fi)
	}()

	select {
	case err := <-errCh:
		t.Errorf("Unexpected send on errCh: %v", err)
	case <-time.After(1 * time.Second):
		// Wait a brief period to ensure nothing is sent.
	}

	// Let the Sync complete.
	fi.ToggleSynced(true)

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for informers to sync.")
	}
}

func TestStartInformersFailure(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: false}

	stopCh := make(chan struct{})
	go func() {
		errCh <- StartInformers(stopCh, fi)
	}()

	select {
	case err := <-errCh:
		t.Errorf("Unexpected send on errCh: %v", err)
	case <-time.After(1 * time.Second):
		// Wait a brief period to ensure nothing is sent.
	}

	// Now close the stopCh and we should see an error sent.
	close(stopCh)

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("Unexpected success syncing informers after stopCh closed.")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timed out waiting for informers to sync.")
	}
}

func TestRunInformersSuccess(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: true}

	stopCh := make(chan struct{})
	go func() {
		_, err := RunInformers(stopCh, fi)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for informers to sync.")
	}

	close(stopCh)
}

func TestRunInformersEventualSuccess(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: false}

	stopCh := make(chan struct{})
	go func() {
		_, err := RunInformers(stopCh, fi)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		t.Fatalf("Unexpected send on errCh: %v", err)
	case <-time.After(1 * time.Second):
		// Wait a brief period to ensure nothing is sent.
	}

	// Let the Sync complete.
	fi.ToggleSynced(true)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for informers to sync.")
	}

	close(stopCh)
}

func TestRunInformersFailure(t *testing.T) {
	errCh := make(chan error)
	defer close(errCh)

	fi := &fixedInformer{sunk: false}

	stopCh := make(chan struct{})
	go func() {
		_, err := RunInformers(stopCh, fi)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		t.Errorf("Unexpected send on errCh: %v", err)
	case <-time.After(1 * time.Second):
		// Wait a brief period to ensure nothing is sent.
	}

	// Now close the stopCh and we should see an error sent.
	close(stopCh)

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("Unexpected success syncing informers after stopCh closed.")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for informers to sync.")
	}
}

func TestRunInformersFinished(t *testing.T) {
	fi := &fixedInformer{sunk: true}
	defer func() {
		if !fi.Done() {
			t.Fatalf("Test didn't wait for informers to finish")
		}
	}()

	ctx, cancel := context.WithCancel(TestContextWithLogger(t))

	waitInformers, err := RunInformers(ctx.Done(), fi)
	if err != nil {
		t.Fatalf("Failed to start informers: %v", err)
	}

	cancel()

	ch := make(chan struct{})
	go func() {
		waitInformers()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for informers to finish.")
	}
}

func TestGetResyncPeriod(t *testing.T) {
	ctx := context.Background()

	if got := GetResyncPeriod(ctx); got != DefaultResyncPeriod {
		t.Errorf("GetResyncPeriod() = %v, wanted %v", got, nil)
	}

	bob := 30 * time.Second
	ctx = WithResyncPeriod(ctx, bob)

	if want, got := bob, GetResyncPeriod(ctx); got != want {
		t.Errorf("GetResyncPeriod() = %v, wanted %v", got, want)
	}

	tribob := 90 * time.Second
	if want, got := tribob, GetTrackerLease(ctx); got != want {
		t.Errorf("GetTrackerLease() = %v, wanted %v", got, want)
	}
}

func TestGetEventRecorder(t *testing.T) {
	ctx := context.Background()

	if got := GetEventRecorder(ctx); got != nil {
		t.Errorf("GetEventRecorder() = %v, wanted nil", got)
	}

	ctx = WithEventRecorder(ctx, record.NewFakeRecorder(1000))

	if got := GetEventRecorder(ctx); got == nil {
		t.Error("GetEventRecorder() = nil, wanted non-nil")
	}
}
