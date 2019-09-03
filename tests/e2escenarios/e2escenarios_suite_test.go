package e2escenarios

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestE2eScenarios(t *testing.T) {
	RegisterFailHandler(Fail)
	time := time.Now()
	xmlFileName := fmt.Sprintf("../../reports/junit_%d-%d-%d_%02d-%02d-%02d.xml", time.Year(), time.Month(),
		time.Day(), time.Hour(), time.Minute(), time.Second())
	junitReporter := reporters.NewJUnitReporter(xmlFileName)
	RunSpecsWithDefaultAndCustomReporters(t, "odo e2e scenarios", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	if _, err := os.Stat("../../reports"); os.IsNotExist(err) {
		os.Mkdir("../../reports", os.ModePerm)
	}
})
