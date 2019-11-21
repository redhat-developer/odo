package e2escenarios

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2eScenarios(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "odo e2e scenarios")
	// Keep CustomReporters commented till https://github.com/onsi/ginkgo/issues/628 is fixed
	//RunSpecsWithDefaultAndCustomReporters(t, "odo e2e scenarios", []Reporter{reporter.JunitReport(t, "../../reports")})
}
