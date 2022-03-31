package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = SynchronizedBeforeSuite(helper.TestSuiteBeforeAllSpecsFunc, helper.TestSuiteBeforeEachSpecFunc)

func TestLoginlogout(t *testing.T) {
	helper.RunTestSpecs(t, "Login Logout Suite")
}
