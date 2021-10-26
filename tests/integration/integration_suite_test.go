// +build !race

package integration

import (
	"testing"

	"github.com/openshift/odo/v2/tests/helper"
)

func TestIntegration(t *testing.T) {
	helper.RunTestSpecs(t, "Integration Suite")
}
