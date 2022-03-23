package helper

import (
	"encoding/json"
	_ "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/segment"
	"io/ioutil"
)

func CreateTelemetryDebugFile() {
	tempFile, err := ioutil.TempFile("", "telemetery")
	Expect(err).NotTo(HaveOccurred())
	tempFile.WriteString("hello")
	Expect(segment.SetDebugTelemetry(tempFile.Name())).NotTo(HaveOccurred())
	Expect(tempFile.Close()).NotTo(HaveOccurred())
}

func GetTelemetryDebugData() segment.TelemetryData {
	var data []byte
	var td segment.TelemetryData
	telemetryFile := segment.GetDebugTelemetry()
	FileShouldEventuallyContainSubstring(telemetryFile, "event", 5)
	data, err := ioutil.ReadFile(telemetryFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(data, &td)).NotTo(HaveOccurred())
	return td
}
