package application

import (
	"os"
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
			testName:     "Case: App prefix set to AUTOMATIC",
			existingApps: []config.ApplicationInfo{},
			wantRE:       "application-*",
			needPrefix:   true,
			prefix:       "AUTOMATIC",
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {

			odoConfigFile, kubeConfigFile, err := testingutil.SetUp(
				testingutil.ConfigDetails{
					FileName:      "odo-test-config",
					Config:        testingutil.FakeOdoConfig("odo-test-config", tt.needPrefix, tt.prefix),
					ConfigPathEnv: "ODOCONFIG",
				}, testingutil.ConfigDetails{
					FileName:      "kube-test-config",
					Config:        testingutil.FakeKubeClientConfig(),
					ConfigPathEnv: "KUBECONFIG",
				},
			)
			defer testingutil.CleanupEnv([]*os.File{odoConfigFile, kubeConfigFile}, t)
			if err != nil {
				t.Errorf("failed setting up the test env. Error: %v", err)
			}

			name, err := GetDefaultAppName(tt.existingApps)
			if err != nil {
				t.Errorf("failed to setup mock environment. Error: %v", err)
			}

			r, _ := regexp.Compile(tt.wantRE)
			match := r.MatchString(name)
			if !match {
				fetchedConfig, _ := config.New()
				t.Errorf("randomly generated application name %s does not match regexp %s and config is %+v\nthe prefix is %s", name, tt.wantRE, fetchedConfig, *fetchedConfig.OdoSettings.NamePrefix)
			}
		})
	}
}
