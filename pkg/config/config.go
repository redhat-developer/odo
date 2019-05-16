package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	localConfigEnvName    = "LOCALODOCONFIG"
	configFileName        = "config.yaml"
	localConfigKind       = "LocalConfig"
	localConfigAPIVersion = "odo.openshift.io/v1alpha1"
)

type ComponentStorageSettings struct {
	Name string `yaml:"Name,omitempty"`
	Size string `yaml:"Size,omitempty"`
	Path string `yaml:"Path,omitempty"`
}

// ComponentSettings holds all component related information
type ComponentSettings struct {
	// The builder image to use
	Type *string `yaml:"Type,omitempty"`

	// SourceLocation is path to binary in current/context dir, it can be the
	// git url in case of source type being git
	SourceLocation *string `yaml:"SourceLocation,omitempty"`

	// Ref is component source git ref but can be levaraged for more in future
	Ref *string `yaml:"Ref,omitempty"`

	// Type is type of component source: git/local/binary
	SourceType *SrcType `yaml:"SourceType,omitempty"`

	// Ports is a slice of ports to be exposed when a component is created
	// the format of the port is "PORT/PROTOCOL" e.g. "8080/TCP"
	Ports *[]string `yaml:"Ports,omitempty"`

	Application *string `yaml:"Application,omitempty"`

	Project *string `yaml:"Project,omitempty"`

	Name *string `yaml:"Name,omitempty"`

	MinMemory *string `yaml:"MinMemory,omitempty"`

	MaxMemory *string `yaml:"MaxMemory,omitempty"`

	Storage *[]ComponentStorageSettings `yaml:"Storage,omitempty"`

	// Ignore if set to true then odoignore file should be considered
	Ignore *bool `yaml:"Ignore,omitempty"`

	MinCPU *string `yaml:"MinCPU,omitempty"`

	MaxCPU *string `yaml:"MaxCPU,omitempty"`

	Envs EnvVarList `yaml:"Envs,omitempty"`

	Url *[]ConfigUrl `yaml:"Url,omitempty"`
}

// ConfigUrl holds URL related information
type ConfigUrl struct {
	// Name of the URL
	Name string `yaml:"Name,omitempty"`
	// Port number for the url of the component, required in case of components which expose more than one service port
	Port int `yaml:"Port,omitempty"`
}

