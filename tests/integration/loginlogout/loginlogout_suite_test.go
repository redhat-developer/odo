package integration

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestLoginlogout(t *testing.T) {
	helper.RunTestSpecs(t, "Login Logout Suite")
}
