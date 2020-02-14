/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Veroute.on 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconciler

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
)

const (
	exampleStatusFailed = "ExampleStatusFailed"
)

func TestNil_Is(t *testing.T) {
	var err error
	if EventIs(err, NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "")) {
		t.Error("Did not expect error to be a ReconcilerEvent")
	}
}

func TestError_Is(t *testing.T) {
	err := errors.New("some other error")
	if EventIs(err, NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "")) {
		t.Error("Did not expect error to be a ReconcilerEvent")
	}
}

func TestNew_Is(t *testing.T) {
	err := NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "this is an example error, %s", "yep")
	if !EventIs(err, NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "")) {
		t.Error("Expected error to be a [Warn, ExampleStatusFailed]")
	}
}

func TestNewOtherType_Is(t *testing.T) {
	err := NewEvent(corev1.EventTypeNormal, exampleStatusFailed, "this is an example error, %s", "yep")
	if EventIs(err, NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "")) {
		t.Error("Expected error to be a [Normal, ExampleStatusFailed], filtered by eventtype failed")
	}
}

func TestNewWrappedErrors_Is(t *testing.T) {
	err := NewEvent(corev1.EventTypeNormal, exampleStatusFailed, "this is a wrapped error, %w", io.ErrUnexpectedEOF)
	if !EventIs(err, io.ErrUnexpectedEOF) {
		t.Error("Event expected to be a wrapped ErrUnexpectedEOF but was not")
	}
}

func TestNewOtherReason_Is(t *testing.T) {
	err := NewEvent(corev1.EventTypeWarning, "otherReason", "this is an example error, %s", "yep")
	if EventIs(err, NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "")) {
		t.Error("Did not expect event to be [Warn, ExampleStatusFailed]")
	}
}

func TestNew_As(t *testing.T) {
	err := NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "this is an example error, %s", "yep")

	var event *ReconcilerEvent
	if !EventAs(err, &event) {
		t.Errorf("Expected error to be a ReconcilerEvent, is not")
	}

	if event.EventType != "Warning" {
		t.Errorf("Mismatched EventType, expected Warning, got %s", event.EventType)
	}
	if event.Reason != exampleStatusFailed {
		t.Errorf("Mismatched Reason, expected ExampleStatusFailed, got %s", event.Reason)
	}
}

func TestNil_As(t *testing.T) {
	var err error

	var event *ReconcilerEvent
	if EventAs(err, &event) {
		t.Error("Did not expect error to be a ReconcilerEvent")
	}
}

func TestNew_Error(t *testing.T) {
	err := NewEvent(corev1.EventTypeWarning, exampleStatusFailed, "this is an example error, %s", "yep")

	const want = "this is an example error, yep"
	got := err.Error()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected diff (-want, +got) = %v", diff)
	}
}
