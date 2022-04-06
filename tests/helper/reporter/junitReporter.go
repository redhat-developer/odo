package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
)

// JunitReport takes test object and filepath as argument, returns junitReporter object
func JunitReport(filePath string) *reporters.JUnitReporter {
	time := time.Now()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_ = os.Mkdir(filePath, os.ModePerm)
	}
	xmlFileName := fmt.Sprintf(filepath.Join(filePath, "junit_%d-%d-%d_%02d-%02d-%02d_%d.xml"), time.Year(), time.Month(),
		time.Day(), time.Hour(), time.Minute(), time.Second(), config.GinkgoConfig.ParallelNode)
	junitReporter := reporters.NewJUnitReporter(xmlFileName)
	return junitReporter
}
