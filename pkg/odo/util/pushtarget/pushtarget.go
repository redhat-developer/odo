package pushtarget

import (
	"os"

	"github.com/openshift/odo/pkg/preference"
	"k8s.io/klog"
)

// env variables
const (
	// Setting this env to `docker` will enable pushing to docker containers
	// and will override the setting in the preferences file.

	OdoPushTarget = "ODO_PUSH_TARGET"
)

// IsPushTargetDocker checks if the push target preference has been set to docker
func IsPushTargetDocker() bool {

	// ODO's push target can be told to use docker by:
	//		- setting an env variable "ODO_PUSH_TARGET=docker" or
	//		- setting odo preference using "odo preference set pushtarget docker"

	var (
		pushTargetEnv string
	)

	// Check "ODO_PUSH_TARGET" env variable.
	pushTargetEnv, _ = os.LookupEnv(OdoPushTarget)
	if pushTargetEnv == preference.DockerPushTarget || pushTargetEnv == preference.KubePushTarget {
		return pushTargetEnv == preference.DockerPushTarget
	} else if pushTargetEnv != "" {
		// Log an error and return false if an invalid value was passed in to env var and return false
		klog.Error("Invalid value passed in to ")
		return false
	}

	// Fetch odo preferences and check if pushtarget is set
	cfg, err := preference.New()
	if err != nil {
		klog.Errorf("failed to read odo preferences config. err: '%v'\n", err)
		return false
	}

	return cfg.GetPushTarget() == preference.DockerPushTarget
}
