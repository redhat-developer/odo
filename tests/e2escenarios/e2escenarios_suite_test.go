package e2escenarios

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = SynchronizedBeforeSuite(helper.TestSuiteBeforeAllSpecsFunc, helper.TestSuiteBeforeEachSpecFunc)

func TestE2eScenarios(t *testing.T) {
	helper.RunTestSpecs(t, "odo e2e scenarios")
}
