package experimental

import (
	"os"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/preference"
)

// env variables
const (
	// Setting this env to true will expose experimental features to the user
	OdoExperimentalEnv = "ODO_EXPERIMENTAL"
)

// IsExperimentalModeEnabled checks if the experimental mode has been enabled
// via env variable or via odo preferences and returns boolean value.
// By default experimental mode is disabled
func IsExperimentalModeEnabled() bool {

	// Experimental mode can be set by:
	//		- setting an env variable "ODO_EXPERIMENTAL=true" or
	//		- setting odo preference using "odo preference set experimental true"

	var (
		experimentalPreference bool = false
		experimentalEnv        bool = false
	)

	// Fetch odo preferences and check if experimental mode is set
	cfg, err := preference.New()
	if err != nil {
		log.Errorf("failed to read odo preferences config. err: '%v'\n", err)
	} else {
		experimentalPreference = cfg.GetExperimental()
	}

	// Check "ODO_EXPERIMENTAL" env variable
	experimentalEnvStr, _ := os.LookupEnv(OdoExperimentalEnv)
	if experimentalEnvStr == "true" {
		experimentalEnv = true
	}

	return (experimentalPreference || experimentalEnv)
}
