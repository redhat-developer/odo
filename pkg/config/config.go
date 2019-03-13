package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
)

const (
	localConfigEnvName    = "LOCALODOCONFIG"
	configFileName        = "config.yaml"
	localConfigKind       = "LocalConfig"
	localConfigAPIVersion = "odo.openshift.io/v1alpha1"
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
	metav1.TypeMeta   `yaml:",inline"`
	componentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// ProxyLocalConfig holds all the parameter that local config does but exposes all
// of it, used for serialization.
type ProxyLocalConfig struct {
	metav1.TypeMeta   `yaml:",inline"`
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
	return NewLocalConfigInfo()
}

// NewLocalConfigInfo gets the LocalConfigInfo from local config file and creates the local config file in case it's
// not present then it
func NewLocalConfigInfo() (*LocalConfigInfo, error) {
	configFile, err := getLocalConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	c := LocalConfigInfo{
		LocalConfig: NewLocalConfig(),
		Filename:    configFile,
	}

	// if the config file doesn't exist then we dont worry about it and return
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		return &c, nil
	}
	err = getFromFile(&c.LocalConfig, c.Filename)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func getFromFile(lc *LocalConfig, filename string) error {
	plc := NewProxyLocalConfig()

	err := util.GetFromFile(&plc, filename)
	if err != nil {
		return err
	}
	lc.TypeMeta = plc.TypeMeta
	lc.componentSettings = plc.ComponentSettings
	return nil
}

func writeToFile(lc *LocalConfig, filename string) error {
	plc := NewProxyLocalConfig()
	plc.TypeMeta = lc.TypeMeta
	plc.ComponentSettings = lc.componentSettings
	return util.WriteToFile(&plc, filename)
}

// NewLocalConfig creates an empty LocalConfig struct with typeMeta populated
func NewLocalConfig() LocalConfig {
	return LocalConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       localConfigKind,
			APIVersion: localConfigAPIVersion,
		},
	}
}

// NewProxyLocalConfig creates an empty ProxyLocalConfig struct with typeMeta populated
func NewProxyLocalConfig() ProxyLocalConfig {
	lc := NewLocalConfig()
	return ProxyLocalConfig{
		TypeMeta: lc.TypeMeta,
	}
}

// SetConfiguration sets the common config settings like component type, min memory
// max memory etc.
// TODO: Use reflect to set parameters
func (lci *LocalConfigInfo) SetConfiguration(parameter string, value interface{}) (err error) {
	strValue := value.(string)
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "componenttype":
			lci.componentSettings.ComponentType = &strValue
		case "componentname":
			lci.componentSettings.ComponentName = &strValue
		case "minmemory":
			lci.componentSettings.MinMemory = &strValue
		case "maxmemory":
			lci.componentSettings.MaxMemory = &strValue
		case "memory":
			lci.componentSettings.MaxMemory = &strValue
			lci.componentSettings.MinMemory = &strValue
		case "ignore":
			val, err := strconv.ParseBool(strings.ToLower(strValue))
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, strValue)
			}
			lci.componentSettings.Ignore = &val
		case "mincpu":
			lci.componentSettings.MinCPU = &strValue
		case "maxcpu":
			lci.componentSettings.MaxCPU = &strValue
		case "cpu":
			lci.componentSettings.MinCPU = &strValue
			lci.componentSettings.MaxCPU = &strValue

		}

		return writeToFile(&lci.LocalConfig, lci.Filename)
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

// GetConfiguration uses reflection to get the parameter from the localconfig struct, currently
// it only searches the componentSettings
func (lci *LocalConfigInfo) GetConfiguration(parameter string) (interface{}, bool) {

	switch strings.ToLower(parameter) {
	case "cpu":
		if lci.componentSettings.MinCPU == nil {
			return nil, true
		}
		return *lci.componentSettings.MinCPU, true
	case "memory":
		if lci.componentSettings.MinMemory == nil {
			return nil, true
		}
		return *lci.componentSettings.MinMemory, true
	}

	return util.GetConfiguration(lci.componentSettings, parameter)
}

// DeleteConfiguration is used to delete config from local odo config
func (lci *LocalConfigInfo) DeleteConfiguration(parameter string) error {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {

		switch parameter {
		case "cpu":
			lci.componentSettings.MinCPU = nil
			lci.componentSettings.MaxCPU = nil
		case "memory":
			lci.componentSettings.MinMemory = nil
			lci.componentSettings.MaxMemory = nil
		default:
			if err := util.DeleteConfiguration(&lci.componentSettings, parameter); err != nil {
				return err
			}
		}
		return writeToFile(&lci.LocalConfig, lci.Filename)
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

func (lci *LocalConfigInfo) GetComponentSettings() ComponentSettings {
	return lci.componentSettings
}

func (lci *LocalConfigInfo) SetComponentSettings(cs ComponentSettings) error {
	lci.componentSettings = cs
	return writeToFile(&lci.LocalConfig, lci.Filename)
}

// GetComponentType returns type of component (builder image name) in the config
// and if absent then returns default
func (lc *LocalConfig) GetComponentType() string {
	if lc.componentSettings.ComponentType == nil {
		return ""
	}
	return *lc.componentSettings.ComponentType
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
