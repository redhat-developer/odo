package segment

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/preference"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
)

func TestGetRegistryOptions(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	err = tempConfigFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempConfigFile.Name())

	t.Setenv(preference.GlobalConfigEnvName, tempConfigFile.Name())

	type want struct {
		localeUserEmpty bool
		skipTLSVerify   bool
		caller          string
	}
	tests := []struct {
		testName      string
		consent       bool
		telemetryFile bool
		caller        string
		cfg           preference.Client
		want          want
	}{
		{
			testName:      "Registry options with telemetry consent and telemetry file",
			consent:       true,
			telemetryFile: true,
			want: want{
				localeUserEmpty: true,
				skipTLSVerify:   false,
				caller:          "odo",
			},
		},
		{
			testName:      "Registry options with telemetry consent and no telemetry file",
			consent:       true,
			telemetryFile: false,
			want: want{
				localeUserEmpty: false,
				skipTLSVerify:   false,
				caller:          "odo",
			},
		},

		{
			testName:      "Registry options without telemetry consent and telemetry file",
			consent:       false,
			telemetryFile: true,
			want: want{
				localeUserEmpty: true,
				skipTLSVerify:   false,
				caller:          "odo",
			},
		},
		{
			testName:      "Registry options without telemetry consent and no telemetry file",
			consent:       false,
			telemetryFile: false,
			want: want{
				localeUserEmpty: true,
				skipTLSVerify:   false,
				caller:          "odo",
			},
		},
		{
			testName:      "Registry options without telemetry consent and no telemetry file, with caller",
			consent:       false,
			telemetryFile: false,
			caller:        "vscode",
			want: want{
				localeUserEmpty: true,
				skipTLSVerify:   false,
				caller:          "odo-vscode",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := scontext.NewContext(context.Background())
			var envConfig config.Configuration
			if tt.telemetryFile {
				envConfig.OdoDebugTelemetryFile = pointer.String("/a/telemetry/file")
			}
			if tt.caller != "" {
				envConfig.TelemetryCaller = tt.caller
			}
			ctx = envcontext.WithEnvConfig(ctx, envConfig)
			scontext.SetTelemetryStatus(ctx, tt.consent)

			ro := GetRegistryOptions(ctx)

			if len(ro.Telemetry.Locale) == 0 != tt.want.localeUserEmpty || len(ro.Telemetry.User) == 0 != tt.want.localeUserEmpty {
				t.Errorf("Locale %q and User %q emptiness should be %v when telemetry enabled is %v and telemetry file is %v", ro.Telemetry.Locale, ro.Telemetry.User, tt.want.localeUserEmpty, tt.consent, tt.telemetryFile)
			}

			if ro.SkipTLSVerify != tt.want.skipTLSVerify {
				t.Errorf("SkipTLSVerify should be set to %v by default", tt.want.skipTLSVerify)
			}

			if ro.Telemetry.Client != tt.want.caller {
				t.Errorf("caller should be %q but is %q", tt.want.caller, ro.Telemetry.Client)
			}
		})
	}
}
