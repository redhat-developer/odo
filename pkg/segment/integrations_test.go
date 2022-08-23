package segment

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/devfile/registry-support/registry-library/library"
	"github.com/redhat-developer/odo/pkg/preference"
)

func TestGetRegistryOptions(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()

	t.Setenv(preference.GlobalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		testName      string
		consent       string
		telemetryFile bool
		cfg           preference.Client
	}{
		{
			testName:      "Registry options with telemetry consent and telemetry file",
			consent:       "true",
			telemetryFile: true,
		},
		{
			testName:      "Registry options with telemetry consent and no telemetry file",
			consent:       "true",
			telemetryFile: false,
		},

		{
			testName:      "Registry options without telemetry consent and telemetry file",
			consent:       "false",
			telemetryFile: true,
		},
		{
			testName:      "Registry options without telemetry consent and no telemetry file",
			consent:       "false",
			telemetryFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			cfg, err := preference.NewClient()
			if err != nil {
				t.Error(err)
			}
			err = cfg.SetConfiguration(preference.ConsentTelemetrySetting, tt.consent)
			if err != nil {
				t.Error(err)
			}

			if tt.telemetryFile {
				t.Setenv(DebugTelemetryFileEnv, "/a/telemetry/file")
			}

			ro := GetRegistryOptions()
			err = verifyRegistryOptions(cfg.GetConsentTelemetry(), tt.telemetryFile, ro)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func verifyRegistryOptions(isSet bool, telemetryFile bool, ro library.RegistryOptions) error {
	if ro.SkipTLSVerify {
		return errors.New("SkipTLSVerify should be set to false by default")
	}

	return verifyTelemetryData(isSet, telemetryFile, ro.Telemetry)
}

func verifyTelemetryData(isSet bool, telemetryFile bool, data library.TelemetryData) error {
	if !isSet || telemetryFile {
		if data.Locale == "" && data.User == "" {
			return nil
		}

		return fmt.Errorf("Locale %s and User %s should be unset when telemetry is not enabled ", data.Locale, data.User)

	} else {
		//we don't care what value locale and user have been set to.  We just want to make sure they are not empty
		if data.Locale != "" && data.User != "" {
			return nil
		}

		return fmt.Errorf("Locale %s and User %s should be set when telemetry is enabled ", data.Locale, data.User)
	}
}
