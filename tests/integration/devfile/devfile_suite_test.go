package devfile

import (
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/redhat-developer/odo/tests/helper"
)

var _ = SynchronizedBeforeSuite(helper.TestSuiteBeforeAllSpecsFunc, helper.TestSuiteBeforeEachSpecFunc)

func TestDevfiles(t *testing.T) {
	helper.RunTestSpecs(t, "Devfile Suite")
}
