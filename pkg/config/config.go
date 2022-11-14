package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	DevfileProxy          *string `split_words:"true"`
	DockerCmd             *string `split_words:"true"`
	Globalodoconfig       *string `split_words:"true"`
	OdoDebugTelemetryFile *string `split_words:"true"`
	OdoDisableTelemetry   *bool   `split_words:"true"`
	OdoLogLevel           *int    `split_words:"true"`
	OdoTrackingConsent    *string `split_words:"true"`
	PodmanCmd             *string `split_words:"true"`
	TelemetryCaller       *string `split_words:"true"`
	OdoExperimentalMode   *bool   `split_words:"true"`
}

func GetConfiguration() (*Configuration, error) {
	var s Configuration
	err := envconfig.Process("", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
