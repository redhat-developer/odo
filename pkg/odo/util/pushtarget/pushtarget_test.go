package pushtarget

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/openshift/odo/pkg/preference"
)

func TestPushTargetDocker(t *testing.T) {

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
			name:      "case 1: no pushtarget setting set",
			setEnv:    "",
			setConfig: "",
			want:      false,
		},
		{
			name:      "case 2: set pushtarget to docker in config",
			setEnv:    "",
			setConfig: preference.DockerPushTarget,
			want:      true,
		},
		{
			name:      "case 3: set pushtarget to kube in config",
			setEnv:    "",
			setConfig: preference.KubePushTarget,
			want:      false,
		},
		{
			name:      "case 4: enable docker pushtarget in env",
			setEnv:    preference.DockerPushTarget,
			setConfig: "",
			want:      true,
		},
		{
			name:      "case 5: set pushtarget to kube in env",
			setEnv:    preference.KubePushTarget,
			setConfig: "",
			want:      false,
		},
		{
			name:      "case 6: override pushtarget prefrence with docker pushtarget env",
			setEnv:    preference.DockerPushTarget,
			setConfig: preference.KubePushTarget,
			want:      true,
		},
		{
			name:      "case 7: override pushtarget prefrence with kube pushtarget env",
			setEnv:    preference.KubePushTarget,
			setConfig: preference.DockerPushTarget,
			want:      false,
		},
		{
			name:      "case 8: invalid env var",
			setEnv:    "foo",
			setConfig: "",
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
