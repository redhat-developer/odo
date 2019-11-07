package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo/reporters"
)

// JunitReport takes test object and filepath as argument, returns junitReporter object
func JunitReport(t *testing.T, filePath string) *reporters.JUnitReporter {
	time := time.Now()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_ = os.Mkdir(filePath, os.ModePerm)
	}
	xmlFileName := fmt.Sprintf(filepath.Join(filePath, "junit_%d-%d-%d_%02d-%02d-%02d.xml"), time.Year(), time.Month(),
		time.Day(), time.Hour(), time.Minute(), time.Second())
	junitReporter := reporters.NewJUnitReporter(xmlFileName)
	return junitReporter
}
