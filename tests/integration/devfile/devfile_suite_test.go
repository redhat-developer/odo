package devfile

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestDevfiles(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Suite")
}