// LocalConfig holds all the config relavent to a specific Component.
type LocalConfig struct {
	typeMeta          metav1.TypeMeta   `yaml:",inline"`
	componentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// proxyLocalConfig holds all the parameter that local config does but exposes all
// of it, used for serialization.
type proxyLocalConfig struct {
	metav1.TypeMeta   `yaml:",inline"`
	ComponentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// LocalConfigInfo wraps the local config and provides helpers to
// serialize it.
type LocalConfigInfo struct {
	Filename         string `yaml:"FileName,omitempty"`
	LocalConfig      `yaml:",omitempty"`
	configFileExists bool
}

func getLocalConfigFile(cfgDir string) (string, error) {
	if env, ok := os.LookupEnv(localConfigEnvName); ok {
		return env, nil
	}

	if cfgDir == "" {
		var err error
		cfgDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(cfgDir, ".odo", configFileName), nil
}

// New returns the localConfigInfo
func New() (*LocalConfigInfo, error) {
	return NewLocalConfigInfo("")
}

// NewLocalConfigInfo gets the LocalConfigInfo from local config file and creates the local config file in case it's
// not present then it
func NewLocalConfigInfo(cfgDir string) (*LocalConfigInfo, error) {
	configFile, err := getLocalConfigFile(cfgDir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	c := LocalConfigInfo{
		LocalConfig:      NewLocalConfig(),
		Filename:         configFile,
		configFileExists: true,
	}

	// if the config file doesn't exist then we dont worry about it and return
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		c.configFileExists = false
		return &c, nil
	}
	err = getFromFile(&c.LocalConfig, c.Filename)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func getFromFile(lc *LocalConfig, filename string) error {
	plc := newProxyLocalConfig()

	err := util.GetFromFile(&plc, filename)
	if err != nil {
		return err
	}
	lc.typeMeta = plc.TypeMeta
	lc.componentSettings = plc.ComponentSettings
	return nil
}

// NewLocalConfig creates an empty LocalConfig struct with typeMeta populated
func NewLocalConfig() LocalConfig {
	return LocalConfig{
		typeMeta: metav1.TypeMeta{
			Kind:       localConfigKind,
			APIVersion: localConfigAPIVersion,
		},
	}
}

// newProxyLocalConfig creates an empty proxyLocalConfig struct with typeMeta populated
func newProxyLocalConfig() proxyLocalConfig {
	lc := NewLocalConfig()
	return proxyLocalConfig{
		TypeMeta: lc.typeMeta,
	}
}

// SetConfiguration sets the common config settings like component type, min memory
// max memory etc.
// TODO: Use reflect to set parameters
func (lci *LocalConfigInfo) SetConfiguration(parameter string, value interface{}) (err error) {
	// getting the second arg makes sure that this never panics
	strValue, _ := value.(string)
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "type":
			lci.componentSettings.Type = &strValue
		case "application":
			lci.componentSettings.Application = &strValue
		case "project":
			lci.componentSettings.Project = &strValue
		case "sourcetype":
			cmpSourceType, err := GetSrcType(strValue)
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, strValue)
			}
			lci.componentSettings.SourceType = &cmpSourceType
		case "ref":
			lci.componentSettings.Ref = &strValue
		case "sourcelocation":
			lci.componentSettings.SourceLocation = &strValue
		case "ports":
			arrValue := strings.Split(strValue, ",")
			lci.componentSettings.Ports = &arrValue
		case "name":
			lci.componentSettings.Name = &strValue
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
		case "storage":
			storageSetting, _ := value.(ComponentStorageSettings)
			if lci.componentSettings.Storage != nil {
				*lci.componentSettings.Storage = append(*lci.componentSettings.Storage, storageSetting)
			} else {
				lci.componentSettings.Storage = &[]ComponentStorageSettings{storageSetting}
			}
		case "cpu":
			lci.componentSettings.MinCPU = &strValue
			lci.componentSettings.MaxCPU = &strValue
		case "url":
			urlValue := value.(ConfigUrl)
			if lci.componentSettings.Url != nil {
				*lci.componentSettings.Url = append(*lci.componentSettings.Url, urlValue)
			} else {
				lci.componentSettings.Url = &[]ConfigUrl{urlValue}
			}
		}

		return lci.writeToFile()
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

// DeleteConfigDir Deletes the config directory with the config file
func (lci *LocalConfigInfo) DeleteConfigDir() error {
	return os.RemoveAll(filepath.Dir(lci.Filename))
}

// IsSet uses reflection to get the parameter from the localconfig struct, currently
// it only searches the componentSettings
func (lci *LocalConfigInfo) IsSet(parameter string) bool {

	switch strings.ToLower(parameter) {
	case "cpu":
		return (lci.componentSettings.MinCPU != nil && lci.componentSettings.MaxCPU != nil) &&
			(*lci.componentSettings.MinCPU == *lci.componentSettings.MaxCPU)
	case "memory":
		return (lci.componentSettings.MinMemory != nil && lci.componentSettings.MaxMemory != nil) &&
			(*lci.componentSettings.MinMemory == *lci.componentSettings.MaxMemory)
	}

	return util.IsSet(lci.componentSettings, parameter)
}

// ConfigFileExists if a config file exists or not
func (lci *LocalConfigInfo) ConfigFileExists() bool {
	return lci.configFileExists
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
		return lci.writeToFile()
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)

}

// DeleteUrl is used to delete config from local odo config
func (lci *LocalConfigInfo) DeleteUrl(parameter string) error {
	for i, url := range *lci.componentSettings.Url {
		if url.Name == parameter {
			s := *lci.componentSettings.Url
			s = append(s[:i], s[i+1:]...)
			lci.componentSettings.Url = &s
		}
	}
	return lci.writeToFile()
}

// DeleteFromConfigurationList is used to delete a value from a list from the local odo config
// parameter is the name of the config parameter
// value is the value to be deleted
func (lci *LocalConfigInfo) DeleteFromConfigurationList(parameter string, value string) error {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "storage":
			for i, storage := range lci.GetStorage() {
				if storage.Name == value {
					*lci.componentSettings.Storage = append((*lci.componentSettings.Storage)[:i], (*lci.componentSettings.Storage)[i+1:]...)
				}
			}
			return lci.writeToFile()
		}
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in local odo config", parameter)
}

// GetComponentSettings returns the componentSettings from local config
func (lci *LocalConfigInfo) GetComponentSettings() ComponentSettings {
	return lci.componentSettings
}

// SetComponentSettings sets the componentSettings from to the local config and writes to the file
func (lci *LocalConfigInfo) SetComponentSettings(cs ComponentSettings) error {
	lci.componentSettings = cs
	return lci.writeToFile()
}

// SetEnvVars sets the env variables on the component settings
func (lci *LocalConfigInfo) SetEnvVars(envVars EnvVarList) error {
	lci.componentSettings.Envs = envVars
	return lci.writeToFile()
}

// GetEnvVars gets the env variables from the component settings
func (lci *LocalConfigInfo) GetEnvVars() EnvVarList {
	if lci.componentSettings.Envs == nil {
		return EnvVarList{}
	}
	return lci.componentSettings.Envs
}

