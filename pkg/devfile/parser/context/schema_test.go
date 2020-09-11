package parser

import (
	"testing"

	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
)

func TestValidateDevfileSchema(t *testing.T) {

	t.Run("valid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: validJsonRawContent200(),
			}
		)

		err := d.ValidateDevfileSchema()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})

	t.Run("invalid 2.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v200.JsonSchema200,
				rawContent: []byte("{}"),
			}
		)

		err := d.ValidateDevfileSchema()
		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})
}
