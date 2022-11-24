package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Configuration struct {
	DevfileProxy          *string `env:"DEVFILE_PROXY,noinit"`
	DockerCmd             string  `env:"DOCKER_CMD,default=docker"`
	Globalodoconfig       *string `env:"GLOBALODOCONFIG,noinit"`
	OdoDebugTelemetryFile *string `env:"ODO_DEBUG_TELEMETRY_FILE,noinit"`
	OdoDisableTelemetry   *bool   `env:"ODO_DISABLE_TELEMETRY,noinit"`
	OdoLogLevel           *int    `env:"ODO_LOG_LEVEL,noinit"`
	OdoTrackingConsent    *string `env:"ODO_TRACKING_CONSENT,noinit"`
	PodmanCmd             string  `env:"PODMAN_CMD,default=podman"`
	TelemetryCaller       string  `env:"TELEMETRY_CALLER,default="`
	OdoExperimentalMode   bool    `env:"ODO_EXPERIMENTAL_MODE,default=false"`
}

// GetConfiguration initializes a Configuration for odo by using the system environment.
// See GetConfigurationWith for a more configurable version.
func GetConfiguration() (*Configuration, error) {
	return GetConfigurationWith(envconfig.OsLookuper())
}

// GetConfigurationWith initializes a Configuration for odo by using the specified envconfig.Lookuper to resolve values.
// It is recommended to use this function (instead of GetConfiguration) if you don't need to depend on the current system environment,
// typically in unit tests.
func GetConfigurationWith(lookuper envconfig.Lookuper) (*Configuration, error) {
	var s Configuration
	err := envconfig.ProcessWith(context.Background(), &s, lookuper)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
