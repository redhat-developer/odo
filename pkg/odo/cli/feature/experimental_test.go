package feature

import (
	"fmt"
	"testing"
)

func TestIsExperimental(t *testing.T) {
	type args struct {
		feature OdoFeature
	}
	type env struct {
		hasExperimentalModeEnvVar bool
		experimentalMode          string
	}
	type testCase struct {
		name     string
		env      env
		args     args
		wantFunc func(experimentalModeEnv string, feat OdoFeature) bool
	}

	var testCases []testCase
	featListToTest := []OdoFeature{{}, {id: "404-should-not-exist", description: "unknown feature"}}
	featListToTest = append(featListToTest, _experimentalFeatures...)
	for _, feat := range featListToTest {
	envVarLoop:
		for _, envVarSet := range []bool{false, true} {
			for _, envVar := range []string{"", "true", "false", "yes", "no", "foobar"} {
				envVarSet := envVarSet
				envVar := envVar
				feat := feat
				name := fmt.Sprintf("%s not set", OdoExperimentalModeEnvVar)
				if envVarSet {
					name = fmt.Sprintf("%s=%s", OdoExperimentalModeEnvVar, envVar)
				}
				testCases = append(testCases, testCase{
					name: fmt.Sprintf("%s, feature=%s, env=%s", name, feat, envVar),
					env: env{
						hasExperimentalModeEnvVar: envVarSet,
						experimentalMode:          envVar,
					},
					args: args{feature: feat},
					wantFunc: func(experimentalModeEnv string, feat OdoFeature) bool {
						if experimentalModeEnv != "true" {
							return false
						}
						for _, f := range _experimentalFeatures {
							if f == feat {
								return true
							}
						}
						return false
					},
				})
				if !envVarSet {
					// we don't care about other values of envVar since we won't be setting them as env vars
					continue envVarLoop
				}
			}
		}
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env.hasExperimentalModeEnvVar {
				t.Setenv(OdoExperimentalModeEnvVar, tt.env.experimentalMode)
			}

			got := IsExperimental(tt.args.feature)

			var want bool
			if tt.env.hasExperimentalModeEnvVar {
				want = tt.wantFunc(tt.env.experimentalMode, tt.args.feature)
			}
			if got != want {
				t.Errorf("IsExperimental: expected %v, but got %v. Env: %v", want, got, tt.env)
			}
		})
	}
}
