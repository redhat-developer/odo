package utils

import (
	"reflect"
	"testing"
)

func TestGetSupportedShells(t *testing.T) {
	tests := []struct {
		testName string
		shells   map[string]string
		expected []string
	}{
		{
			testName: "default",
			shells:   supportedShells,
			expected: []string{"bash", "zsh"},
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual := getSupportedShells()
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Errorf("expected: %+v, got: %+v", tt.expected, actual)
			}
		})
	}
}
