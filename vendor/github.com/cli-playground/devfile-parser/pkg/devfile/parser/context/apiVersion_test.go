package parser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSetDevfileAPIVersion(t *testing.T) {

	const (
		apiVersion          = "1.0.0"
		validJson           = `{"apiVersion": "1.0.0"}`
		emptyJson           = "{}"
		emptyApiVersionJson = `{"apiVersion": ""}`
	)

	// test table
	tests := []struct {
		name    string
		rawJson []byte
		want    string
		wantErr error
	}{
		{
			name:    "valid apiVersion",
			rawJson: []byte(validJson),
			want:    apiVersion,
			wantErr: nil,
		},
		{
			name:    "apiVersion not present",
			rawJson: []byte(emptyJson),
			want:    "",
			wantErr: fmt.Errorf("apiVersion or schemaVersion not present in devfile"),
		},
		{
			name:    "apiVersion empty",
			rawJson: []byte(emptyApiVersionJson),
			want:    "",
			wantErr: fmt.Errorf("apiVersion in devfile cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// new devfile context object
			d := DevfileCtx{rawContent: tt.rawJson}

			// SetDevfileAPIVersion
			gotErr := d.SetDevfileAPIVersion()
			got := d.apiVersion

			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("unexpected error: '%v', wantErr: '%v'", gotErr, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("want: '%v', got: '%v'", tt.want, got)
			}
		})
	}
}

func TestGetApiVersion(t *testing.T) {

	const (
		apiVersion = "1.0.0"
	)

	t.Run("get apiVersion", func(t *testing.T) {

		var (
			d    = DevfileCtx{apiVersion: apiVersion}
			want = apiVersion
			got  = d.GetApiVersion()
		)

		if got != want {
			t.Errorf("want: '%v', got: '%v'", want, got)
		}
	})
}
