package segment

import (
	"context"

	"github.com/Xuanwo/go-locale"

	registryLibrary "github.com/devfile/registry-support/registry-library/library"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"k8s.io/klog"
)

// getTelemetryForDevfileRegistry returns a populated TelemetryData object that contains some odo telemetry (with client consent), such as the anonymous ID and
// locale in addition to the generic client name "odo"
func getTelemetryForDevfileRegistry(ctx context.Context) (registryLibrary.TelemetryData, error) {

	td := registryLibrary.TelemetryData{
		Client: TelemetryClient,
	}

	envConfig := envcontext.GetEnvConfig(ctx)

	if envConfig.TelemetryCaller != "" {
		td.Client += "-" + envConfig.TelemetryCaller
	}

	if envConfig.OdoDebugTelemetryFile != nil {
		return td, nil
	}

	if !scontext.GetTelemetryStatus(ctx) {
		return td, nil
	}

	// if there is an error, tag will be undetermined
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
func GetRegistryOptions(ctx context.Context) registryLibrary.RegistryOptions {
	td, err := getTelemetryForDevfileRegistry(ctx)
	if err != nil {
		// this error should not prevent basic telemetry from being sent
		klog.Errorf("An error prevented additional telemetry to be set %v", err)
	}

	registryOptions := registryLibrary.RegistryOptions{
		SkipTLSVerify: false,
		Telemetry:     td,
		Filter:        registryLibrary.RegistryFilter{},
	}

	return registryOptions
}
