package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestServicecatalog(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Servicecatalog Suite")
	// Keep CustomReporters commented till https://github.com/onsi/ginkgo/issues/628 is fixed
	// RunSpecsWithDefaultAndCustomReporters(t, "Servicecatalog Suite", []Reporter{reporter.JunitReport(t, "../../../reports")})
}
