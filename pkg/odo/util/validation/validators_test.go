package validation

import (
	"strings"
	"testing"
)

func TestIntegerValidator(t *testing.T) {
	err := IntegerValidator(1)
	if err != nil {
		t.Error("integer validator should validate integers")
	}

	err = IntegerValidator("1")
	if err != nil {
		t.Error("integer validator should validate integers as string")
	}

	err = IntegerValidator(new(interface{}))
	if err == nil {
		t.Error("integer validator shouldn't validate unknown types")
	} else {
		if !strings.Contains(err.Error(), "don't know how to convert") {
			t.Error("integer validator should report error that it can't convert unknown type")
		}
	}
}

func TestNilValidator(t *testing.T) {
	err := NilValidator(new(interface{}))
	if err != nil {
		t.Error("nil validator should always validate any input")
	}

	err = NilValidator(nil)
	if err != nil {
		t.Error("nil validator should always validate even nil")
	}
}
