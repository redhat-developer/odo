package integration_test

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestOperatorhub(t *testing.T) {
	helper.RunTestSpecs(t, "Operatorhub Suite")
}
