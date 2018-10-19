package application

import (
	"regexp"
	"testing"

	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/testingutil"
)

func TestGetDefaultAppName(t *testing.T) {
	tests := []struct {
		testName     string
		existingApps []config.ApplicationInfo
		wantRE       string
		needPrefix   bool
		prefix       string
	}{
		{
			testName:     "Case: App prefix not configured",
			existingApps: []config.ApplicationInfo{},
			wantRE:       "app-*",
			needPrefix:   false,
		},
		{
			testName:     "Case: App prefix set to testing",
			existingApps: []config.ApplicationInfo{},
			wantRE:       "testing-*",
			needPrefix:   true,
			prefix:       "testing",
		},
		{
			testName:     "Case: App prefix set to $DIR",
			existingApps: []config.ApplicationInfo{},
			wantRE:       "testing-*",
			needPrefix:   true,
			prefix:       "$DIR",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			var configInfo config.ConfigInfo
			odoconfigfile := "odo-test-config"

			configInfo = testingutil.FakeOdoConfig(tt.needPrefix, tt.prefix)

			configFile, err := testingutil.SetUp(&configInfo, odoconfigfile)
			if err != nil {
				t.Errorf("Failed to do required environment setup. Error %v", err)
			}

			defer testingutil.CleanupEnv(configFile, t)

			name, err := GetDefaultAppName(tt.existingApps)
			if err != nil {
				t.Errorf("Failed to setup mock environment. Error: %v", err)
			}

			r, _ := regexp.Compile(tt.wantRE)
			match := r.MatchString(name)
			if !match {
				fetchedConfig, _ := config.New()
				t.Errorf("Randomly generated application name %s does not match regexp %s and config is %+v\nthe prefix is %s", name, tt.wantRE, fetchedConfig, *fetchedConfig.OdoSettings.Prefix)
			}
		})
	}
}