func (lci *LocalConfigInfo) writeToFile() error {
	plc := newProxyLocalConfig()
	plc.TypeMeta = lci.typeMeta
	plc.ComponentSettings = lci.componentSettings
	return util.WriteToFile(&plc, lci.Filename)
}

// GetType returns type of component (builder image name) in the config
func (lc *LocalConfig) GetType() string {
	if lc.componentSettings.Type == nil {
		return ""
	}
	return *lc.componentSettings.Type
}

// GetSourceLocation returns the sourcelocation, returns default if nil
func (lc *LocalConfig) GetSourceLocation() string {
	if lc.componentSettings.SourceLocation == nil {
		return ""
	}
	return *lc.componentSettings.SourceLocation
}

// GetRef returns the ref, returns default if nil
func (lc *LocalConfig) GetRef() string {
	if lc.componentSettings.Ref == nil {
		return ""
	}
	return *lc.componentSettings.Ref
}

// GetSourceType returns the source type, returns default if nil
func (lc *LocalConfig) GetSourceType() SrcType {
	if lc.componentSettings.SourceType == nil {
		return ""
	}
	return *lc.componentSettings.SourceType
}

// GetPorts returns the ports, returns default if nil
func (lc *LocalConfig) GetPorts() []string {
	if lc.componentSettings.Ports == nil {
		return nil
	}
	return *lc.componentSettings.Ports
}

// GetApplication returns the app, returns default if nil
func (lc *LocalConfig) GetApplication() string {
	if lc.componentSettings.Application == nil {
		return ""
	}
	return *lc.componentSettings.Application
}

// GetProject returns the project, returns default if nil
func (lc *LocalConfig) GetProject() string {
	if lc.componentSettings.Project == nil {
		return ""
	}
	return *lc.componentSettings.Project
}

// GetName returns the Name, returns default if nil
func (lc *LocalConfig) GetName() string {
	if lc.componentSettings.Name == nil {
		return ""
	}
	return *lc.componentSettings.Name
}

// GetMinMemory returns the MinMemory, returns default if nil
func (lc *LocalConfig) GetMinMemory() string {
	if lc.componentSettings.MinMemory == nil {
		return ""
	}
	return *lc.componentSettings.MinMemory
}

// GetMaxMemory returns the MaxMemory, returns default if nil
func (lc *LocalConfig) GetMaxMemory() string {
	if lc.componentSettings.MaxMemory == nil {
		return ""
	}
	return *lc.componentSettings.MaxMemory
}

// GetIgnore returns the Ignore, returns default if nil
func (lc *LocalConfig) GetIgnore() bool {
	if lc.componentSettings.Ignore == nil {
		return false
	}
	return *lc.componentSettings.Ignore
}

// GetMinCPU returns the MinCPU, returns default if nil
func (lc *LocalConfig) GetMinCPU() string {
	if lc.componentSettings.MinCPU == nil {
		return ""
	}
	return *lc.componentSettings.MinCPU
}

// GetMaxCPU returns the MaxCPU, returns default if nil
func (lc *LocalConfig) GetMaxCPU() string {
	if lc.componentSettings.MaxCPU == nil {
		return ""
	}
	return *lc.componentSettings.MaxCPU
}

// GetUrl returns the ConfigUrl, returns default if nil
func (lc *LocalConfig) GetUrl() []ConfigUrl {
	if lc.componentSettings.Url == nil {
		return []ConfigUrl{}
	}
	return *lc.componentSettings.Url
}

// GetStorage returns the Storage, returns empty if nil
func (lc *LocalConfig) GetStorage() []ComponentStorageSettings {
	if lc.componentSettings.Storage == nil {
		return []ComponentStorageSettings{}
	}
	return *lc.componentSettings.Storage
}

