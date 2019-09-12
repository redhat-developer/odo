package util

import (
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/occlient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestFilterHiddenServices(t *testing.T) {
	tests := []struct {
		name     string
		input    []occlient.Service
		expected []occlient.Service
	}{
		{
			name:     "Case 1: empty input",
			input:    []occlient.Service{},
			expected: []occlient.Service{},
		},
		{
			name: "Case 2: non empty input",
			input: []occlient.Service{
				{
					Name:   "n1",
					Hidden: true,
				},
				{
					Name:   "n2",
					Hidden: false,
				},
				{
					Name:   "n3",
					Hidden: true,
				},
				{
					Name:   "n4",
					Hidden: false,
				},
			},
			expected: []occlient.Service{
				{
					Name:   "n2",
					Hidden: false,
				},
				{
					Name:   "n4",
					Hidden: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FilterHiddenServices(tt.input)
			if !reflect.DeepEqual(tt.expected, output) {
				t.Errorf("got: %+v, wanted: %+v", output, tt.expected)
			}
		})
	}
}

func TestFilterHiddenComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    []catalog.Catalog
		expected []catalog.Catalog
	}{
		{
			name:     "Case 1: empty input",
			input:    []catalog.Catalog{},
			expected: []catalog.Catalog{},
		},
		{
			name: "Case 2: non empty input",
			input: []catalog.Catalog{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1",
					},
					Spec: catalog.CatalogSpec{
						NonHiddenTags: []string{"1", "latest"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n2",
					},
					Spec: catalog.CatalogSpec{
						NonHiddenTags: []string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n3",
					},
					Spec: catalog.CatalogSpec{
						NonHiddenTags: []string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n4",
					},
					Spec: catalog.CatalogSpec{
						NonHiddenTags: []string{"10"},
					},
				},
			},
			expected: []catalog.Catalog{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n1",
					},
					Spec: catalog.CatalogSpec{
						NonHiddenTags: []string{"1", "latest"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "n4",
					},
					Spec: catalog.CatalogSpec{
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
