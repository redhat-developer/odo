package recoverability

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestRecoverabilityScenarios(t *testing.T) {
	helper.RunTestSpecs(t, "odo recoverability scenarios")
}
