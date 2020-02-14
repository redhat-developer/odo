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

package psbinding

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	. "knative.dev/pkg/testing/duck"
)

func TestExact(t *testing.T) {
	em := make(exactMatcher, 1)

	want := &TestBindable{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "blah",
			Name:      "asdf",
		},
		Spec: TestBindableSpec{
			Foo: "bar",
		},
	}

	gvk := want.GetGroupVersionKind()
	key := exactKey{
		Group:     gvk.Group,
		Kind:      gvk.Kind,
		Namespace: want.GetNamespace(),
		Name:      want.GetName(),
	}

	// Before we Add something, we shouldn't be able to Get anything.
	if got, ok := em.Get(key); ok {
		t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
	}

	// Now Add it.
	em.Add(key, want)

	// After we Add something, we should be able to Get it.
	if got, ok := em.Get(key); !ok {
		t.Errorf("Get(%+v) = %v, %v; wanted b, true", key, want, ok)
	} else if !cmp.Equal(got, want) {
		t.Errorf("Get (-want, +got): %s", cmp.Diff(want, got))
	}

	otherKey := exactKey{
		Group:     "apps",
		Kind:      "Deployment",
		Namespace: "foo",
		Name:      "bar",
	}

	// After we Add something, we still shouldn't return things for other keys.
	if got, ok := em.Get(otherKey); ok {
		t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
	}
}

func TestInexact(t *testing.T) {
	want := &TestBindable{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "blah",
			Name:      "asdf",
			Labels: map[string]string{
				"foo": "bar",
				"baz": "blah",
			},
		},
		Spec: TestBindableSpec{
			Foo: "bar",
		},
	}
	ls := labels.Set(want.Labels)

	gvk := want.GetGroupVersionKind()
	key := inexactKey{
		Group:     gvk.Group,
		Kind:      gvk.Kind,
		Namespace: want.GetNamespace(),
	}

	t.Run("empty matcher doesn't match", func(t *testing.T) {
		im := make(inexactMatcher, 1)
		if got, ok := im.Get(key, ls); ok {
			t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
		}
	})

	t.Run("matcher with exact labels matches", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Use exactly the labels from the resource.
		selector := ls.AsSelector()

		im.Add(key, selector, want)

		// With an appropriate selector, we match and get the binding.
		if got, ok := im.Get(key, ls); !ok {
			t.Errorf("Get(%+v) = %v, %v; wanted b, true", key, want, ok)
		} else if !cmp.Equal(got, want) {
			t.Errorf("Get (-want, +got): %s", cmp.Diff(want, got))
		}
	})

	t.Run("matcher for everything matches", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Match everything.
		selector := labels.Everything()

		im.Add(key, selector, want)

		// With an appropriate selector, we match and get the binding.
		if got, ok := im.Get(key, ls); !ok {
			t.Errorf("Get(%+v) = %v, %v; wanted b, true", key, want, ok)
		} else if !cmp.Equal(got, want) {
			t.Errorf("Get (-want, +got): %s", cmp.Diff(want, got))
		}
	})

	t.Run("matcher for nothing does not match", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Match nothing.
		selector := labels.Nothing()

		im.Add(key, selector, want)

		if got, ok := im.Get(key, ls); ok {
			t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
		}
	})

	t.Run("matcher with a subset of labels matches", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Use a subset of the resources labels.
		selector := labels.Set(map[string]string{
			"foo": "bar",
		}).AsSelector()

		im.Add(key, selector, want)

		// With an appropriate selector, we match and get the binding.
		if got, ok := im.Get(key, ls); !ok {
			t.Errorf("Get(%+v) = %v, %v; wanted b, true", key, want, ok)
		} else if !cmp.Equal(got, want) {
			t.Errorf("Get (-want, +got): %s", cmp.Diff(want, got))
		}
	})

	t.Run("matcher with overlapping labels does not match", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Use a subset of the resources labels.
		selector := labels.Set(map[string]string{
			"foo": "bar",
			"not": "found",
		}).AsSelector()

		im.Add(key, selector, want)

		// We shouldn't match because the second labels shouldn't match.
		if got, ok := im.Get(key, ls); ok {
			t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
		}
	})

	t.Run("matcher with exact labels doesn't match a different namespace", func(t *testing.T) {
		im := make(inexactMatcher, 1)

		// Use exactly the labels from the resource.
		selector := ls.AsSelector()

		im.Add(key, selector, want)

		otherKey := key
		otherKey.Namespace = "another"

		// We shouldn't match because the second labels shouldn't match.
		if got, ok := im.Get(otherKey, ls); ok {
			t.Errorf("Get(%+v) = %v, %v; wanted nil, false", key, got, ok)
		}
	})
}
