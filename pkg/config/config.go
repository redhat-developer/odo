package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"

	"github.com/pkg/errors"
)

const (
	localConfigEnvName = "LOCALODOCONFIG"
	configFileName     = "config.yaml"
)

// Info is implemented by configuration managers
type Info interface {
	SetConfiguration(parameter string, value string) error
	GetConfiguration(parameter string) (interface{}, bool)
	DeleteConfiguration(parameter string) error
}

// ComponentSettings holds all component related information
type ComponentSettings struct {

	// The builder image to use
	ComponentType *string `yaml:"ComponentType,omitempty"`

	ComponentName *string `yaml:"ComponentName,omitempty"`

	MinMemory *string `yaml:"MinMemory,omitempty"`

	MaxMemory *string `yaml:"MaxMemory,omitempty"`

	// Ignore if set to true then odoignore file should be considered
	Ignore *bool `yaml:"Ignore,omitempty"`

	MinCPU *string `yaml:"MinCPU,omitempty"`

	MaxCPU *string `yaml:"MaxCPU,omitempty"`
}

// LocalConfig holds all the config relavent to a specific Component.
type LocalConfig struct {
	ComponentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// LocalConfigInfo wraps the local config and provides helpers to
// serialize it.
type LocalConfigInfo struct {
	Filename    string `yaml:"FileName,omitempty"`
	LocalConfig `yaml:",omitempty"`
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

// New returns the localConfigInfo
func New() (*LocalConfigInfo, error) {
	return NewLocalConfig()
}

// NewLocalConfig gets the LocalConfigInfo from local config file and creates the local config file in case it's
// not present then it
func NewLocalConfig() (*LocalConfigInfo, error) {
	configFile, err := getLocalConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	c := LocalConfigInfo{
		LocalConfig: LocalConfig{},
	}
	c.Filename = configFile

	// if the config file doesn't exist then we dont worry about it and return
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		return &c, nil
	}
	err = util.GetFromFile(&c.LocalConfig, c.Filename)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SetConfiguration sets the common config settings like component type, min memory
// max memory etc.
// TODO: Use reflect to set parameters
func (lci *LocalConfigInfo) SetConfiguration(parameter string, value string) (err error) {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "componenttype":
			lci.ComponentSettings.ComponentType = &value
		case "componentname":
			lci.ComponentSettings.ComponentName = &value
		case "minmemory":
			lci.ComponentSettings.MinMemory = &value
		case "maxmemory":
			lci.ComponentSettings.MaxMemory = &value
		case "memory":
			lci.ComponentSettings.MaxMemory = &value
			lci.ComponentSettings.MinMemory = &value
		case "ignore":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			lci.ComponentSettings.Ignore = &val
		case "mincpu":
			lci.ComponentSettings.MinCPU = &value
		case "maxcpu":
			lci.ComponentSettings.MaxCPU = &value
		case "cpu":
			lci.ComponentSettings.MinCPU = &value
			lci.ComponentSettings.MaxCPU = &value

		}

		return util.WriteToFile(&lci.LocalConfig, lci.Filename)
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

// GetConfiguration uses reflection to get the parameter from the localconfig struct, currently
// it only searches the componentSettings
func (lci *LocalConfigInfo) GetConfiguration(parameter string) (interface{}, bool) {

	switch strings.ToLower(parameter) {
	case "cpu":
		if lci.ComponentSettings.MinCPU == nil {
			return nil, true
		}
		return *lci.ComponentSettings.MinCPU, true
	case "memory":
		if lci.ComponentSettings.MinMemory == nil {
			return nil, true
		}
		return *lci.ComponentSettings.MinMemory, true
	}

	return util.GetConfiguration(lci.ComponentSettings, parameter)
}

// DeleteConfiguration is used to delete config from local odo config
func (lci *LocalConfigInfo) DeleteConfiguration(parameter string) error {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {

		switch parameter {
		case "cpu":
			lci.ComponentSettings.MinCPU = nil
			lci.ComponentSettings.MaxCPU = nil
		case "memory":
			lci.ComponentSettings.MinMemory = nil
			lci.ComponentSettings.MaxMemory = nil
		default:
			if err := util.DeleteConfiguration(&lci.ComponentSettings, parameter); err != nil {
				return err
			}
		}
		return util.WriteToFile(&lci.LocalConfig, lci.Filename)
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

// GetComponentType returns type of component (builder image name) in the config
// and if absent then returns default
func (lc *LocalConfig) GetComponentType() string {
	if lc.ComponentSettings.ComponentType == nil {
		return ""
	}
	return *lc.ComponentSettings.ComponentType
}

const (

	// ComponentType is the name of the setting controlling the component type i.e. builder image
	ComponentType = "ComponentType"
	// ComponentTypeDescription is human-readable description of the componentType setting
	ComponentTypeDescription = "The type of component"
	// ComponentName is the name of the setting controlling the component name
	ComponentName = "ComponentName"
	// ComponentNameDescription is human-readable description of the componentType setting
	ComponentNameDescription = "The name of the component"
	// MinMemory is the name of the setting controlling the min memory a component consumes
	MinMemory = "MinMemory"
	// MinMemoryDescription is the name of the setting controlling the minimum memory
	MinMemoryDescription = "The minimum memory a component is provided"
	// MaxMemory is the name of the setting controlling the min memory a component consumes
	MaxMemory = "MaxMemory"
	// MaxMemoryDescription is the name of the setting controlling the maximum memory
	MaxMemoryDescription = "The maximum memory a component can consume"
	// Memory is the name of the setting controlling the memory a component consumes
	Memory = "Memory"
	// MemoryDescription is the name of the setting controlling the min and max memory to same value
	MemoryDescription = "The minimum and maximum Memory a component can consume"
	// Ignore is the name of the setting controlling the min memory a component consumes
	Ignore = "Ignore"
	// IgnoreDescription is the name of the setting controlling the use of .odoignore file
	IgnoreDescription = "Consider the .odoignore file for push and watch"
	// MinCPU is the name of the setting controlling minimum cpu
	MinCPU = "MinCPU"
	// MinCPUDescription is the name of the setting controlling the min CPU value
	MinCPUDescription = "The minimum cpu a component can consume"
	// MaxCPU is the name of the setting controlling the use of .odoignore file
	MaxCPU = "MaxCPU"
	//MaxCPUDescription is the name of the setting controlling the max CPU value
	MaxCPUDescription = "The maximum cpu a component can consume"
	// CPU is the name of the setting controlling the cpu a component consumes
	CPU = "CPU"
	// CPUDescription is the name of the setting controlling the min and max CPU to same value
	CPUDescription = "The minimum and maximum CPU a component can consume"
)

var (
	supportedLocalParameterDescriptions = map[string]string{
		ComponentType: ComponentTypeDescription,
		ComponentName: ComponentNameDescription,
		MinMemory:     MinMemoryDescription,
		MaxMemory:     MaxMemoryDescription,
		Memory:        MemoryDescription,
		Ignore:        IgnoreDescription,
		MinCPU:        MinCPUDescription,
		MaxCPU:        MaxCPUDescription,
		CPU:           CPUDescription,
	}

	lowerCaseLocalParameters = util.GetLowerCaseParameters(GetLocallySupportedParameters())
)

// FormatLocallySupportedParameters outputs supported parameters and their description
func FormatLocallySupportedParameters() (result string) {
	for _, v := range GetLocallySupportedParameters() {
		result = result + v + " - " + supportedLocalParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Local Parameters:\n" + result
}

func asLocallySupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseLocalParameters[lower]
}

// GetLocallySupportedParameters returns the name of the supported global parameters
func GetLocallySupportedParameters() []string {
	return util.GetSortedKeys(supportedLocalParameterDescriptions)
}
