package feature

import (
	"os"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/log"
)

const (
	OdoExperimentalModeEnvVar = "ODO_EXPERIMENTAL_MODE"
	OdoExperimentalModeTrue   = "true"
)

var _experimentalFeatures []OdoFeature

// IsExperimental returns whether the given feature can be used as an experimental CLI feature,
// which means that the following two conditions should be met: i) the ODO_EXPERIMENTAL_MODE environment variable is enabled,
// and ii) the provided feature is explicitly listed as experimental in the list of supported experimental features.
func IsExperimental(feat OdoFeature) bool {
	if os.Getenv(OdoExperimentalModeEnvVar) != OdoExperimentalModeTrue {
		return false
	}

	for _, expFeat := range _experimentalFeatures {
		if expFeat.id == feat.id {
			klog.V(4).Infof("Feature %q is marked as experimental", feat.id)
			log.Experimentalf("Experimental mode enabled for %s. Use at your own risk. More details on https://odo.dev/docs/user-guides/advanced/experimental-mode",
				feat.description)
			return true
		}
	}

	klog.V(4).Infof("Feature %q not found in the known list of experimental features.", feat.id)
	return false
}
