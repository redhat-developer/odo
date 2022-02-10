package segment

import (
	"github.com/Xuanwo/go-locale"
	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	registryConsts "github.com/redhat-developer/odo/pkg/odo/cli/preference/registry/consts"
	"github.com/redhat-developer/odo/pkg/preference"
	"k8s.io/klog"
)

// getTelemetryForDevfileRegistry returns a populated TelemetryData object that contains some odo telemetry (with client consent), such as the anonymous ID and
// locale in addition to the generic client name "odo"

func getTelemetryForDevfileRegistry() (registryLibrary.TelemetryData, error) {

	td := registryLibrary.TelemetryData{
		Client: registryConsts.TelemetryClient,
	}

	// TODO(feloy) Get from DI
	cfg, err := preference.NewClient()
	if err != nil {
		return td, err
	}

	if !IsTelemetryEnabled(cfg) {
		return td, nil
	}

	//if there is an error, tag will be undetermined
	tag, _ := locale.Detect()
	td.Locale = tag.String()

	user, err := getUserIdentity(GetTelemetryFilePath())
	if err != nil {
		// default to the generic user ID if the anonymous ID cannot be found
		td.User = td.Client
		return td, err
	}

	td.User = user
	return td, nil

}

// GetRegistryOptions returns a populated RegistryOptions object containing all the properties needed to make a devfile registry library call
func GetRegistryOptions() registryLibrary.RegistryOptions {
	td, err := getTelemetryForDevfileRegistry()
	if err != nil {
		//this error should not prevent basic telemetry from being sent
		klog.Errorf("An error prevented additional telemetry to be set %v", err)
	}

	registryOptions := registryLibrary.RegistryOptions{
		SkipTLSVerify: false,
		Telemetry:     td,
		Filter:        registryLibrary.RegistryFilter{},
	}

	return registryOptions
}
