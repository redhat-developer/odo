package v1

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"

	corev1 "k8s.io/api/core/v1"
)

func TestSetObjectReference(t *testing.T) {
	testCases := []struct {
		name         string
		testRef      corev1.ObjectReference
		startRefs    *[]corev1.ObjectReference
		expectedRefs *[]corev1.ObjectReference
		shouldError  bool
	}{
		{
			name: "add when empty",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{},
			expectedRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
			},
			shouldError: false,
		},
		{
			name: "simple add",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
			},
			shouldError: false,
		},
		{
			name: "replace reference",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
				UID:        "fooid",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
					UID:        "fooid",
				},
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			shouldError: false,
		},
		{
			name: "error on newObject not minObjectReference",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				APIVersion: "test.example.io",
			},
			startRefs:    &[]corev1.ObjectReference{},
			expectedRefs: &[]corev1.ObjectReference{},
			shouldError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetObjectReference(tc.startRefs, tc.testRef)
			if err != nil && !tc.shouldError {
				t.Fatalf("Error occurred unexpectedly: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			if !equality.Semantic.DeepEqual(*tc.startRefs, *tc.expectedRefs) {
				t.Errorf("Unexpected object refs '%v', expected '%v'", tc.startRefs, tc.expectedRefs)
			}
		})
	}
	return
}

func TestRemoveObjectReference(t *testing.T) {
	testCases := []struct {
		name         string
		testRef      corev1.ObjectReference
		startRefs    *[]corev1.ObjectReference
		expectedRefs *[]corev1.ObjectReference
		shouldError  bool
	}{
		{
			name: "remove when empty",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs:    &[]corev1.ObjectReference{},
			expectedRefs: &[]corev1.ObjectReference{},
			shouldError:  false,
		},
		{
			name: "simple remove",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			shouldError: false,
		},
		{
			name: "remove last",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{},
			shouldError:  false,
		},
		{
			// Not sure if this is possible by using SetObjectReference
			// but testing this anyway
			name: "remove matching",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
					UID:        "myuid",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			shouldError: false,
		},
		{
			name: "error on rmObject not minObjectReference",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
				{
					Kind:       "BarKind",
					Namespace:  "test-namespace",
					Name:       "bar",
					APIVersion: "test.example.io",
				},
			},
			expectedRefs: &[]corev1.ObjectReference{},
			shouldError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RemoveObjectReference(tc.startRefs, tc.testRef)
			if err != nil && !tc.shouldError {
				t.Fatalf("Error occurred unexpectedly: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			if !equality.Semantic.DeepEqual(*tc.startRefs, *tc.expectedRefs) {
				t.Errorf("Unexpected object refs '%v', expected '%v'", tc.startRefs, tc.expectedRefs)
			}
		})
	}
	return
}

func TestFindObjectReference(t *testing.T) {
	testCases := []struct {
		name        string
		testRef     corev1.ObjectReference
		startRefs   *[]corev1.ObjectReference
		expectedRef *corev1.ObjectReference
		shouldError bool
	}{
		{
			name: "simple find",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs: &[]corev1.ObjectReference{
				{
					Kind:       "FooKind",
					Namespace:  "test-namespace",
					Name:       "foo",
					APIVersion: "test.example.io",
				},
			},
			expectedRef: &corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			shouldError: false,
		},
		{
			name: "find when empty",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				Namespace:  "test-namespace",
				Name:       "foo",
				APIVersion: "test.example.io",
			},
			startRefs:   &[]corev1.ObjectReference{},
			expectedRef: nil,
			shouldError: false,
		},
		{
			name: "err when not minimal object reference",
			testRef: corev1.ObjectReference{
				Kind:       "FooKind",
				APIVersion: "test.example.io",
			},
			startRefs:   &[]corev1.ObjectReference{},
			expectedRef: nil,
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			foundRef, err := FindObjectReference(*tc.startRefs, tc.testRef)
			if err != nil && !tc.shouldError {
				t.Fatalf("Error occurred unexpectedly: %v", err)
			}
			if err != nil && tc.shouldError {
				return
			}
			if !equality.Semantic.DeepEqual(foundRef, tc.expectedRef) {
				t.Errorf("Unexpected object ref '%v', expected '%v'", foundRef, tc.expectedRef)
			}
		})
	}
	return
}
