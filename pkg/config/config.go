package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Configuration struct {
	DevfileProxy          *string `env:"DEVFILE_PROXY"`
	DockerCmd             *string `env:"DOCKER_CMD"`
	Globalodoconfig       *string `env:"GLOBALODOCONFIG"`
	OdoDebugTelemetryFile *string `env:"ODO_DEBUG_TELEMETRY_FILE"`
	OdoDisableTelemetry   *bool   `env:"ODO_DISABLE_TELEMETRY"`
	OdoLogLevel           *int    `env:"ODO_LOG_LEVEL"`
	OdoTrackingConsent    *string `env:"ODO_TRACKING_CONSENT"`
	PodmanCmd             *string `env:"PODMAN_CMD"`
	TelemetryCaller       *string `env:"TELEMETRY_CALLER"`
	OdoExperimentalMode   *bool   `env:"ODO_EXPERIMENTAL_MODE"`
}

func GetConfiguration() (*Configuration, error) {
	var s Configuration
	err := envconfig.Process(context.Background(), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
