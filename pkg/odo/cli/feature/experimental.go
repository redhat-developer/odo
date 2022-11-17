package feature

import (
	"context"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
)

const (
	OdoExperimentalModeEnvVar = "ODO_EXPERIMENTAL_MODE"
	OdoExperimentalModeTrue   = "true"
)

func isExperimentalModeEnabled(ctx context.Context) bool {
	return envcontext.GetEnvConfig(ctx).OdoExperimentalMode
}
