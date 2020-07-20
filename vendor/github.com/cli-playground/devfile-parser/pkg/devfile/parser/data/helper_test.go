package data

import (
	"reflect"
	"strings"
	"testing"

	v100 "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/1.0.0"
)

func TestNewDevfileData(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version  = apiVersion100
			want     = reflect.TypeOf(&v100.Devfile100{})
			obj, err = NewDevfileData(string(version))
			got      = reflect.TypeOf(obj)
		)

		// got and want should be equal
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got: '%v', want: '%s'", got, want)
		}

		// no error should be received
		if err != nil {
			t.Errorf("did not expect an error '%v'", err)
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			_, err  = NewDevfileData(string(version))
		)

		// no error should be received
		if err == nil {
			t.Errorf("did not expect an error '%v'", err)
		}
	})
}

func TestGetDevfileJSONSchema(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version  = apiVersion100
			want     = v100.JsonSchema100
			got, err = GetDevfileJSONSchema(string(version))
		)

		if err != nil {
			t.Errorf("did not expect an error '%v'", err)
		}

		if strings.Compare(got, want) != 0 {
			t.Errorf("incorrect json schema")
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			_, err  = GetDevfileJSONSchema(string(version))
		)

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}

func TestIsApiVersionSupported(t *testing.T) {

	t.Run("valid devfile apiVersion", func(t *testing.T) {

		var (
			version = apiVersion100
			want    = true
			got     = IsApiVersionSupported(string(version))
		)

		if got != want {
			t.Errorf("want: '%t', got: '%t'", want, got)
		}
	})

	t.Run("invalid devfile apiVersion", func(t *testing.T) {

		var (
			version = "invalidVersion"
			want    = false
			got     = IsApiVersionSupported(string(version))
		)

		if got != want {
			t.Errorf("expected an error, didn't get one")
		}
	})
}
