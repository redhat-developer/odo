/*
Copyright 2020 The Knative Authors.

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
package testing

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "knative.dev/pkg/testing"
	"knative.dev/pkg/tracker"
)

func TestFakeTracker(t *testing.T) {
	t1 := &Resource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "foo",
		},
	}
	t2 := &Resource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "bar.baz.this-is-fine",
		},
	}

	ref1 := tracker.Reference{
		APIVersion: "fakeapi",
		Kind:       "Fake",
		Namespace:  "foo",
		Name:       "bar",
	}
	ref2 := tracker.Reference{
		APIVersion: "fakeapi2",
		Kind:       "Fake2",
		Namespace:  "foo2",
		Name:       "bar2",
	}

	trk := &FakeTracker{}

	// Adding t1 to ref1 and then removing it results in ref1 stopping tracking.
	trk.TrackReference(ref1, t1)
	if !isTracking(trk, ref1) {
		t.Fatalf("Tracker is not tracking %v", ref1)
	}
	trk.OnDeletedObserver(t1)
	if isTracking(trk, ref1) {
		t.Fatalf("Tracker is still tracking %v", ref1)
	}

	// Adding t1, t2 to ref1 and t2 to ref2, then removing t2 results in ref2 stopping
	// tracking.
	trk.TrackReference(ref1, t1)
	trk.TrackReference(ref1, t2)
	trk.TrackReference(ref2, t2)
	if !isTracking(trk, ref1) {
		t.Fatalf("Tracker is not tracking %v", ref1)
	}
	if !isTracking(trk, ref2) {
		t.Fatalf("Tracker is not tracking %v", ref2)
	}
	trk.OnDeletedObserver(t2)
	if !isTracking(trk, ref1) {
		t.Fatalf("Tracker is not tracking %v", ref1)
	}
	if isTracking(trk, ref2) {
		t.Fatalf("Tracker is still tracking %v", ref2)
	}
	trk.OnDeletedObserver(t1)
	if isTracking(trk, ref1) {
		t.Fatalf("Tracker is still tracking %v", ref1)
	}
}

func isTracking(tracker *FakeTracker, ref1 tracker.Reference) bool {
	for _, tracking := range tracker.References() {
		if tracking == ref1 {
			return true
		}
	}
	return false
}
