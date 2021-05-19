package integration

import (
	"testing"

	"github.com/openshift/odo/tests/helper"
)

func TestLoginlogout(t *testing.T) {
	helper.CmdShouldPass("oc", "login", "-u", "developer", "-p", "password@123")
	helper.RunTestSpecs(t, "Login Logout Suite")
}
