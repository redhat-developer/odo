package kclient

import (
	"testing"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/google/go-cmp/cmp"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

func TestGetResourceSpecDefinitionFromSwagger(t *testing.T) {
	tests := []struct {
		name    string
		swagger []byte
		group   string
		version string
		kind    string
		want    *spec.Schema
		wantErr bool
	}{
		{
			name:    "not found CRD",
			swagger: []byte("{}"),
			group:   "aGroup",
			version: "aVersion",
			kind:    "aKind",
			want:    nil,
			wantErr: true,
		},
		{
			name: "found CRD without spec",
			swagger: []byte(`{
  "definitions": {
	"com.dev4devs.postgresql.v1alpha1.Database": {
		"type": "object",
		"x-kubernetes-group-version-kind": [
			{
				"group": "postgresql.dev4devs.com",
				"kind": "Database",
				"version": "v1alpha1"
			}
		]
	}
  }
}`),
			group:   "postgresql.dev4devs.com",
			version: "v1alpha1",
			kind:    "Database",
			want:    nil,
			wantErr: false,
		},
		{
			name: "found CRD with spec",
			swagger: []byte(`{
  "definitions": {
	"com.dev4devs.postgresql.v1alpha1.Database": {
		"type": "object",
		"x-kubernetes-group-version-kind": [
			{
				"group": "postgresql.dev4devs.com",
				"kind": "Database",
				"version": "v1alpha1"
			}
		],
		"properties": {
			"spec": {
				"type": "object"
			}
		}
	}
  }
}`),
			group:   "postgresql.dev4devs.com",
			version: "v1alpha1",
			kind:    "Database",
			want: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: []string{"object"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getResourceSpecDefinitionFromSwagger(tt.swagger, tt.group, tt.version, tt.kind)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(jsonreference.Ref{}, jsonpointer.Pointer{})); diff != "" {
				t.Errorf("getResourceSpecDefinitionFromSwagger mismatch (-want +got):\n%s", diff)
			}
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, gotErr)
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
					AdditionalProperties: &spec.SchemaOrBool{
						Allows: false,
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
					AdditionalProperties: &spec.SchemaOrBool{
						Allows: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOpenAPISpec(&tt.repr)
			if diff := cmp.Diff(tt.want, *result, cmp.AllowUnexported(jsonreference.Ref{}, jsonpointer.Pointer{})); diff != "" {
				t.Errorf("toOpenAPISpec mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
