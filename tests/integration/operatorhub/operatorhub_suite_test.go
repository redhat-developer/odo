package integration_test

import (
	"testing"

	"github.com/openshift/odo/v2/tests/helper"
)

func TestOperatorhub(t *testing.T) {
	helper.RunTestSpecs(t, "Operatorhub Suite")
}
