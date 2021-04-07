// +build !race

package integration

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestIntegration(t *testing.T) {
	helper.RunTestSpecs(t, "Integration Suite")
}
