package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoginlogout(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loginlogout Suite")
	// Keep CustomReporters commented till https://github.com/onsi/ginkgo/issues/628 is fixed
	// RunSpecsWithDefaultAndCustomReporters(t, "Loginlogout Suite", []Reporter{reporter.JunitReport(t, "../../../reports")})
}
