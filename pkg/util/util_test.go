package util

import "testing"

func TestNamespaceOpenShiftObject(t *testing.T) {

	name, err := NamespaceOpenShiftObject("bar", "foo")
	if err != nil {
		t.Fatalf("Error with namespacing: %s", err)
	}

	if name != "foo-bar" {
		t.Error("Expected foo-bar")
		t.Errorf("Actual output: %s", name)
	}

}
