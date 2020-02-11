package experimental

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/openshift/odo/pkg/preference"
)

func TestIsExperimentalModeEnabled(t *testing.T) {

	const (
		experimentalSetting string = "Experimental"
	)

	// temp config file
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()

	// set global config env
	os.Setenv(preference.GlobalConfigEnvName, tempConfigFile.Name())

	// test table
	tests := []struct {
		name      string
		setEnv    bool
		setConfig bool
		want      bool
	}{
		{
			name:      "enable experimental in config",
			setEnv:    false,
			setConfig: true,
			want:      true,
		},
		{
			name:      "disable experimental",
			setEnv:    false,
			setConfig: false,
			want:      false,
		},
		{
			name:      "enable experimental in env",
			setEnv:    true,
			setConfig: false,
			want:      true,
		},
	}

	// execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// set env if setEnv is true
			if tt.setEnv {
				err := os.Setenv(OdoExperimentalEnv, "true")
				if err != nil {
					t.Errorf("failed to set env %s. err: '%v'", OdoExperimentalEnv, err)
				}
				defer os.Unsetenv(OdoExperimentalEnv)
			}

			// create new preference file
			cfg, err := preference.NewPreferenceInfo()
			if err != nil {
				t.Error(err)
			}

			// set config if setConfig is true
			if tt.setConfig {
				// set experimental preference to true
				err = cfg.SetConfiguration(experimentalSetting, "true")
				if err != nil {
					t.Errorf("failed to set config. err: '%v'", err)
				}
			} else {
				err = cfg.SetConfiguration(experimentalSetting, "false")
				if err != nil {
					t.Errorf("failed to set config. err: '%v'", err)
				}
			}

			// get value
			got := IsExperimentalModeEnabled()

			if got != tt.want {
				t.Errorf("got:%t, want:%t", got, tt.want)
			}
		})
	}
}
