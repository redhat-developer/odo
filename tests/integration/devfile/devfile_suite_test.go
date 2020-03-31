package devfile

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDevfiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devfile Suite")
	// Keep CustomReporters commented till https://github.com/onsi/ginkgo/issues/628 is fixed
	// RunSpecsWithDefaultAndCustomReporters(t, "Project Suite", []Reporter{reporter.JunitReport(t, "../../../reports")})
}
