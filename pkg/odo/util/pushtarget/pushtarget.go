package pushtarget

// env variables
const (
	// Setting this env to `docker` will enable pushing to docker containers
	// and will override the setting in the preferences file.

	OdoPushTarget = "ODO_PUSH_TARGET"
)

// IsPushTargetDocker checks if the push target preference has been set to docker
// Currently hardcoded to return false
func IsPushTargetDocker() bool {

	return false
}
