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
		wantErr      bool
		wantRE       string
		needPrefix   bool
	}{
		{
			testName:     "Case: App prefix not configured",
			existingApps: []config.ApplicationInfo{},
			wantErr:      false,
			wantRE:       "app-*",
			needPrefix:   false,
		},
		{
			testName:     "Case: App prefix configured",
			existingApps: []config.ApplicationInfo{},
			wantErr:      false,
			wantRE:       "testing-*",
			needPrefix:   true,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			var configInfo config.ConfigInfo
			odoconfigfile := "odo-test-config"

			configInfo = testingutil.FakeOdoConfig(tt.needPrefix, "")

			configFile, err := testingutil.SetUp(&configInfo, odoconfigfile)
			if err != nil {
				t.Errorf("Failed to do required environment setup. Error %v", err)
			}

			defer testingutil.CleanupEnv(configFile, t)

			name, err := GetDefaultAppName(tt.existingApps)
			if err != nil {
				t.Errorf("Failed to setup mock environment. Error: %v", err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Expected err: %v, but err is %v", tt.wantErr, err)
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
