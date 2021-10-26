package devfile

import (
	"testing"

	"github.com/openshift/odo/v2/tests/helper"
)

func TestDevfiles(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Suite")
}
