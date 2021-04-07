package devfile

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestDevfiles(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Suite")
}
