package feature

import (
	"context"
)

// OdoFeature represents a uniquely identifiable feature of odo.
// It can either be a CLI command or flag.
type OdoFeature struct {
	// isExperimental indicates whether this feature should be considered in early or intermediate stages of development.
	// Features that are not experimental by default will always be enabled, regardless of the experimental mode.
	isExperimental bool
}

var (
	// GenericRunOnFlag is the feature supporting the `--run-on` generic CLI flag.
	GenericRunOnFlag = OdoFeature{
		isExperimental: true,
	}
)

// IsEnabled returns whether the specified feature should be enabled or not.
// If the feature is not marked as experimental, it should always be enabled.
// Otherwise, it is enabled only if the experimental mode is enabled (see the IsExperimentalModeEnabled package-level function).
func IsEnabled(ctx context.Context, feat OdoFeature) bool {
	// Features not marked as experimental are always enabled, regardless of the experimental mode
	if !feat.isExperimental {
		return true
	}

	// Features marked as experimental are enabled only if the experimental mode is set
	return IsExperimentalModeEnabled(ctx)
}
