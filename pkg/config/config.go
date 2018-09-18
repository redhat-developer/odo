package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	configEnvName  = "ODOCONFIG"
	configFileName = "odo"
	//DefaultTimeout for openshift server connection check
	DefaultTimeout = 1
)

// OdoSettings holds all odo specific configurations
type OdoSettings struct {
	// Controls if an update notification is shown or not
	UpdateNotification *bool `json:"updatenotification,omitempty"`
	// Holds the prefix part of generated random application name
	NamePrefix *string `json:"nameprefix,omitempty"`
	// Timeout for openshift server connection check
	Timeout *int `json:"timeout,omitempty"`
}

// ApplicationInfo holds all important information about one application
type ApplicationInfo struct {
	// name of the application
	Name string `json:"name"`
	// is this application active? Only one application can be active at the time
	Active bool `json:"active"`
	// name of the openshift project this application belongs to
	Project string `json:"project"`
	// last active component for  this application
	ActiveComponent string `json:"activeComponent"`
}

type Config struct {

	// odo specific configuration settings
	OdoSettings OdoSettings `json:"settings"`
	// remember active applications and components per project
	// when project or applications is switched we can go back to last active app/component

	// Currently active application
	// multiple applications can be active but each one has to be in different project
	// there shouldn't be more active applications in one project
	ActiveApplications []ApplicationInfo `json:"activeApplications"`
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

func getOdoConfigFile() (string, error) {
	if env, ok := os.LookupEnv(configEnvName); ok {
		return env, nil
	}

	if file := getDefaultConfigFile(); len(file) > 0 {
		return file, nil
	}

	return "", errors.New("unable to get config file")
}

func New() (*ConfigInfo, error) {
	configFile, err := getOdoConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	// Check whether directory present or not
	_, err = os.Stat(filepath.Dir(configFile))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(configFile), 0755)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create directory")
		}
	}
	// Check whether config file is present or not
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
		return errors.Wrap(err, "unable to unmarshal odo config file")
	}

	return nil
}

func (c *ConfigInfo) writeToFile() error {
	data, err := yaml.Marshal(&c.Config)
	if err != nil {
		return errors.Wrap(err, "unable to marshal odo config data")
	}

	err = ioutil.WriteFile(c.Filename, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file %v", c.Filename)
	}

	return nil
}

// SetConfiguration modifies Odo configurations in the config file
// as of now being used for timeout, updatenotification
func (c *ConfigInfo) SetConfiguration(parameter string, value string) error {
	switch parameter {

	case "timeout":
		typedval, err := strconv.Atoi(value)
		if err != nil {
			return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
		}
		if typedval < 0 {
			return errors.Errorf("cannot set timeout to less than 0")
		}
		c.OdoSettings.Timeout = &typedval

	case "updatenotification":
		val, err := strconv.ParseBool(strings.ToLower(value))
		if err != nil {
			return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
		}
		c.OdoSettings.UpdateNotification = &val

	case "nameprefix":
		c.OdoSettings.NamePrefix = &value
	default:
		return errors.Errorf("unknown parameter :'%s' is not a parameter in odo config", parameter)
	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to set %s", parameter)
	}
	return nil
}

// GetTimeout returns the value of Timeout from config
func (c *ConfigInfo) GetTimeout() int {
	// default timeout value is 1
	if c.OdoSettings.Timeout == nil {
		return DefaultTimeout
	}
	return *c.OdoSettings.Timeout
}

// GetUpdateNotification returns the value of UpdateNotification from config
func (c *ConfigInfo) GetUpdateNotification() bool {
	if c.OdoSettings.UpdateNotification == nil {
		return true
	}
	return *c.OdoSettings.UpdateNotification
}

// GetNamePrefix returns the value of Prefix from config
func (c *ConfigInfo) GetNamePrefix() string {
	if c.OdoSettings.NamePrefix == nil {
		return ""
	}
	return *c.OdoSettings.NamePrefix
}

// SetActiveComponent sets active component for given project and application.
// application must exist
func (c *ConfigInfo) SetActiveComponent(component string, application string, project string) error {
	found := false

	if c.ActiveApplications != nil {
		for i, app := range c.ActiveApplications {
			if app.Project == project && app.Name == application {
				c.ActiveApplications[i].ActiveComponent = component
				found = true
				break
			}
		}
	}

	if !found {
		return errors.Errorf("unable to set %s component as active, application %s in %s project doesn't exists", component, application, project)
	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to set %s as active component", component)
	}
	return nil
}

// Sets the active component as blank of the given project in the configuration file
func (c *ConfigInfo) UnsetActiveComponent(project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	for i, app := range c.ActiveApplications {
		if app.Project == project && c.ActiveApplications[i].ActiveComponent != "" {
			c.ActiveApplications[i].ActiveComponent = ""
		}
	}

	// Write the configuration to file
	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to write configuration file")
	}
	return nil

}

// Sets the active application as blank of the given project in the configuration file
func (c *ConfigInfo) UnsetActiveApplication(project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	for i, cfgApp := range c.ActiveApplications {
		if cfgApp.Project == project && c.ActiveApplications[i].Active {
			c.ActiveApplications[i].Active = false
		}
	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrap(err, "unable to write configuration file")
	}
	return nil
}

// GetActiveComponent if no component is set as current returns empty string
func (c *ConfigInfo) GetActiveComponent(application string, project string) string {
	if c.ActiveApplications != nil {
		for _, app := range c.ActiveApplications {
			if app.Project == project && app.Name == application && app.Active == true {
				return app.ActiveComponent
			}
		}
	}
	return ""
}

// GetActiveApplication get currently active application for given project
// if no application is active return empty string
func (c *ConfigInfo) GetActiveApplication(project string) string {
	if c.ActiveApplications != nil {
		for _, app := range c.ActiveApplications {
			if app.Project == project && app.Active == true {
				return app.Name
			}
		}
	}
	return ""
}

// SetActiveApplication set application as active for given project
func (c *ConfigInfo) SetActiveApplication(application string, project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	found := false
	for i, app := range c.ActiveApplications {
		// if application exists set is as Active
		if app.Name == application && app.Project == project {
			c.ActiveApplications[i].Active = true
			found = true
			break
		}
	}

	// if application doesn't exists, add it as Active
	if !found {
		return fmt.Errorf("unable set application %s as active in config, it doesn't exist", application)
	}
	// make sure that no other application is active
	for i, app := range c.ActiveApplications {
		if !(app.Name == application && app.Project == project) {
			c.ActiveApplications[i].Active = false
		}
	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}
	return nil
}

// AddApplication add  new application to the config file
// Newly create application is NOT going to be se as Active.
func (c *ConfigInfo) AddApplication(application string, project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	for _, app := range c.ActiveApplications {
		if app.Name == application && app.Project == project {
			return fmt.Errorf("unable to add %s application, it already exists in config file", application)
		}
	}

	// if application doesn't exists add it to slice
	c.ActiveApplications = append(c.ActiveApplications,
		ApplicationInfo{
			Name:    application,
			Project: project,
			Active:  false,
		})

	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to set add %s application", application)
	}
	return nil
}

// DeleteApplication deletes application from given project from config file
func (c *ConfigInfo) DeleteApplication(application string, project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	found := false
	for i, app := range c.ActiveApplications {
		// if application exists set is as Active
		if app.Name == application && app.Project == project {
			// remove current item from array
			c.ActiveApplications = append(c.ActiveApplications[:i], c.ActiveApplications[i+1:]...)
			found = true
		}
	}

	if !found {
		return fmt.Errorf("application %s doesn't exist", application)

	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", application)
	}
	return nil
}
