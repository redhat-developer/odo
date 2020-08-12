package preference

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/util"
)

const (
	GlobalConfigEnvName  = "GLOBALODOCONFIG"
	configFileName       = "preference.yaml"
	preferenceKind       = "Preference"
	preferenceAPIVersion = "odo.dev/v1alpha1"

	//DefaultTimeout for openshift server connection check (in seconds)
	DefaultTimeout = 1

	// DefaultPushTimeout is the default timeout for pods (in seconds)
	DefaultPushTimeout = 240

	// DefaultBuildTimeout is the default build timeout for pods (in seconds)
	DefaultBuildTimeout = 300

	// UpdateNotificationSetting is the name of the setting controlling update notification
	UpdateNotificationSetting = "UpdateNotification"

	// UpdateNotificationSettingDescription is human-readable description for the update notification setting
	UpdateNotificationSettingDescription = "Flag to control if an update notification is shown or not (Default: true)"

	// NamePrefixSetting is the name of the setting controlling name prefix
	NamePrefixSetting = "NamePrefix"

	// NamePrefixSettingDescription is human-readable description for the name prefix setting
	NamePrefixSettingDescription = "Use this value to set a default name prefix (Default: current directory name)"

	// TimeoutSetting is the name of the setting controlling timeout for connection check
	TimeoutSetting = "Timeout"

	// BuildTimeoutSetting is the name of the setting controlling BuildTimeout
	BuildTimeoutSetting = "BuildTimeout"

	// PushTimeoutSetting is the name of the setting controlling PushTimeout
	PushTimeoutSetting = "PushTimeout"

	// ExperimentalSetting is the name of the setting confrolling exposure of features in development/experimental mode
	ExperimentalSetting = "Experimental"

	// ExperimentalDescription is human-readable description for the experimental setting
	ExperimentalDescription = "Set this value to true to expose features in development/experimental mode"

	// PushTargetSetting is the name of the setting confrolling the push target for odo (docker or kube)
	PushTargetSetting = "PushTarget"

	// PushTargetDescription is human-readable description for the pushtarget setting
	PushTargetDescription = "Set this value to 'kube' or 'docker' to tell odo where to push applications to. (Default: kube)"

	// RegistryCacheTimeSetting is human-readable description for the registrycachetime setting
	RegistryCacheTimeSetting = "RegistryCacheTime"

	// Constants for PushTarget values

	// DockerPushTarget represents the value of the push target when it's set to Docker
	DockerPushTarget = "docker"

	// KubePushTarget represents the value of the push target when it's set to Kube
	KubePushTarget = "kube"

	// DefaultDevfileRegistryName is the name of default devfile registry
	DefaultDevfileRegistryName = "DefaultDevfileRegistry"

	// DefaultDevfileRegistryURL is the URL of default devfile registry
	DefaultDevfileRegistryURL = "https://github.com/odo-devfiles/registry"

	// DefaultRegistryCacheTime is time (in minutes) for how long odo will cache information from Devfile registry
	DefaultRegistryCacheTime = 15
)

// TimeoutSettingDescription is human-readable description for the timeout setting
var TimeoutSettingDescription = fmt.Sprintf("Timeout (in seconds) for OpenShift server connection check (Default: %d)", DefaultTimeout)

// PushTimeoutSettingDescription adds a description for PushTimeout
var PushTimeoutSettingDescription = fmt.Sprintf("PushTimeout (in seconds) for waiting for a Pod to come up (Default: %d)", DefaultPushTimeout)

// BuildTimeoutSettingDescription adds a description for BuildTimeout
var BuildTimeoutSettingDescription = fmt.Sprintf("BuildTimeout (in seconds) for waiting for a build of the git component to complete (Default: %d)", DefaultBuildTimeout)

// RegistryCacheTimeDescription adds a description for RegistryCacheTime
var RegistryCacheTimeDescription = fmt.Sprintf("For how long (in minutes) odo will cache information from Devfile registry (Default: %d)", DefaultRegistryCacheTime)

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

var (
	// records information on supported parameters
	supportedParameterDescriptions = map[string]string{
		UpdateNotificationSetting: UpdateNotificationSettingDescription,
		NamePrefixSetting:         NamePrefixSettingDescription,
		TimeoutSetting:            TimeoutSettingDescription,
		BuildTimeoutSetting:       BuildTimeoutSettingDescription,
		PushTimeoutSetting:        PushTimeoutSettingDescription,
		ExperimentalSetting:       ExperimentalDescription,
		PushTargetSetting:         PushTargetDescription,
		RegistryCacheTimeSetting:  RegistryCacheTimeDescription,
	}

	// set-like map to quickly check if a parameter is supported
	lowerCaseParameters = util.GetLowerCaseParameters(GetSupportedParameters())
)

// PreferenceInfo wraps the preference and provides helpers to
// serialize it.
type PreferenceInfo struct {
	Filename   string `yaml:"FileName,omitempty"`
	Preference `yaml:",omitempty"`
}

