package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	configEnvName  = "OCDEVCONFIG"
	configFileName = "ocdev"
)

type Application struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

type Config struct {
	Applications       []Application `json:"applications"`
	CurrentApplication string        `json:"currentApplication"`
}

type ConfigInfo struct {
	Filename string
	Config
}

func getDefaultConfigFile() string {
	currentUser, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(currentUser.HomeDir, ".kube", configFileName)
}

func getOcdevConfigFile() (string, error) {
	if env, ok := os.LookupEnv(configEnvName); ok {
		return env, nil
	}

	if file := getDefaultConfigFile(); len(file) > 0 {
		return file, nil
	}

	return "", errors.New("unable to get config file")
}

func New() (*ConfigInfo, error) {
	configFile, err := getOcdevConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ocdev config file")
	}

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create config file")
		}
		defer file.Close()
	}

	c := ConfigInfo{}
	c.Filename = configFile
	c.get()
	return &c, nil
}

func (c *ConfigInfo) get() error {
	configData, err := ioutil.ReadFile(c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %v", c.Filename)
	}

	err = yaml.Unmarshal(configData, &c)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal ocdev config file")
	}

	return nil
}

func (c *ConfigInfo) set() error {
	data, err := yaml.Marshal(&c.Config)
	if err != nil {
		return errors.Wrap(err, "unable to marshal ocdev config data")
	}

	err = ioutil.WriteFile(c.Filename, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file %v", c.Filename)
	}

	return nil
}

func (c *ConfigInfo) ApplicationExists(inputApp *Application) bool {
	for _, app := range c.Applications {
		if reflect.DeepEqual(inputApp, &app) {
			return true
		}
	}
	return false
}

func (c *ConfigInfo) AddApplication(app *Application) error {
	c.Applications = append(c.Applications, *app)
	if err := c.set(); err != nil {
		return errors.Wrap(err, "unable to set config data")
	}
	return nil
}
