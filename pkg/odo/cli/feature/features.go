package feature

import (
	"context"
	"sort"

	"github.com/redhat-developer/odo/pkg/log"
)

// OdoFeature represents a uniquely identifiable feature of odo.
// It can either be a CLI command or flag.
type OdoFeature struct {
	// id is a free-form but unique identifier of this feature.
	id string

	// isExperimental indicates whether this feature should be considered in early or intermediate stages of development.
	// Features that are not experimental by default will always be enabled, regardless of the experimental mode.
	isExperimental bool

	// description provides a human-readable overview of this feature.
	// Note that this will be visible by the end users if this feature is experimental and the experimental mode is enabled.
	description string
}

var (
	// GenericRunOnFlag is the feature supporting the `--run-on` generic CLI flag.
	GenericRunOnFlag = OdoFeature{
		id:             "generic-run-on",
		isExperimental: true,
		description:    "flag: --run-on",
	}

	enabledFeatures = map[OdoFeature]struct{}{}
)

// IsEnabled returns whether the specified feature should be enabled or not.
// If the feature is not marked as experimental, it should always be enabled.
// Otherwise, it is enabled only if the experimental mode is enabled (see the isExperimentalModeEnabled package-level function).
func IsEnabled(ctx context.Context, feat OdoFeature) bool {
	// Features not marked as experimental are always enabled, regardless of the experimental mode
	if !feat.isExperimental {
		return true
	}

	// Features marked as experimental are enabled only if the experimental mode is set
	experimentalModeEnabled := isExperimentalModeEnabled(ctx)
	if experimentalModeEnabled {
		enabledFeatures[feat] = struct{}{}
	}
	return experimentalModeEnabled
}

func DisplayWarnings() {
	features := make([]OdoFeature, 0, len(enabledFeatures))
	for k := range enabledFeatures {
		features = append(features, k)
	}
	sort.Slice(features, func(i, j int) bool {
		return features[i].id < features[j].id
	})
	for _, feat := range features {
		log.Experimentalf("Experimental mode enabled for %s. Use at your own risk. More details on https://odo.dev/docs/user-guides/advanced/experimental-mode",
			feat.description)
	}
}