const (
	// Type is the name of the setting controlling the component type i.e. builder image
	Type = "Type"
	// TypeDescription is human-readable description of the componentType setting
	TypeDescription = "The type of component"
	// Name is the name of the setting controlling the component name
	Name = "Name"
	// NameDescription is human-readable description of the componentType setting
	NameDescription = "The name of the component"
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
	MemoryDescription = "The minimum and maximum memory a component can consume"
	// Ignore is the name of the setting controlling the min memory a component consumes
	Ignore = "Ignore"
	// IgnoreDescription is the name of the setting controlling the use of .odoignore file
	IgnoreDescription = "Consider the .odoignore file for push and watch"
	// MinCPU is the name of the setting controlling minimum cpu
	MinCPU = "MinCPU"
	// MinCPUDescription is the name of the setting controlling the min CPU value
	MinCPUDescription = "The minimum CPU a component can consume"
	// MaxCPU is the name of the setting controlling the use of .odoignore file
	MaxCPU = "MaxCPU"
	//MaxCPUDescription is the name of the setting controlling the max CPU value
	MaxCPUDescription = "The maximum CPU a component can consume"
	// CPU is the name of the setting controlling the cpu a component consumes
	CPU = "CPU"
	// CPUDescription is the name of the setting controlling the min and max CPU to same value
	CPUDescription = "The minimum and maximum CPU a component can consume"
	// SourceLocation indicates path of the source e.g. location of the git repo
	SourceLocation = "SourceLocation"
	// SourceType indicates type of component source -- git/binary/local
	SourceType = "SourceType"
	// Ref indicates git ref for the component source
	Ref = "Ref"
	// Ports is the space separated list of user specified ports to be opened in the component
	Ports = "Ports"
	// Application indicates application of which component is part of
	Application = "Application"
	// Project indicates project the component is part of
	Project = "Project"
	// ProjectDescription is the description of project component setting
	ProjectDescription = "Project is the name of the project the component is part of"
	// ApplicationDescription is the description of app component setting
	ApplicationDescription = "Application is the name of application the component needs to be part of"
	// PortsDescription is the desctription of the ports component setting
	PortsDescription = "Ports to be opened in the component"
	// RefDescription is the description of ref setting
	RefDescription = "Git ref to use for creating component from git source"
	// SourceTypeDescription is the description of type setting
	SourceTypeDescription = "Type of component source - git/binary/local"
	// Storage is the name of the setting controlling storage
	Storage = "Storage"
	// StorageDescription is the description of the storage
	StorageDescription = "Storage of the component"
	// SourceLocationDescription is the human-readable description of path setting
	SourceLocationDescription = "The path indicates the location of binary file or git source"
	// Url
	Url = "Url"
	// UrlDescription is the description of URL
	UrlDescription = "URL to access the component"
)

var (
	supportedLocalParameterDescriptions = map[string]string{
		Type:           TypeDescription,
		Name:           NameDescription,
		Application:    ApplicationDescription,
		Project:        ProjectDescription,
		SourceLocation: SourceLocationDescription,
		SourceType:     SourceTypeDescription,
		Ref:            RefDescription,
		Ports:          PortsDescription,
		MinMemory:      MinMemoryDescription,
		MaxMemory:      MaxMemoryDescription,
		Memory:         MemoryDescription,
		Ignore:         IgnoreDescription,
		MinCPU:         MinCPUDescription,
		MaxCPU:         MaxCPUDescription,
		Storage:        StorageDescription,
		CPU:            CPUDescription,
		Url:            UrlDescription,
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

// SrcType is an enum to indicate the type of source of component -- local source/binary or git for the generation of app/component names
type SrcType string

const (
	// GIT as source of component
	GIT SrcType = "git"
	// LOCAL Local source path as source of component
	LOCAL SrcType = "local"
	// BINARY Local Binary as source of component
	BINARY SrcType = "binary"
	// NONE indicates there's no information about the type of source of the component
	NONE SrcType = ""
)

// GetSrcType returns enum equivalent of passed component source type or error if unsupported type passed
func GetSrcType(ctStr string) (SrcType, error) {
	switch strings.ToLower(ctStr) {
	case string(GIT):
		return GIT, nil
	case string(LOCAL):
		return LOCAL, nil
	case string(BINARY):
		return BINARY, nil
	default:
		return NONE, fmt.Errorf("Unsupported component source type: %s", ctStr)
	}
}

// GetOSSourcePath corrects the current sourcePath depending on local or binary configuration,
// if Git has been passed, we simply return the source location from LocalConfig
// this will get the correct source path whether on Windows, macOS or Linux.
//
// This function also takes in the current working directory + context directory in order
// to correctly retrieve WHERE the source is located..
func (lci *LocalConfigInfo) GetOSSourcePath() (path string, err error) {

	sourceType := lci.GetSourceType()
	sourceLocation := lci.GetSourceLocation()

	// Get the component context folder
	// ".odo" is removed as lci.Filename will always return the '.odo' folder.. we don't need that!
	componentContext := strings.Trim(filepath.Dir(lci.Filename), ".odo")

	if sourceLocation == "" {
		return "", fmt.Errorf("Blank source location provided")
	}

	if sourceType == GIT {
		glog.V(4).Info("Git source type detected, not correcting SourcePath location")
		return sourceLocation, nil
	}

	// Validation check if the user passes in a URL despite us being LOCAL or BINARY
	u, err := url.Parse(sourceLocation)
	if err != nil || (u.Scheme == "https" || u.Scheme == "http") {
		return "", fmt.Errorf("URL: %s passed even though source type is: %s", sourceLocation, sourceType)
	}

	// Always piped to "fromslash" so it's correct for the OS..
	// after retrieving the sourceLocation we will covert it to the
	// correct source path depending on the OS.
	absPath, err := util.GetAbsPath(filepath.Join(componentContext, lci.GetSourceLocation()))

	sourceOSPath := filepath.FromSlash(absPath)

	return sourceOSPath, nil
}
