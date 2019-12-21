package parser

import "testing"

func TestSetDevfileAPIVersion(t *testing.T) {

	const (
		apiVersion          = "1.0.0"
		validJson           = `{"apiVersion": "1.0.0"}`
		emptyJson           = "{}"
		emptyApiVersionJson = `{"apiVersion":}`
	)

	t.Run("valid apiVersion", func(t *testing.T) {

		var (
			rawJson = []byte(validJson)
			d       = DevfileCtx{rawContent: rawJson}
			err     = d.SetDevfileAPIVersion()
			got     = d.apiVersion
			want    = apiVersion
		)

		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}

		if got != want {
			t.Errorf("want: '%v', got: '%v'", want, got)
		}
	})

	t.Run("apiVersion not present", func(t *testing.T) {

		var (
			rawJson = []byte(emptyJson)
			d       = DevfileCtx{rawContent: rawJson}
			err     = d.SetDevfileAPIVersion()
		)

		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})

	t.Run("apiVersion empty", func(t *testing.T) {

		var (
			rawJson = []byte(emptyApiVersionJson)
			d       = DevfileCtx{rawContent: rawJson}
			err     = d.SetDevfileAPIVersion()
		)

		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})
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
