package feature

import (
	"context"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"k8s.io/utils/pointer"
)

const (
	OdoExperimentalModeEnvVar = "ODO_EXPERIMENTAL_MODE"
	OdoExperimentalModeTrue   = "true"
)

func isExperimentalModeEnabled(ctx context.Context) bool {
	return pointer.BoolDeref(envcontext.GetEnvConfig(ctx).OdoExperimentalMode, false)
}
