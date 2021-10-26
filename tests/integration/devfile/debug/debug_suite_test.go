package debug

import (
	"testing"

	"github.com/openshift/odo/v2/tests/helper"
)

func TestDebug(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Debug Suite")
}
