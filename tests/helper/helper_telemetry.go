package helper

import (
	"encoding/json"
	_ "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/segment"
	"io/ioutil"
	"os"
)

func CreateTelemetryDebugFile() {
	tempFile, err := ioutil.TempFile("", "telemetery")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Setenv(segment.DebugTelemetryFileEnv, tempFile.Name()))
	Expect(tempFile.Close()).NotTo(HaveOccurred())
}

func GetTelemetryDebugData() segment.TelemetryData {
	data, err := ioutil.ReadFile(segment.GetDebugTelemetry())
	Expect(err).NotTo(HaveOccurred())
	var td segment.TelemetryData
	Expect(json.Unmarshal(data, &td)).NotTo(HaveOccurred())
	return td
}
