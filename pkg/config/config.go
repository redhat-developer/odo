package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	globalConfigEnvName = "GLOBALODOCONFIG"
	localConfigEnvName  = "LOCALODOCONFIG"
	configFileName      = "odo-config"
	//DefaultTimeout for openshift server connection check
	DefaultTimeout = 1
)

type ConfigType int

const (
	Local = iota
	Global
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

// ComponentSettings holds all component related information
type ComponentSettings struct {

	// The builder image to use
	ComponentType *string `json:"component_type,omitempty"`
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

type GlobalConfig struct {
	// global config has all the attributes of global config as well
	*LocalConfig
	// remember active applications and components per project
	// when project or applications is switched we can go back to last active app/component

	// Currently active application
	// multiple applications can be active but each one has to be in different project
	// there shouldn't be more active applications in one project
	ActiveApplications []ApplicationInfo `json:"activeApplications"`

	// Odo settings holds the odo specific global settings
	OdoSettings OdoSettings `json:"settings"`
}

type LocalConfig struct {
	ComponentSettings ComponentSettings `json:"component_settings"`
}

type GlobalConfigInfo struct {
	Filename string
	GlobalConfig
}

type LocalConfigInfo struct {
	Filename string
	LocalConfig
}

func getGlobalConfigFile() (string, error) {
	if env, ok := os.LookupEnv(globalConfigEnvName); ok {
		return env, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentUser.HomeDir, ".odo", configFileName), nil
}

func getLocalConfigFile() (string, error) {
	if env, ok := os.LookupEnv(localConfigEnvName); ok {
		return env, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(wd, ".odo", configFileName), nil
}

// NewGlobalConfig gets the GlobalConfigInfo from global config file and global creates the config file in case it's
// not present then it
func NewGlobalConfig() (*GlobalConfigInfo, error) {
	configFile, err := getGlobalConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	// Check whether directory and file are not present if they aren't then create them
	if err = createIfNotExists(configFile); err != nil {
		return nil, err
	}
	c := GlobalConfigInfo{}
	c.Filename = configFile
	get(c.GlobalConfig, c.Filename)
	return &c, nil
}

func NewLocalConfig() (*LocalConfigInfo, error) {
	configFile, err := getLocalConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	// Check whether directory and file are not present if they aren't then create them
	if err = createIfNotExists(configFile); err != nil {
		return nil, err
	}
	c := LocalConfigInfo{}
	c.Filename = configFile
	get(c.LocalConfig, c.Filename)
	return &c, nil
}

func createIfNotExists(configFile string) error {
	_, err := os.Stat(filepath.Dir(configFile))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(configFile), 0755)
		if err != nil {
			return errors.Wrap(err, "unable to create directory")
		}
	}
	// Check whether config file is present or not
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			return errors.Wrap(err, "unable to create config file")
		}
		defer file.Close()
	}

	return nil
}

func get(c interface{}, filename string) error {
	configData, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to read file %v", filename)
	}

	err = yaml.Unmarshal(configData, &c)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal odo config file")
	}

	return nil
}

func writeToFile(c interface{}, filename string) error {
	data, err := yaml.Marshal(&c)
	if err != nil {
		return errors.Wrap(err, "unable to marshal odo config data")
	}

	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file %v", c)
	}

	return nil
}

// SetConfiguration modifies Odo configurations in the config file
// as of now being used for nameprefix, timeout, updatenotification
func (c *GlobalConfigInfo) SetConfiguration(parameter string, value string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names
		switch p {

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
		}
	} else {
		return errors.Errorf("unknown parameter :'%s' is not a parameter in odo config", parameter)
	}

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to set %s", parameter)
	}
	return nil
}

// GetTimeout returns the value of Timeout from config
func (c *GlobalConfigInfo) GetTimeout() int {
	// default timeout value is 1
	if c.OdoSettings.Timeout == nil {
		return DefaultTimeout
	}
	return *c.OdoSettings.Timeout
}

// GetUpdateNotification returns the value of UpdateNotification from config
func (c *GlobalConfigInfo) GetUpdateNotification() bool {
	if c.OdoSettings.UpdateNotification == nil {
		return true
	}
	return *c.OdoSettings.UpdateNotification
}

// GetNamePrefix returns the value of Prefix from config
func (c *GlobalConfigInfo) GetNamePrefix() string {
	if c.OdoSettings.NamePrefix == nil {
		return ""
	}
	return *c.OdoSettings.NamePrefix
}

