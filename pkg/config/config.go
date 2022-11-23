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

func GetConfiguration() (*Configuration, error) {
	return GetConfigurationWith(envconfig.OsLookuper())
}

func GetConfigurationWith(lookuper envconfig.Lookuper) (*Configuration, error) {
	var s Configuration
	err := envconfig.ProcessWith(context.Background(), &s, lookuper)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
