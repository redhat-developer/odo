package config

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Configuration struct {
	DevfileProxy                  *string       `env:"DEVFILE_PROXY,noinit"`
	DockerCmd                     string        `env:"DOCKER_CMD,default=docker"`
	Globalodoconfig               *string       `env:"GLOBALODOCONFIG,noinit"`
	OdoDebugTelemetryFile         *string       `env:"ODO_DEBUG_TELEMETRY_FILE,noinit"`
	OdoDisableTelemetry           *bool         `env:"ODO_DISABLE_TELEMETRY,noinit"`
	OdoLogLevel                   *int          `env:"ODO_LOG_LEVEL,noinit"`
	OdoTrackingConsent            *string       `env:"ODO_TRACKING_CONSENT,noinit"`
	PodmanCmd                     string        `env:"PODMAN_CMD,default=podman"`
	PodmanCmdInitTimeout          time.Duration `env:"PODMAN_CMD_INIT_TIMEOUT,default=1s"`
	TelemetryCaller               string        `env:"TELEMETRY_CALLER,default="`
	OdoExperimentalMode           bool          `env:"ODO_EXPERIMENTAL_MODE,default=false"`
	PushImages                    bool          `env:"ODO_PUSH_IMAGES,default=true"`
	OdoContainerBackendGlobalArgs []string      `env:"ODO_CONTAINER_BACKEND_GLOBAL_ARGS,noinit,delimiter=;"`
	OdoImageBuildArgs             []string      `env:"ODO_IMAGE_BUILD_ARGS,noinit,delimiter=;"`
	OdoContainerRunArgs           []string      `env:"ODO_CONTAINER_RUN_ARGS,noinit,delimiter=;"`
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
