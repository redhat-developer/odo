package benchmark

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/odo/tests/helper/reporter"
)

func TestBenchmark(t *testing.T) {
	RegisterFailHandler(Fail)

	reporters := []Reporter{}

	fmt.Printf("os.Getenv(PULL_NUMBER) = %#v\n", os.Getenv("PULL_NUMBER"))

	// only when executed on OpenShift CI
	if os.Getenv("CI") == "openshift" {
		// If running on OpenShift CI, add reporter that will submit measurements to Google Sheets table
		// https://docs.google.com/spreadsheets/d/1o-GIoYlZoEyW1F25kAwvvEXzbW4tLck0x3EGCC4_L_A/
		reporters = append(reporters, reporter.NewHTTPMeasurementReporter("https://script.google.com/macros/s/AKfycbyoeFrEXsrkjWOCCjsLOGY5a31Fsv5RTUvgqQP0E5vPo3YvDGE/exec"))
	}
	RunSpecsWithDefaultAndCustomReporters(t, "odo benchmark tests", reporters)
}
