package helper

import (
	"encoding/json"
	_ "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"io/ioutil"
	"os"
)

//CreateTelemetryDebugFile creates a temp file to use for debugging telemetry.
//it also sets up envs and cfg for the same
func CreateTelemetryDebugFile() {
	Expect(os.Setenv(segment.DisableTelemetryEnv, "false")).NotTo(HaveOccurred())
	cfg, _ := preference.NewClient()
	err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "true")
	Expect(err).To(BeNil())
	tempFile, err := ioutil.TempFile("", "telemetry")
	Expect(err).NotTo(HaveOccurred())
	Expect(segment.SetDebugTelemetryFile(tempFile.Name())).NotTo(HaveOccurred())
	Expect(tempFile.Close()).NotTo(HaveOccurred())
}

//GetTelemetryDebugData gets telemetry data dumped into temp file for testing/debugging
func GetTelemetryDebugData() segment.TelemetryData {
	Expect(os.Setenv(segment.DisableTelemetryEnv, "true")).NotTo(HaveOccurred())
	var data []byte
	var td segment.TelemetryData
	telemetryFile := segment.GetDebugTelemetryFile()
	Eventually(func() string {
		data, err := ioutil.ReadFile(telemetryFile)
		Expect(err).To(BeNil())
		return string(data)
	}, 10, 1).Should(ContainSubstring("event"))
	data, err := ioutil.ReadFile(telemetryFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(data, &td)).NotTo(HaveOccurred())
	return td
}