// SetActiveComponent sets active component for given project and application.
// application must exist
func (c *GlobalConfigInfo) SetActiveComponent(componentName string, applicationName string, projectName string) error {
	found := false

	if c.ActiveApplications != nil {
		for i, app := range c.ActiveApplications {
			if app.Project == projectName && app.Name == applicationName {
				c.ActiveApplications[i].ActiveComponent = componentName
				found = true
				break
			}
		}
	}

	if !found {
		return errors.Errorf("unable to set %s componentName as active, applicationName %s in %s projectName doesn't exists", componentName, applicationName, projectName)
	}

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to set %s as active componentName", componentName)
	}
	return nil
}

// UnsetActiveComponent sets the active component as blank of the given project in the configuration file
func (c *GlobalConfigInfo) UnsetActiveComponent(project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	for i, app := range c.ActiveApplications {
		if app.Project == project && c.ActiveApplications[i].ActiveComponent != "" {
			c.ActiveApplications[i].ActiveComponent = ""
		}
	}

	// Write the configuration to file
	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to write configuration file")
	}
	return nil

}

// UnsetActiveApplication sets the active application as blank of the given project in the configuration file
func (c *GlobalConfigInfo) UnsetActiveApplication(project string) error {
	if c.ActiveApplications == nil {
		c.ActiveApplications = []ApplicationInfo{}
	}

	for i, cfgApp := range c.ActiveApplications {
		if cfgApp.Project == project && c.ActiveApplications[i].Active {
			c.ActiveApplications[i].Active = false
		}
	}

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrap(err, "unable to write configuration file")
	}
	return nil
}

// GetActiveComponent if no component is set as current returns empty string
func (c *GlobalConfigInfo) GetActiveComponent(application string, project string) string {
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
func (c *GlobalConfigInfo) GetActiveApplication(project string) string {
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
func (c *GlobalConfigInfo) SetActiveApplication(application string, project string) error {
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

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrap(err, "unable to set current application")
	}
	return nil
}

// AddApplication add  new application to the config file
// Newly create application is NOT going to be se as Active.
func (c *GlobalConfigInfo) AddApplication(application string, project string) error {
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

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to set add %s application", application)
	}
	return nil
}

// DeleteApplication deletes application from given project from config file
func (c *GlobalConfigInfo) DeleteApplication(application string, project string) error {
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

	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to delete application %s", application)
	}
	return nil
}

// DeleteProject deletes applications belonging to the given project from the config file
func (c *GlobalConfigInfo) DeleteProject(projectName string) error {
	// looping in reverse and removing to avoid panic from index out of bounds
	for i := len(c.ActiveApplications) - 1; i >= 0; i-- {
		if c.ActiveApplications[i].Project == projectName {
			// remove current item from array
			c.ActiveApplications = append(c.ActiveApplications[:i], c.ActiveApplications[i+1:]...)
		}
	}
	err := writeToFile(c.GlobalConfig, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to delete project from config")
	}
	return nil
}

const (
	// UpdateNotificationSetting is the name of the setting controlling update notification
	UpdateNotificationSetting = "UpdateNotification"
	// UpdateNotificationSettingDescription is human-readable description for the update notification setting
	UpdateNotificationSettingDescription = "Controls if an update notification is shown or not (true or false)"
	// NamePrefixSetting is the name of the setting controlling name prefix
	NamePrefixSetting = "NamePrefix"
	// NamePrefixSettingDescription is human-readable description for the name prefix setting
	NamePrefixSettingDescription = "Default prefix is the current directory name. Use this value to set a default name prefix"
	// TimeoutSetting Name of the setting controlling timeout for connection check
	TimeoutSetting = "Timeout"
	// TimeoutSettingDescription is human-readable description for the timeout setting
	TimeoutSettingDescription = "Timeout (in seconds) for OpenShift server connection check"
)

var (
	// records information on supported parameters
	supportedParameterDescriptions = map[string]string{
		UpdateNotificationSetting: UpdateNotificationSettingDescription,
		NamePrefixSetting:         NamePrefixSettingDescription,
		TimeoutSetting:            TimeoutSettingDescription,
	}
	// set-like map to quickly check if a parameter is supported
	lowerCaseParameters = getLowerCaseParameters()
)

// FormatSupportedParameters outputs supported parameters and their description
func FormatSupportedParameters() (result string) {
	for _, v := range GetSupportedParameters() {
		result = result + v + " - " + supportedParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Parameters:\n" + result
}

// asSupportedParameter checks that the given parameter is supported and returns a lower case version of it if it is
func asSupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseParameters[lower]
}

// GetSupportedParameters returns the name of the supported parameters
func GetSupportedParameters() []string {
	return util.GetSortedKeys(supportedParameterDescriptions)
}

// getLowerCaseParameters creates a set-like map of supported parameters from the supported parameter names
func getLowerCaseParameters() map[string]bool {
	parameters := GetSupportedParameters()
	result := make(map[string]bool, len(parameters))
	for _, v := range parameters {
		result[strings.ToLower(v)] = true
	}
	return result
}
