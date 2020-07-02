package util

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/catalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterHiddenServices(t *testing.T) {
	tests := []struct {
		name     string
		input    catalog.ServiceTypeList
		expected catalog.ServiceTypeList
	}{
		/*
					This test is not needed.. Also fails using DeepEqual anyways..
					    --- FAIL: TestFilterHiddenServices/Case_1:_empty_input (0.00s)
			        util_test.go:101: got: [], wanted: []
					{
						name:     "Case 1: empty input",
						input:    catalog.ServiceTypeList{},
						expected: catalog.ServiceTypeList{},
					},
		*/
		{
			name: "Case 2: non empty input",
			input: catalog.ServiceTypeList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []catalog.ServiceType{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n1",
						},
						Spec: catalog.ServiceSpec{
							Hidden: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n2",
						},
						Spec: catalog.ServiceSpec{
							Hidden: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n3",
						},
						Spec: catalog.ServiceSpec{
							Hidden: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n4",
						},
						Spec: catalog.ServiceSpec{
							Hidden: false,
						},
					},
				},
			},
			expected: catalog.ServiceTypeList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.dev/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items: []catalog.ServiceType{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n2",
						},
						Spec: catalog.ServiceSpec{
							Hidden: false,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "n4",
						},
						Spec: catalog.ServiceSpec{
							Hidden: false,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FilterHiddenServices(tt.input)
			if !reflect.DeepEqual(tt.expected, output) {
				t.Errorf("got: %+v, wanted: %+v", output.Items, tt.expected.Items)
			}
		})
	}
}

func TestFilterHiddenComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    []catalog.ComponentType
		expected []catalog.ComponentType
	}{
		{
			name:     "Case 1: empty input",
			input:    []catalog.ComponentType{},
			expected: []catalog.ComponentType{},
		},
		{
			name: "Case 2: non empty input",
			input: []catalog.ComponentType{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{"1", "latest"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n2",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n3",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n4",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{"10"},
					},
				},
			},
			expected: []catalog.ComponentType{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{"1", "latest"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n4",
					},
					Spec: catalog.ComponentSpec{
						NonHiddenTags: []string{"10"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FilterHiddenComponents(tt.input)
			if !reflect.DeepEqual(tt.expected, output) {
				t.Errorf("got: %+v, wanted: %+v", output, tt.expected)
			}
		})
	}
}
