package helper

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	_ "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
)

const (
	DebugTelemetryFileEnv = "ODO_DEBUG_TELEMETRY_FILE"
)

func setDebugTelemetryFile(value string) error {
	return os.Setenv(DebugTelemetryFileEnv, value)
}

// EnableTelemetryDebug creates a temp file to use for debugging telemetry.
// it also sets up envs and cfg for the same
func EnableTelemetryDebug() {
	Expect(os.Setenv(segment.TrackingConsentEnv, "yes")).NotTo(HaveOccurred())

	ctx := context.Background()
	envConfig, err := config.GetConfiguration()
	Expect(err).To(BeNil())
	ctx = envcontext.WithEnvConfig(ctx, *envConfig)

	cfg, _ := preference.NewClient(ctx)
	err = cfg.SetConfiguration(preference.ConsentTelemetrySetting, "true")
	Expect(err).To(BeNil())
	tempFile, err := ioutil.TempFile("", "telemetry")
	Expect(err).NotTo(HaveOccurred())
	Expect(setDebugTelemetryFile(tempFile.Name())).NotTo(HaveOccurred())
	Expect(tempFile.Close()).NotTo(HaveOccurred())
}

func GetDebugTelemetryFile() string {
	return os.Getenv(DebugTelemetryFileEnv)
}

// GetTelemetryDebugData gets telemetry data dumped into temp file for testing/debugging
func GetTelemetryDebugData() segment.TelemetryData {
	var data []byte
	var td segment.TelemetryData
	telemetryFile := GetDebugTelemetryFile()
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

// ResetTelemetry resets the telemetry back to original values
func ResetTelemetry() {
	Expect(os.Setenv(segment.TrackingConsentEnv, "no")).NotTo(HaveOccurred())
	Expect(os.Unsetenv(DebugTelemetryFileEnv))

	ctx := context.Background()
	envConfig, err := config.GetConfiguration()
	Expect(err).To(BeNil())
	ctx = envcontext.WithEnvConfig(ctx, *envConfig)

	cfg, _ := preference.NewClient(ctx)
	err = cfg.SetConfiguration(preference.ConsentTelemetrySetting, "true")
	Expect(err).NotTo(HaveOccurred())
	Expect(segment.IsTelemetryEnabled(cfg, *envConfig)).To(BeFalse())
}
