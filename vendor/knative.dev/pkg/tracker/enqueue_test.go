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

package tracker

import (
	"regexp"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/kmeta"
	. "knative.dev/pkg/testing"
)

func TestHappyPathsExact(t *testing.T) {
	calls := 0
	f := func(key types.NamespacedName) {
		calls++
	}

	trk := New(f, 100*time.Millisecond)

	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
		},
	}
	or := kmeta.ObjectReference(thing1)
	ref := Reference{
		APIVersion: or.APIVersion,
		Kind:       or.Kind,
		Namespace:  or.Namespace,
		Name:       or.Name,
	}

	thing2 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "reffer.knative.dev/v1alpha1",
			Kind:       "Thing2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "bar.baz.this-is-fine",
		},
	}

	// Not tracked yet
	{
		trk.OnChanged(thing1)
		if got, want := calls, 0; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Tracked gets called
	{
		if err := trk.Track(ref.ObjectReference(), thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 1; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		trk.OnChanged(thing1)
		if got, want := calls, 2; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Still gets called
	{
		trk.OnChanged(thing1)
		if got, want := calls, 3; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Check that after the sleep duration, we stop getting called.
	time.Sleep(101 * time.Millisecond)
	{
		trk.OnChanged(thing1)
		if got, want := calls, 3; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
		if _, stillThere := trk.(*impl).exact[ref]; stillThere {
			t.Fatal("Timeout passed, but exact for objectReference is still there")
		}
	}

	// Starts getting called again
	{
		if err := trk.Track(ref.ObjectReference(), thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 4; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		trk.OnChanged(thing1)
		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// OnChanged non-accessor
	{
		// Check that passing in a resource that doesn't implement
		// accessor won't panic.
		trk.OnChanged("not an accessor")

		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// OnChanged non-accessor in DeletedFinalStateUnknown
	{
		// Check that passing in a DeletedFinalStateUnknown instance
		// with a resource that doesn't implement accessor won't get
		// Tracked called, and won't panic.
		trk.OnChanged(cache.DeletedFinalStateUnknown{
			Key: "ns/foo",
			Obj: "not an accessor",
		})

		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Tracked gets called by DeletedFinalStateUnknown
	{
		trk.OnChanged(cache.DeletedFinalStateUnknown{
			Key: "ns/foo",
			Obj: thing1,
		})
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Stops tracking explicitly
	{
		trk.OnDeletedObserver(thing2)
		trk.OnChanged(thing1)
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Track bad object
	{
		if err := trk.Track(ref.ObjectReference(), struct{}{}); err == nil {
			t.Fatal("Track() = nil, wanted error")
		}
	}
}

func TestAllowedObjectReferences(t *testing.T) {
	trk := New(func(key types.NamespacedName) {}, 10*time.Millisecond)
	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
		},
	}
	tests := []struct {
		name   string
		objRef corev1.ObjectReference
	}{{
		name: "Pod",
		objRef: corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace:  "default",
			Name:       "test",
		},
	}, {
		name: "Non-core resource",
		objRef: corev1.ObjectReference{
			APIVersion: "custom.example.com/v1alpha17",
			Kind:       "Widget",
			Namespace:  "default",
			Name:       "test",
		},
	}, {
		name: "Complex Kind",
		objRef: corev1.ObjectReference{
			APIVersion: "custom.example.com/v1alpha17",
			Kind:       "Widget_v3",
			Namespace:  "default",
			Name:       "test",
		},
	}, {
		name: "Dashed Namespace",
		objRef: corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Namespace:  "not-default",
			Name:       "test",
		},
	}, {
		name: "Complex Name",
		objRef: corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Namespace:  "default",
			Name:       "test.example.cluster.local",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := trk.Track(test.objRef, thing1); err != nil {
				t.Fatalf("Track() on %v returned error: %v", test.objRef, err)
			}
		})
	}
}

func TestBadObjectReferences(t *testing.T) {
	trk := New(func(key types.NamespacedName) {}, 10*time.Millisecond)
	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
		},
	}

	tests := []struct {
		name   string
		objRef corev1.ObjectReference
		match  string
	}{{
		name: "Missing APIVersion",
		objRef: corev1.ObjectReference{
			// APIVersion: "build.knative.dev/v1alpha1",
			Kind:      "Build",
			Namespace: "default",
			Name:      "kaniko",
		},
		match: "APIVersion",
	}, {
		name: "Bad char in APIVersion",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev%v1alpha1",
			Kind:       "Build",
			Namespace:  "default",
			Name:       "kaniko",
		},
		match: "APIVersion",
	}, {
		name: "Extra slashes in APIVersion",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1/alpha1",
			Kind:       "Build",
			Namespace:  "default",
			Name:       "kaniko",
		},
		match: "APIVersion",
	}, {
		name: "Missing Kind",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			// Kind:      "Build",
			Namespace: "default",
			Name:      "kaniko",
		},
		match: "Kind",
	}, {
		name: "Invalid Kind",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build.1",
			Namespace:  "default",
			Name:       "kaniko",
		},
		match: "Kind",
	}, {
		name: "Missing Namespace",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			// Namespace: "default",
			Name: "kaniko",
		},
		match: "Namespace",
	}, {
		name: "Capital in Namespace",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			Namespace:  "Default",
			Name:       "kaniko",
		},
		match: "Namespace",
	}, {
		name: "Domain-separated Namespace",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			Namespace:  "not.default",
			Name:       "kaniko",
		},
		match: "Namespace",
	}, {
		name: "Missing Name",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			Namespace:  "default",
			// Name:      "kaniko",
		},
		match: "Name",
	}, {
		name: "Capital in Name",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			Namespace:  "default",
			Name:       "Kaniko",
		},
		match: "Name",
	}, {
		name: "Bad char in Name",
		objRef: corev1.ObjectReference{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
			Namespace:  "default",
			Name:       "kaniko_small",
		},
		match: "Name",
	}, {
		name:   "Missing All",
		objRef: corev1.ObjectReference{
			// APIVersion: "build.knative.dev/v1alpha1",
			// Kind:       "Build",
			// Namespace:  "default",
			// Name:      "kaniko",
		},
		match: "\nAPIVersion:.*\nKind:.*\nNamespace:",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := trk.Track(test.objRef, thing1); err == nil {
				t.Fatal("Track() = nil, wanted error")
			} else {
				match, e2 := regexp.Match(test.match, []byte(err.Error()))
				if e2 != nil {
					t.Fatalf("Failed to compile %q: %v", e2, test.match)
				} else if !match {
					t.Fatalf("Track() = %v, wanted match: %s", err, test.match)
				}
			}
		})
	}
}

