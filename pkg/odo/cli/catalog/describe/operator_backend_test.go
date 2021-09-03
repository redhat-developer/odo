package describe

import (
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
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

func TestToOpenAPISpec(t *testing.T) {
	tests := []struct {
		name string
		repr olm.CRDDescription
		want spec.Schema
	}{
		{
			name: "one-level property",
			repr: olm.CRDDescription{
				SpecDescriptors: []olm.SpecDescriptor{
					{
						Path:        "path1",
						DisplayName: "name to display 1",
						Description: "description 1",
					},
				},
			},
			want: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"path1": {
							SchemaProps: spec.SchemaProps{
								Type:        []string{"string"},
								Description: "description 1",
								Title:       "name to display 1",
							},
						},
					},
				},
			},
		},

		{
			name: "multiple-levels property",
			repr: olm.CRDDescription{
				SpecDescriptors: []olm.SpecDescriptor{
					{
						Path:        "subpath1.path1",
						DisplayName: "name to display 1.1",
						Description: "description 1.1",
					},
					{
						Path:        "subpath1.path2",
						DisplayName: "name to display 1.2",
						Description: "description 1.2",
					},
					{
						Path:        "subpath2.path1",
						DisplayName: "name to display 2.1",
						Description: "description 2.1",
					},
				},
			},
			want: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
					Properties: map[string]spec.Schema{
						"subpath1": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
								Properties: map[string]spec.Schema{
									"path1": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 1.1",
											Title:       "name to display 1.1",
										},
									},
									"path2": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 1.2",
											Title:       "name to display 1.2",
										},
									},
								},
							},
						},
						"subpath2": {
							SchemaProps: spec.SchemaProps{
								Type: []string{"object"},
								Properties: map[string]spec.Schema{
									"path1": {
										SchemaProps: spec.SchemaProps{
											Type:        []string{"string"},
											Description: "description 2.1",
											Title:       "name to display 2.1",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOpenAPISpec(&tt.repr)
			if !reflect.DeepEqual(*result, tt.want) {
				t.Errorf("Failed %s:\n\ngot: %+v\n\nwant: %+v", t.Name(), result, tt.want)
			}
		})
	}
}
