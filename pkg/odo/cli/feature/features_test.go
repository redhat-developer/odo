package feature

import "testing"

func TestIsEnabled(t *testing.T) {
	type args struct {
		feature OdoFeature
	}
	type env struct {
		hasExperimentalModeEnvVar bool
		experimentalMode          string
	}
	type testCase struct {
		name string
		env  env
		args args
		want bool
	}

	nonExperimentalFeature := OdoFeature{
		id:          "my-awesome-feature",
		description: "command: my awesome feature",
	}
	experimentalFeature := OdoFeature{
		id:             "my-wip-flag",
		isExperimental: true,
		description:    "flag: --my-awesome-flag",
	}

	for _, tt := range []testCase{
		{
			name: "non-experimental feature should always be enabled",
			args: args{feature: nonExperimentalFeature},
			want: true,
		},
		{
			name: "non-experimental feature should always be enabled regardless of experimental mode",
			args: args{feature: nonExperimentalFeature},
			env: env{
				hasExperimentalModeEnvVar: true,
				experimentalMode:          OdoExperimentalModeTrue,
			},
			want: true,
		},
		{
			name: "non-experimental feature should always be enabled even if experimental mode is not enabled",
			args: args{feature: nonExperimentalFeature},
			env: env{
				hasExperimentalModeEnvVar: true,
				experimentalMode:          "false",
			},
			want: true,
		},
		{
			name: "experimental feature should be disabled if experimental mode env var is not set",
			args: args{feature: experimentalFeature},
			want: false,
		},
		{
			name: "experimental feature should be disabled if experimental mode has an unknown value",
			args: args{feature: experimentalFeature},
			env: env{
				hasExperimentalModeEnvVar: true,
				experimentalMode:          "false",
			},
			want: false,
		},
		{
			name: "experimental feature should be disabled if experimental mode has an unknown value",
			args: args{feature: experimentalFeature},
			env: env{
				hasExperimentalModeEnvVar: true,
				experimentalMode:          "foobar",
			},
			want: false,
		},
		{
			name: "experimental feature should be enabled only if experimental mode is enabled",
			args: args{feature: experimentalFeature},
			env: env{
				hasExperimentalModeEnvVar: true,
				experimentalMode:          OdoExperimentalModeTrue,
			},
			want: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env.hasExperimentalModeEnvVar {
				t.Setenv(OdoExperimentalModeEnvVar, tt.env.experimentalMode)
			}

			got := IsEnabled(tt.args.feature)

			if got != tt.want {
				t.Errorf("IsEnabled: expected %v, but got %v. Env: %v", tt.want, got, tt.env)
			}
		})
	}
}