func TestBadReferences(t *testing.T) {
	trk := New(func(key types.NamespacedName) {}, 10*time.Millisecond)
	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
		},
	}

	tests := []struct {
		name   string
		objRef Reference
		match  string
	}{{
		name:   "Missing All",
		objRef: Reference{},
		match:  "\nAPIVersion:.*\nKind:.*\nNamespace:",
	}, {
		name: "Name and Selector",
		objRef: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Name:       "foo",
			Selector:   &metav1.LabelSelector{},
		},
		match: "both Name and Selector",
	}, {
		name: "bad label key",
		objRef: Reference{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Namespace:  "default",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"a bad key": "bar",
				},
			},
		},
		match: "a bad key",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := trk.TrackReference(test.objRef, thing1); err == nil {
				t.Fatal("Track() = nil, wanted error")
			} else {
				match, e2 := regexp.Match(test.match, []byte(err.Error()))
				if e2 != nil {
					t.Fatalf("Failed to compile %q: %v", e2, test.match)
				} else if !match {
					t.Fatalf("Track() = %v, wanted match: %s", err, test.match)
				}
			}
		})
	}
}

func TestHappyPathsInexact(t *testing.T) {
	calls := 0
	f := func(key types.NamespacedName) {
		calls++
	}

	trk := New(f, 100*time.Millisecond)

	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
			Labels: map[string]string{
				"foo": "bar",
				// An extra label.
				"baz": "blah",
			},
		},
	}
	or := kmeta.ObjectReference(thing1)
	ref := Reference{
		APIVersion: or.APIVersion,
		Kind:       or.Kind,
		Namespace:  or.Namespace,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"foo": "bar",
			},
		},
	}

	thing2 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "reffer.knative.dev/v1alpha1",
			Kind:       "Thing2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "bar.baz.this-is-fine",
		},
	}

	// Not tracked yet
	{
		trk.OnChanged(thing1)
		if got, want := calls, 0; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Tracked gets called
	{
		if err := trk.TrackReference(ref, thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 1; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		trk.OnChanged(thing1)
		if got, want := calls, 2; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Still gets called
	{
		trk.OnChanged(thing1)
		if got, want := calls, 3; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Check that after the sleep duration, we stop getting called.
	time.Sleep(101 * time.Millisecond)
	{
		trk.OnChanged(thing1)
		if got, want := calls, 3; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
		if _, stillThere := trk.(*impl).exact[ref]; stillThere {
			t.Fatal("Timeout passed, but exact for objectReference is still there")
		}
	}

	// Starts getting called again
	{
		if err := trk.TrackReference(ref, thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 4; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		trk.OnChanged(thing1)
		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// OnChanged non-accessor
	{
		// Check that passing in a resource that doesn't implement
		// accessor won't panic.
		trk.OnChanged("not an accessor")

		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// OnChanged non-accessor in DeletedFinalStateUnknown
	{
		// Check that passing in a DeletedFinalStateUnknown instance
		// with a resource that doesn't implement accessor won't get
		// Tracked called, and won't panic.
		trk.OnChanged(cache.DeletedFinalStateUnknown{
			Key: "ns/foo",
			Obj: "not an accessor",
		})

		if got, want := calls, 5; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Tracked gets called by DeletedFinalStateUnknown
	{
		trk.OnChanged(cache.DeletedFinalStateUnknown{
			Key: "ns/foo",
			Obj: thing1,
		})
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Not called when something about the reference matching changes.
	{
		if err := trk.TrackReference(ref, thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 6; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		thing1unlabeled := thing1.DeepCopy()
		thing1unlabeled.Labels = nil
		trk.OnChanged(thing1unlabeled)
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}

		thing1othernamespace := thing1.DeepCopy()
		thing1othernamespace.Namespace = "another"
		trk.OnChanged(thing1othernamespace)
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}

		thing1othergroup := thing1.DeepCopy()
		thing1othergroup.APIVersion = "apps/v1"
		trk.OnChanged(thing1othergroup)
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}

		thing1otherkind := thing1.DeepCopy()
		thing1otherkind.Kind = "deployment"
		trk.OnChanged(thing1otherkind)
		if got, want := calls, 6; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}

		// But with labels is still called
		trk.OnChanged(thing1)
		if got, want := calls, 7; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Stops tracking explicitly
	{
		trk.OnDeletedObserver(thing2)
		trk.OnChanged(thing1)
		if got, want := calls, 7; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Track bad object
	{
		if err := trk.TrackReference(ref, struct{}{}); err == nil {
			t.Fatal("Track() = nil, wanted error")
		}
	}
}

func TestHappyPathsByBoth(t *testing.T) {
	calls := 0
	f := func(key types.NamespacedName) {
		calls++
	}

	trk := New(f, 100*time.Millisecond)

	thing1 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ref.knative.dev/v1alpha1",
			Kind:       "Thing1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
			Labels: map[string]string{
				"foo": "bar",
				// An extra label.
				"baz": "blah",
			},
		},
	}
	or := kmeta.ObjectReference(thing1)
	ref1 := Reference{
		APIVersion: or.APIVersion,
		Kind:       or.Kind,
		Namespace:  or.Namespace,
		Name:       or.Name,
	}
	ref2 := Reference{
		APIVersion: or.APIVersion,
		Kind:       or.Kind,
		Namespace:  or.Namespace,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"foo": "bar",
			},
		},
	}

	thing2 := &Resource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "reffer.knative.dev/v1alpha1",
			Kind:       "Thing2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "bar.baz.this-is-fine",
		},
	}

	// Not tracked yet
	{
		trk.OnChanged(thing1)
		if got, want := calls, 0; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Tracked gets called
	{
		if err := trk.TrackReference(ref1, thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 1; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		if err := trk.TrackReference(ref2, thing2); err != nil {
			t.Fatalf("Track() = %v", err)
		}
		// New registrations should result in an immediate callback.
		if got, want := calls, 2; got != want {
			t.Fatalf("Track() = %v, wanted %v", got, want)
		}

		// The callback should be called for each of the tracks (exact and inexact)
		trk.OnChanged(thing1)
		if got, want := calls, 4; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}

	// Stops tracking explicitly
	{
		trk.OnDeletedObserver(thing2)
		trk.OnChanged(thing1)
		if got, want := calls, 4; got != want {
			t.Fatalf("OnChanged() = %v, wanted %v", got, want)
		}
	}
}
