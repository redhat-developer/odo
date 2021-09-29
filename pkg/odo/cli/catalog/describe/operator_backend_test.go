package describe

import (
	"testing"

	"github.com/go-openapi/spec"
)

func TestGetTypeString(t *testing.T) {
	tests := []struct {
		name     string
		property spec.Schema
		want     string
	}{
		{
			name: "string type",
			property: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"string"},
				},
			},
			want: "string",
		},
		{
			name: "array of strings type",
			property: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
			want: "[]string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeString(tt.property)
			if result != tt.want {
				t.Errorf("Failed %s: got: %q, want: %q", t.Name(), result, tt.want)
			}
		})
	}
}
