package application

import (
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/config"
	"github.com/redhat-developer/odo/pkg/testingutil"
)

func setUp(config *config.ConfigInfo, testFile string) (*os.File, error) {
	configFile, err := setupTempConfigFile(testFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mock config file")
	}
	if config != nil {
		data, err := yaml.Marshal(&config.Config)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to marshal config %+v", config.Config)
		}
		if _, err := configFile.Write(data); err != nil {
			return nil, errors.Wrapf(err, "unable to write config %s to mock config file %s", config, configFile)
		}
	}
	setupEnv(configFile.Name())
	return configFile, nil
}

// The invocation of setupTempConfigFile puts the onus of invoking the configCleanUp as well
func setupTempConfigFile(confFile string) (*os.File, error) {
	confFolder, err := getConfFolder()
	if err != nil {
		return nil, err
	}
	tmpfile, err := ioutil.TempFile(confFolder, confFile)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create test config file")
	}
	return tmpfile, nil
}

func setupEnv(odoconfigfile string) error {
	err := os.Setenv("ODOCONFIG", odoconfigfile)
	if err != nil {
		return errors.Wrap(err, "unable to set ODOCONFIG to odo-test-config")
	}
	return nil
}

func writeTempConfig(config []byte, f *os.File) error {
	if _, err := f.Write(config); err != nil {
		return err
	}
	return nil
}

func getConfFolder() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	dir, err := ioutil.TempDir(currentUser.HomeDir, ".kube")
	if err != nil {
		return "", err
	}
	return dir, nil
}

func cleanupEnv(confFile *os.File) error {
	defer os.Remove(confFile.Name())
	if err := confFile.Close(); err != nil {
		return err
	}
	return nil
}

func TestGetDefaultAppName(t *testing.T) {
	tests := []struct {
		testName         string
		existingAppNames []string
		wantErr          bool
		wantRE           string
		needPrefix       bool
	}{
		{
			testName:         "Case: App prefix not configured",
			existingAppNames: []string{},
			wantErr:          false,
			wantRE:           "app-*",
			needPrefix:       false,
		},
		{
			testName:         "Case: App prefix configured",
			existingAppNames: []string{},
			wantErr:          false,
			wantRE:           "testing-*",
			needPrefix:       true,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			var configInfo config.ConfigInfo
			odoconfigfile := "odo-test-config"

			configInfo = testingutil.FakeOdoConfig(tt.needPrefix, "")

			configFile, err := setUp(&configInfo, odoconfigfile)
			if err != nil {
				t.Errorf("Failed to do required environment setup. Error %v", err)
			}

			defer cleanupEnv(configFile)

			name, err := GetDefaultAppName(tt.existingAppNames)
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
