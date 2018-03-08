package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const (
	configEnvName  = "OCDEVCONFIG"
	configFileName = "ocdev"
)

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

func (c *ConfigInfo) writeToFile() error {
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
		return fmt.Errorf("unable to set %s component as active, application %s in %s project doesn't exists", component, application, project)
	}

	err := c.writeToFile()
	if err != nil {
		return errors.Wrapf(err, "unable to set %s as active component", component)
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
		}
	}

	// if application doesn't exists, add it as Active
	if !found {
		c.ActiveApplications = append(c.ActiveApplications,
			ApplicationInfo{
				Name:    application,
				Project: project,
				Active:  true,
			})
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
