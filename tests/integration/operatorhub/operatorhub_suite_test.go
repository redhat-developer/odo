package integration_test

import (
	"testing"

	"github.com/redhat-developer/odo/tests/helper"
)

func TestOperatorhub(t *testing.T) {
	helper.RunTestSpecs(t, "Operatorhub Suite")
}
