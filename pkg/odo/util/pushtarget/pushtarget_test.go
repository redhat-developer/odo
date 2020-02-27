package pushtarget

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/openshift/odo/pkg/preference"
)

func TestIsDockerPushTargetEnabled(t *testing.T) {

	const (
		pushTargetSetting string = "PushTarget"
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
		setEnv    string
		setConfig string
		want      bool
	}{
		{
			name:      "no pushtarget setting set",
			setEnv:    "",
			setConfig: "",
			want:      false,
		},
		{
			name:      "set pushtarget to docker in config",
			setEnv:    "",
			setConfig: "docker",
			want:      true,
		},
		{
			name:      "set pushtarget to kube in config",
			setEnv:    "",
			setConfig: "kube",
			want:      false,
		},
		{
			name:      "enable docker pushtarget in env",
			setEnv:    "docker",
			setConfig: "",
			want:      true,
		},
		{
			name:      "set pushtarget to kube in env",
			setEnv:    "kube",
			setConfig: "",
			want:      false,
		},
		{
			name:      "override pushtarget prefrence with docker pushtarget env",
			setEnv:    "docker",
			setConfig: "kube",
			want:      true,
		},
		{
			name:      "override pushtarget prefrence with kube pushtarget env",
			setEnv:    "kube",
			setConfig: "docker",
			want:      false,
		},
	}

	// execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// set env if set to a non-empty string
			if tt.setEnv != "" {
				err := os.Setenv(OdoPushTarget, tt.setEnv)
				if err != nil {
					t.Errorf("failed to set env %s. err: '%v'", OdoPushTarget, err)
				}
				defer os.Unsetenv(OdoPushTarget)
			}

			// create new preference file
			cfg, err := preference.NewPreferenceInfo()
			if err != nil {
				t.Error(err)
			}

			// set config if setConfig is a non-empty string
			if tt.setConfig != "" {
				// set experimental preference to true
				err = cfg.SetConfiguration(pushTargetSetting, tt.setConfig)
				if err != nil {
					t.Errorf("failed to set config. err: '%v'", err)
				}
			}

			// get value
			got := IsPushTargetDocker()

			if got != tt.want {
				t.Errorf("got:%t, want:%t", got, tt.want)
			}
		})
	}
}
