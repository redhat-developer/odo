package debug

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestDebug(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Debug Suite")
}
