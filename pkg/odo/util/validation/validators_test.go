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

func TestNameValidator(t *testing.T) {
	// note that we're just testing a single case here since presumably the underlying implementation is already tested in k8s
	err := NameValidator("some-valid-name")
	if err != nil {
		t.Errorf("name validator should have accepted name, but got: %v instead", err)
	}

	err = NameValidator(new(interface{}))
	if err == nil {
		t.Error("name validator should only attempt to validate non-nil strings")
	} else {
		if !strings.Contains(err.Error(), "can only validate strings") {
			t.Error("name validator should report error that it can only valida strings")
		}
	}
}
