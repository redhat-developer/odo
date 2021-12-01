package testingutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/preference"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

// ConfigDetails struct holds configuration details(odo and/or kube config)
type ConfigDetails struct {
	FileName      string
	Config        interface{}
	ConfigPathEnv string
}

// getConfFolder generates a mock config folder for the unit testing
func getConfFolder() (string, error) {
	var confLocation string
	// If custom home dir is set, skip checking for user.Current to place config
	if len(customHomeDir) != 0 {
		confLocation = customHomeDir
	} else {
		currentUser, err := user.Current()
		if err != nil {
			return "", err
		}
		confLocation = currentUser.HomeDir
	}

	dir, err := ioutil.TempDir(confLocation, ".odo")
	if err != nil {
		return "", err
	}
	return dir, nil
}

// setupTempConfigFile takes config file name - confFile and creates it for unit testing
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

// setupEnv takes odoConfigFile name and sets env var ODOCONFIG to odoConfigFile
// The config logic relies on this env var(if present) to read and/or write config
func setupEnv(envName string, odoconfigfile string) error {
	err := os.Setenv(envName, odoconfigfile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to set %s to %s", envName, odoconfigfile))
	}
	return nil
}

// SetUp sets up the odo and kube config files and returns respective conf file pointers and error
func SetUp(odoConfigDetails ConfigDetails, kubeConfigDetails ConfigDetails) (*os.File, *os.File, error) {
	odoConfigFile, err := setUpConfig(odoConfigDetails.FileName, odoConfigDetails.Config, odoConfigDetails.ConfigPathEnv)
	if err != nil {
		return odoConfigFile, nil, err
	}
	kubeConfigFile, err := setUpConfig(kubeConfigDetails.FileName, kubeConfigDetails.Config, kubeConfigDetails.ConfigPathEnv)
	return odoConfigFile, kubeConfigFile, err
}

// setUpConfig sets up mock config
// Parameters:
//	conf: the config object to write to the mock config file
//	testFile: the name of the mock config file
//  configEnvName: Name of env variable that corresponds to config file
// Returns:
//	file handler for the mock config file
//	error if any

func setUpConfig(testFile string, conf interface{}, configEnvName string) (*os.File, error) {
	foundConfigType := false
	var err error
	var data []byte
	if conf, ok := conf.(preference.PreferenceInfo); ok {
		data, err = yaml.Marshal(conf.Preference)
		foundConfigType = true
	}
	if conf, ok := conf.(clientcmdapi.Config); ok {
		data, err = yaml.Marshal(conf)
		foundConfigType = true
	}
	if conf, ok := conf.(string); ok {
		data = []byte(conf)
		foundConfigType = true
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mock config file")
	}
	if !foundConfigType {
		return nil, fmt.Errorf("Config %+v not of recognisable type", conf)
	}
	configFile, err := setupTempConfigFile(testFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mock config file")
	}
	if conf != nil {
		if _, err := configFile.Write(data); err != nil {
			return nil, errors.Wrapf(err, "unable to write config %+v to mock config file %s", conf, configFile.Name())
		}
	}
	return configFile, setupEnv(configEnvName, configFile.Name())
}

// CleanupEnv cleans up the mock config file and anything that SetupEnv generated
// Parameters:
//	configFile: the mock config file handler
//	t: testing pointer to log errors if any
func CleanupEnv(confFiles []*os.File, t *testing.T) {
	for _, confFile := range confFiles {
		if confFile == nil {
			continue
		}
		if err := confFile.Close(); err != nil {
			t.Errorf("failed to cleanup the test env. Error: %v", err)
		}
		os.Remove(confFile.Name())
		os.Remove(filepath.Dir(confFile.Name()))
	}
}

// FakeOdoConfig returns mock odo config
// It takes a confPath which is the path to the config
func FakeOdoConfig(confPath string, needNamePrefix bool, namePrefix string) preference.PreferenceInfo {
	odoConfig := preference.PreferenceInfo{
		Filename:   confPath,
		Preference: preference.Preference{},
	}
	if needNamePrefix {
		odoConfig.OdoSettings = preference.OdoSettings{
			NamePrefix: &namePrefix,
		}
	}
	return odoConfig
}

// FakeKubeClientConfig returns mock kube client config
func FakeKubeClientConfig() string {
	return `apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://192.168.42.237:8443
  name: 192-168-42-237:8443
contexts:
- context:
    cluster: 192-168-42-237:8443
    namespace: testing
    user: developer/192-168-42-237:8443
  name: myproject/192-168-42-237:8443/developer
current-context: myproject/192-168-42-237:8443/developer
kind: Config
preferences: {}
users:
- name: developer/192-168-42-237:8443
  user:
    token: C0E6Gkmi3n_Se2QKx6Unw3Y3Zu4mJHgzdrMVK0DsDwc`
}