// OdoSettings holds all odo specific configurations
type OdoSettings struct {
	// Controls if an update notification is shown or not
	UpdateNotification *bool `yaml:"UpdateNotification,omitempty"`

	// Holds the prefix part of generated random application name
	NamePrefix *string `yaml:"NamePrefix,omitempty"`

	// Timeout for OpenShift server connection check
	Timeout *int `yaml:"Timeout,omitempty"`

	// BuildTimeout for OpenShift build timeout check
	BuildTimeout *int `yaml:"BuildTimeout,omitempty"`

	// PushTimeout for OpenShift pod timeout check
	PushTimeout *int `yaml:"PushTimeout,omitempty"`

	// Experimental for exposing features in development/experimental mode
	Experimental *bool `yaml:"Experimental,omitempty"`

	// PushTarget for telling odo which platform to push to (either kube or docker)
	PushTarget *string `yaml:"PushTarget,omitempty"`

	// RegistryList for telling odo to connect to all the registries in the registry list
	RegistryList *[]Registry `yaml:"RegistryList,omitempty"`

	// RegistryCacheTime how long odo should cache information from registry
	RegistryCacheTime *int `yaml:"RegistryCacheTime,omitempty"`
}

// Registry includes the registry metadata
type Registry struct {
	Name   string `yaml:"Name,omitempty"`
	URL    string `yaml:"URL,omitempty"`
	Secure bool
}

// Preference stores all the preferences related to odo
type Preference struct {
	metav1.TypeMeta `yaml:",inline"`

	// Odo settings holds the odo specific global settings
	OdoSettings OdoSettings `yaml:"OdoSettings,omitempty"`
}

func getPreferenceFile() (string, error) {
	if env, ok := os.LookupEnv(GlobalConfigEnvName); ok {
		return env, nil
	}

	if len(customHomeDir) != 0 {
		return filepath.Join(customHomeDir, ".odo", configFileName), nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentUser.HomeDir, ".odo", configFileName), nil
}

// New returns the PreferenceInfo to retain the expected behavior
func New() (*PreferenceInfo, error) {
	return NewPreferenceInfo()
}

// NewPreference creates an empty Preference struct with type meta information
func NewPreference() Preference {
	return Preference{
		TypeMeta: metav1.TypeMeta{
			Kind:       preferenceKind,
			APIVersion: preferenceAPIVersion,
		},
	}
}

// NewPreferenceInfo gets the PreferenceInfo from preference file and creates the preference file in case it's
// not present
func NewPreferenceInfo() (*PreferenceInfo, error) {
	preferenceFile, err := getPreferenceFile()
	klog.V(4).Infof("The path for preference file is %+v", preferenceFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo preference file")
	}

	c := PreferenceInfo{
		Preference: NewPreference(),
		Filename:   preferenceFile,
	}

	// If the preference file doesn't exist then we return with default preference
	if _, err = os.Stat(preferenceFile); os.IsNotExist(err) {
		// Handle user has preference file but doesn't use dynamic registry before
		defaultRegistryList := []Registry{
			{
				Name:   DefaultDevfileRegistryName,
				URL:    DefaultDevfileRegistryURL,
				Secure: false,
			},
		}
		c.OdoSettings.RegistryList = &defaultRegistryList
		return &c, nil
	}

	err = util.GetFromFile(&c.Preference, c.Filename)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// RegistryHandler handles registry add, update and delete operations
func (c *PreferenceInfo) RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error {
	var registryList []Registry
	var err error
	registryExist := false

	// Registry list is empty
	if c.OdoSettings.RegistryList == nil {
		registryList, err = handleWithoutRegistryExist(registryList, operation, registryName, registryURL, isSecure)
		if err != nil {
			return err
		}
	} else {
		// The target registry exists in the registry list
		registryList = *c.OdoSettings.RegistryList
		for index, registry := range registryList {
			if registry.Name == registryName {
				registryExist = true
				registryList, err = handleWithRegistryExist(index, registryList, operation, registryName, registryURL, forceFlag, isSecure)
				if err != nil {
					return err
				}
			}
		}

		// The target registry doesn't exist in the registry list
		if !registryExist {
			registryList, err = handleWithoutRegistryExist(registryList, operation, registryName, registryURL, isSecure)
			if err != nil {
				return err
			}
		}
	}

	c.OdoSettings.RegistryList = &registryList
	err = util.WriteToFile(&c.Preference, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to write the configuration of %s operation to preference file", operation)
	}

	return nil
}

func handleWithoutRegistryExist(registryList []Registry, operation string, registryName string, registryURL string, isSecure bool) ([]Registry, error) {
	switch operation {

	case "add":
		registry := Registry{
			Name:   registryName,
			URL:    registryURL,
			Secure: isSecure,
		}
		registryList = append(registryList, registry)

	case "update":
		return nil, errors.Errorf("failed to update registry: registry %s doesn't exist", registryName)

	case "delete":
		return nil, errors.Errorf("failed to delete registry: registry %s doesn't exist", registryName)
	}

	return registryList, nil
}

func handleWithRegistryExist(index int, registryList []Registry, operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) ([]Registry, error) {
	switch operation {

	case "add":
		return nil, errors.Errorf("failed to add registry: registry %s already exists", registryName)

	case "update":
		if !forceFlag {
			if !ui.Proceed(fmt.Sprintf("Are you sure you want to update registry %s", registryName)) {
				log.Info("Aborted by the user")
				return registryList, nil
			}
		}

		registryList[index].URL = registryURL
		registryList[index].Secure = isSecure
		log.Info("Successfully updated registry")

	case "delete":
		if !forceFlag {
			if !ui.Proceed(fmt.Sprintf("Are you sure you want to delete registry %s", registryName)) {
				log.Info("Aborted by the user")
				return registryList, nil
			}
		}

		copy(registryList[index:], registryList[index+1:])
		registryList[len(registryList)-1] = Registry{}
		registryList = registryList[:len(registryList)-1]
		log.Info("Successfully deleted registry")
	}

	return registryList, nil
}

// SetConfiguration modifies Odo configurations in the config file
// as of now being used for nameprefix, timeout, updatenotification
// TODO: Use reflect to set parameters
func (c *PreferenceInfo) SetConfiguration(parameter string, value string) error {
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

		case "buildtimeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.BuildTimeout = &typedval

		case "pushtimeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.PushTimeout = &typedval

		case "registrycachetime":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.RegistryCacheTime = &typedval

		case "updatenotification":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			c.OdoSettings.UpdateNotification = &val

		case "nameprefix":
			c.OdoSettings.NamePrefix = &value

		case "experimental":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			c.OdoSettings.Experimental = &val

		case "pushtarget":
			val := strings.ToLower(value)
			if val != DockerPushTarget && val != KubePushTarget {
				return errors.Errorf("cannot set pushtarget to values other than '%s' or '%s'", DockerPushTarget, KubePushTarget)
			}
			c.OdoSettings.PushTarget = &val
		}
	} else {
		return errors.Errorf("unknown parameter :'%s' is not a parameter in odo preference", parameter)
	}

	err := util.WriteToFile(&c.Preference, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to set %s", parameter)
	}
	return nil
}

