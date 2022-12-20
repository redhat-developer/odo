package feature

import (
	"context"
	"testing"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
)

func TestIsEnabled(t *testing.T) {
	type args struct {
		feature OdoFeature
	}
	type env struct {
		experimentalMode bool
	}
	type testCase struct {
		name string
		env  env
		args args
		want bool
	}

	nonExperimentalFeature := OdoFeature{}
	experimentalFeature := OdoFeature{isExperimental: true}

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
				experimentalMode: true,
			},
			want: true,
		},
		{
			name: "non-experimental feature should always be enabled even if experimental mode is not enabled",
			args: args{feature: nonExperimentalFeature},
			env: env{
				experimentalMode: false,
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
				experimentalMode: false,
			},
			want: false,
		},
		{
			name: "experimental feature should be enabled only if experimental mode is enabled",
			args: args{feature: experimentalFeature},
			env: env{
				experimentalMode: true,
			},
			want: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.Configuration{}
			cfg.OdoExperimentalMode = tt.env.experimentalMode
			ctx = envcontext.WithEnvConfig(ctx, cfg)

			got := IsEnabled(ctx, tt.args.feature)

			if got != tt.want {
				t.Errorf("IsEnabled: expected %v, but got %v. Env: %v", tt.want, got, tt.env)
			}
		})
	}
}
