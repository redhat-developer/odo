package feature

import (
	"os"
)

const (
	OdoExperimentalModeEnvVar = "ODO_EXPERIMENTAL_MODE"
	OdoExperimentalModeTrue   = "true"
)

func isExperimentalModeEnabled() bool {
	return os.Getenv(OdoExperimentalModeEnvVar) == OdoExperimentalModeTrue
}
