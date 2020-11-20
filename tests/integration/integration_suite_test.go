// +build !race

package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper/reporter"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Integration Suite", []Reporter{reporter.JunitReport(t, "../../reports/")})
}
