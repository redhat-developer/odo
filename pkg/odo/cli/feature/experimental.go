package feature

import (
	"context"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
)

const OdoExperimentalModeEnvVar = "ODO_EXPERIMENTAL_MODE"

// IsExperimentalModeEnabled returns whether the experimental mode is enabled or not,
// which means by checking the value of the "ODO_EXPERIMENTAL_MODE" environment variable.
func IsExperimentalModeEnabled(ctx context.Context) bool {
	return envcontext.GetEnvConfig(ctx).OdoExperimentalMode
}
