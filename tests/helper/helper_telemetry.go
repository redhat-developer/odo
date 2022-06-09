package helper

import (
	"encoding/json"
	"io/ioutil"
	"os"

	_ "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
)

func setDebugTelemetryFile(value string) error {
	return os.Setenv(segment.DebugTelemetryFileEnv, value)
}

//EnableTelemetryDebug creates a temp file to use for debugging telemetry.
//it also sets up envs and cfg for the same
func EnableTelemetryDebug() {
	Expect(os.Setenv(segment.DisableTelemetryEnv, "false")).NotTo(HaveOccurred())
	cfg, _ := preference.NewClient()
	err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "true")
	Expect(err).To(BeNil())
	tempFile, err := ioutil.TempFile("", "telemetry")
	Expect(err).NotTo(HaveOccurred())
	Expect(setDebugTelemetryFile(tempFile.Name())).NotTo(HaveOccurred())
	Expect(tempFile.Close()).NotTo(HaveOccurred())
}

//GetTelemetryDebugData gets telemetry data dumped into temp file for testing/debugging
func GetTelemetryDebugData() segment.TelemetryData {
	var data []byte
	var td segment.TelemetryData
	telemetryFile := segment.GetDebugTelemetryFile()
	Eventually(func() string {
		d, err := ioutil.ReadFile(telemetryFile)
		Expect(err).To(BeNil())
		return string(d)
	}, 10, 1).Should(ContainSubstring("event"))
	data, err := ioutil.ReadFile(telemetryFile)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(data, &td)).NotTo(HaveOccurred())
	return td
}

//ResetTelemetry resets the telemetry back to original values
func ResetTelemetry() {
	Expect(os.Setenv(segment.DisableTelemetryEnv, "true")).NotTo(HaveOccurred())
	Expect(os.Unsetenv(segment.DebugTelemetryFileEnv))
	cfg, _ := preference.NewClient()
	err := cfg.SetConfiguration(preference.ConsentTelemetrySetting, "true")
	Expect(err).NotTo(HaveOccurred())
	Expect(segment.IsTelemetryEnabled(cfg)).To(BeFalse())
}
