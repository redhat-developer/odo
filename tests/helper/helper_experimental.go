package helper

import (
	"os"

	_ "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
)

// EnableExperimentalMode enables the experimental mode, so that experimental features of odo can be used.
func EnableExperimentalMode() {
	err := os.Setenv(feature.OdoExperimentalModeEnvVar, "true")
	Expect(err).ShouldNot(HaveOccurred())
}

// ResetExperimentalMode disables the experimental mode.
//
// Note that calling any experimental feature of odo right is expected to error out if experimental mode is not enabled.
func ResetExperimentalMode() {
	if _, ok := os.LookupEnv(feature.OdoExperimentalModeEnvVar); ok {
		err := os.Unsetenv(feature.OdoExperimentalModeEnvVar)
		Expect(err).ShouldNot(HaveOccurred())
	}
}
