package kclient

import (
	"reflect"
	"testing"

	"github.com/go-openapi/spec"
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %+v\n\ngot %+v", tt.want, got)
			}
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, gotErr)
			}
		})
	}
}
