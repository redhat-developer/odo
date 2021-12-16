package segment

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

	err = os.Setenv(preference.GlobalConfigEnvName, tempConfigFile.Name())

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		testName string
		consent  string
		cfg      preference.Client
	}{
		{
			testName: "Registry options with telemetry consent",
			consent:  "true",
		},

		{
			testName: "Registry options without telemetry consent",
			consent:  "false",
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

			ro := GetRegistryOptions()
			err = verifyRegistryOptions(cfg.GetConsentTelemetry(), ro)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func verifyRegistryOptions(isSet bool, ro library.RegistryOptions) error {
	if ro.SkipTLSVerify {
		return errors.New("SkipTLSVerify should be set to false by default")
	}

	return verifyTelemetryData(isSet, ro.Telemetry)
}

func verifyTelemetryData(isSet bool, data library.TelemetryData) error {
	if !isSet {
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