// DeleteConfiguration delete Odo configurations in the global config file
// as of now being used for nameprefix, timeout, updatenotification
func (c *PreferenceInfo) DeleteConfiguration(parameter string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names

		if err := util.DeleteConfiguration(&c.OdoSettings, p); err != nil {
			return err
		}
	} else {
		return errors.Errorf("unknown parameter :'%s' is not a parameter in the odo preference", parameter)
	}

	err := util.WriteToFile(&c.Preference, c.Filename)
	if err != nil {
		return errors.Wrapf(err, "unable to set %s", parameter)
	}
	return nil
}

// IsSet checks if the value is set in the preference
func (c *PreferenceInfo) IsSet(parameter string) bool {
	return util.IsSet(c.OdoSettings, parameter)
}

// GetTimeout returns the value of Timeout from config
// and if absent then returns default
func (c *PreferenceInfo) GetTimeout() int {
	// default timeout value is 1
	return util.GetIntOrDefault(c.OdoSettings.Timeout, DefaultTimeout)
}

// GetBuildTimeout gets the value set by BuildTimeout
func (c *PreferenceInfo) GetBuildTimeout() int {
	// default timeout value is 300
	return util.GetIntOrDefault(c.OdoSettings.BuildTimeout, DefaultBuildTimeout)
}

// GetPushTimeout gets the value set by PushTimeout
func (c *PreferenceInfo) GetPushTimeout() int {
	// default timeout value is 1
	return util.GetIntOrDefault(c.OdoSettings.PushTimeout, DefaultPushTimeout)
}

// GetRegistryCacheTime gets the value set by RegistryCacheTime
func (c *PreferenceInfo) GetRegistryCacheTime() int {
	return util.GetIntOrDefault(c.OdoSettings.RegistryCacheTime, DefaultRegistryCacheTime)
}

// GetUpdateNotification returns the value of UpdateNotification from preferences
// and if absent then returns default
func (c *PreferenceInfo) GetUpdateNotification() bool {
	return util.GetBoolOrDefault(c.OdoSettings.UpdateNotification, true)
}

// GetNamePrefix returns the value of Prefix from preferences
// and if absent then returns default
func (c *PreferenceInfo) GetNamePrefix() string {
	return util.GetStringOrEmpty(c.OdoSettings.NamePrefix)
}

// GetExperimental returns the value of Experimental from preferences
// and if absent then returns default
// default value: false, experimental mode is disabled by default
func (c *PreferenceInfo) GetExperimental() bool {
	return util.GetBoolOrDefault(c.OdoSettings.Experimental, false)
}

// GetPushTarget returns the value of PushTarget from preferences
// and if absent then returns defualt
// default value: kube, docker push target needs to be manually enabled
func (c *PreferenceInfo) GetPushTarget() string {
	return util.GetStringOrDefault(c.OdoSettings.PushTarget, KubePushTarget)
}

// FormatSupportedParameters outputs supported parameters and their description
func FormatSupportedParameters() (result string) {
	for _, v := range GetSupportedParameters() {
		result = result + " " + v + " - " + supportedParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Global Parameters:\n" + result
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
