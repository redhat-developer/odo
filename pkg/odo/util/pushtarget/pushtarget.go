package pushtarget

import (
	"os"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/preference"
)

// env variables
const (
	// Setting this env to `docker` will enable pushing to docker containers
	// and will override the setting in the preferences file.

	OdoPushTarget = "ODO_PUSHTARGET"
)

// IsPushTargetDocker checks if the push target preference has been set to docker
func IsPushTargetDocker() bool {

	// ODO's push target can be told to use docker by:
	//		- setting an env variable "ODO_PUSHTARGET=docker" or
	//		- setting odo preference using "odo preference set pushtarget docker"

	var (
		pushTargetEnv        string
		pushTargetPreference string
	)

	// Check "ODO_PUSHTARGET" env variable.
	pushTargetEnv, _ = os.LookupEnv(OdoPushTarget)
	if pushTargetEnv == "docker" || pushTargetEnv == "kube" {
		return pushTargetEnv == "docker"
	}

	// Fetch odo preferences and check if pushtarget is set
	cfg, err := preference.New()
	if err != nil {
		log.Errorf("failed to read odo preferences config. err: '%v'\n", err)
	} else {
		pushTargetPreference = cfg.GetPushTarget()
	}

	return pushTargetPreference == "docker"
}
