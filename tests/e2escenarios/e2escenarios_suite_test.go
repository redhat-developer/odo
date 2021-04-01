package e2escenarios

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestE2eScenarios(t *testing.T) {
	helper.RunTestSpecs(t, "odo e2e scenarios")
}
