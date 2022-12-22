package integration

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestDocumentation(t *testing.T) {
	helper.RunTestSpecs(t, "Documentation Suite")
}
