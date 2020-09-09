package parser

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPopulateFromBytes(t *testing.T) {

	const (
		InvalidURL = "blah"
	)

	t.Run("valid data passed", func(t *testing.T) {

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(validJsonRawContent200())
			if err != nil {
				t.Error(err)
			}
		}))

		var (
			d = DevfileCtx{
				url: testServer.URL,
			}
		)
		defer testServer.Close()

		err := d.PopulateFromURL()

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}
	})

	t.Run("invalid data passed", func(t *testing.T) {

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(InvalidDevfileContent))
			if err != nil {
				t.Error(err)
			}
		}))

		var (
			d = DevfileCtx{
				url: testServer.URL,
			}
		)
		defer testServer.Close()

		err := d.PopulateFromURL()

		if err == nil {
			t.Errorf("expected error, didn't get one ")
		}
	})

	t.Run("invalid filepath", func(t *testing.T) {

		var (
			d = DevfileCtx{
				url: InvalidURL,
			}
		)

		err := d.PopulateFromURL()

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}
