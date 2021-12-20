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

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/util"
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

	// RegistryCacheTimeSetting is human-readable description for the registrycachetime setting
	RegistryCacheTimeSetting = "RegistryCacheTime"

	// DefaultDevfileRegistryName is the name of default devfile registry
	DefaultDevfileRegistryName = "DefaultDevfileRegistry"

	// DefaultDevfileRegistryURL is the URL of default devfile registry
	DefaultDevfileRegistryURL = "https://registry.devfile.io"

	// OldDefaultDevfileRegistryURL is the URL of old default devfile registry for registry migration purpose
	OldDefaultDevfileRegistryURL = "https://github.com/odo-devfiles/registry"

	// DefaultRegistryCacheTime is time (in minutes) for how long odo will cache information from Devfile registry
	DefaultRegistryCacheTime = 15

	// EphemeralSetting specifies if ephemeral volumes needs to be used as source volume.
	EphemeralSetting = "Ephemeral"

	// DefaultEphemeralSettings is a default value for Ephemeral preference
	DefaultEphemeralSettings = true

	// ConsentTelemetrySettings specifies if the user consents to telemetry
	ConsentTelemetrySetting = "ConsentTelemetry"

	// DefaultConsentTelemetry is a default value for ConsentTelemetry preference
	DefaultConsentTelemetrySetting = false
)

// TimeoutSettingDescription is human-readable description for the timeout setting
var TimeoutSettingDescription = fmt.Sprintf("Timeout (in seconds) for OpenShift server connection check (Default: %d)", DefaultTimeout)

// PushTimeoutSettingDescription adds a description for PushTimeout
var PushTimeoutSettingDescription = fmt.Sprintf("PushTimeout (in seconds) for waiting for a Pod to come up (Default: %d)", DefaultPushTimeout)

// BuildTimeoutSettingDescription adds a description for BuildTimeout
var BuildTimeoutSettingDescription = fmt.Sprintf("BuildTimeout (in seconds) for waiting for a build of the git component to complete (Default: %d)", DefaultBuildTimeout)

// RegistryCacheTimeDescription adds a description for RegistryCacheTime
var RegistryCacheTimeDescription = fmt.Sprintf("For how long (in minutes) odo will cache information from Devfile registry (Default: %d)", DefaultRegistryCacheTime)

// EphemeralDescription adds a description for EphemeralSourceVolume
var EphemeralDescription = fmt.Sprintf("If true, odo will create an emptyDir volume to store source code (Default: %t)", DefaultEphemeralSettings)

//TelemetryConsentDescription adds a description for TelemetryConsentSetting
var ConsentTelemetryDescription = fmt.Sprintf("If true, odo will collect telemetry for the user's odo usage (Default: %t)\n\t\t    For more information: https://developers.redhat.com/article/tool-data-collection", DefaultConsentTelemetrySetting)

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
		RegistryCacheTimeSetting:  RegistryCacheTimeDescription,
		EphemeralSetting:          EphemeralDescription,
		ConsentTelemetrySetting:   ConsentTelemetryDescription,
	}

	// set-like map to quickly check if a parameter is supported
	lowerCaseParameters = util.GetLowerCaseParameters(GetSupportedParameters())
)

// odoSettings holds all odo specific configurations
type odoSettings struct {
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

	// RegistryList for telling odo to connect to all the registries in the registry list
	RegistryList *[]Registry `yaml:"RegistryList,omitempty"`

	// RegistryCacheTime how long odo should cache information from registry
	RegistryCacheTime *int `yaml:"RegistryCacheTime,omitempty"`

	// Ephemeral if true creates odo emptyDir to store odo source code
	Ephemeral *bool `yaml:"Ephemeral,omitempty"`

	// ConsentTelemetry if true collects telemetry for odo
	ConsentTelemetry *bool `yaml:"ConsentTelemetry,omitempty"`
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
	OdoSettings odoSettings `yaml:"OdoSettings,omitempty"`
}

// preferenceInfo wraps the preference and provides helpers to
// serialize it.
type preferenceInfo struct {
	Filename   string `yaml:"FileName,omitempty"`
	Preference `yaml:",omitempty"`
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

func NewClient() (Client, error) {
	return newPreferenceInfo()
}

// newPreference creates an empty Preference struct with type meta information
func newPreference() Preference {
	return Preference{
		TypeMeta: metav1.TypeMeta{
			Kind:       preferenceKind,
			APIVersion: preferenceAPIVersion,
		},
	}
}

// newPreferenceInfo gets the PreferenceInfo from preference file and creates the preference file in case it's
// not present
func newPreferenceInfo() (*preferenceInfo, error) {
	preferenceFile, err := getPreferenceFile()
	klog.V(4).Infof("The path for preference file is %+v", preferenceFile)
	if err != nil {
		return nil, errors.Errorf("unable to get odo preference file, run the command with more verbosity to check if the preference path is correct")
	}

	c := preferenceInfo{
		Preference: newPreference(),
		Filename:   preferenceFile,
	}

	// Default devfile registry
	defaultRegistryList := []Registry{
		{
			Name:   DefaultDevfileRegistryName,
			URL:    DefaultDevfileRegistryURL,
			Secure: false,
		},
	}

	// If the preference file doesn't exist then we return with default preference
	if _, err = os.Stat(preferenceFile); os.IsNotExist(err) {
		c.OdoSettings.RegistryList = &defaultRegistryList
		return &c, nil
	}

	err = util.GetFromFile(&c.Preference, c.Filename)
	if err != nil {
		return nil, err
	}

	// Handle user has preference file but doesn't use dynamic registry before
	if c.OdoSettings.RegistryList == nil {
		c.OdoSettings.RegistryList = &defaultRegistryList
	}

	// Handle OCI-based default registry migration
	if c.OdoSettings.RegistryList != nil {
		for index, registry := range *c.OdoSettings.RegistryList {
			if registry.Name == DefaultDevfileRegistryName && registry.URL == OldDefaultDevfileRegistryURL {
				registryList := *c.OdoSettings.RegistryList
				registryList[index].URL = DefaultDevfileRegistryURL
				break
			}
		}
	}

	return &c, nil
}

// RegistryHandler handles registry add, update and delete operations
func (c *preferenceInfo) RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error {
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
		return errors.Errorf("unable to write the configuration of %q operation to preference file", operation)
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

	case "update", "delete":
		return nil, errors.Errorf("failed to %v registry: registry %q doesn't exist", operation, registryName)
	}

	return registryList, nil
}

func handleWithRegistryExist(index int, registryList []Registry, operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) ([]Registry, error) {
	switch operation {

	case "add":
		return nil, errors.Errorf("failed to add registry: registry %q already exists", registryName)

	case "update":
		if !forceFlag {
			if !ui.Proceed(fmt.Sprintf("Are you sure you want to update registry %q", registryName)) {
				log.Info("Aborted by the user")
				return registryList, nil
			}
		}

		registryList[index].URL = registryURL
		registryList[index].Secure = isSecure
		log.Info("Successfully updated registry")

	case "delete":
		if !forceFlag {
			if !ui.Proceed(fmt.Sprintf("Are you sure you want to delete registry %q", registryName)) {
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
func (c *preferenceInfo) SetConfiguration(parameter string, value string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names
		switch p {

		case "timeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Errorf("unable to set %q to %q", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.Timeout = &typedval

		case "buildtimeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be an integer", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.BuildTimeout = &typedval

		case "pushtimeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be an integer", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.PushTimeout = &typedval

		case "registrycachetime":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be an integer", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.RegistryCacheTime = &typedval

		case "updatenotification":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.UpdateNotification = &val

		//	TODO: should we add a validator here? What is the use of nameprefix?
		case "nameprefix":
			c.OdoSettings.NamePrefix = &value

		case "ephemeral":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.Ephemeral = &val

		case "consenttelemetry":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return errors.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.ConsentTelemetry = &val
		}
	} else {
		return errors.Errorf("unknown parameter : %q is not a parameter in odo preference, run help to see list of available parameters", parameter)
	}

	err := util.WriteToFile(&c.Preference, c.Filename)
	if err != nil {
		return errors.Errorf("unable to set %q, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md", parameter)
	}
	return nil
}

// DeleteConfiguration delete Odo configurations in the global config file
// as of now being used for nameprefix, timeout, updatenotification
func (c *preferenceInfo) DeleteConfiguration(parameter string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names

		if err := util.DeleteConfiguration(&c.OdoSettings, p); err != nil {
			return err
		}
	} else {
		return errors.Errorf("unknown parameter :%q is not a parameter in the odo preference", parameter)
	}

	err := util.WriteToFile(&c.Preference, c.Filename)
	if err != nil {
		return errors.Errorf("unable to set %q, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md", parameter)
	}
	return nil
}

// IsSet checks if the value is set in the preference
func (c *preferenceInfo) IsSet(parameter string) bool {
	return util.IsSet(c.OdoSettings, parameter)
}

// GetTimeout returns the value of Timeout from config
// and if absent then returns default
func (c *preferenceInfo) GetTimeout() int {
	// default timeout value is 1
	return util.GetIntOrDefault(c.OdoSettings.Timeout, DefaultTimeout)
}

// GetBuildTimeout gets the value set by BuildTimeout
func (c *preferenceInfo) GetBuildTimeout() int {
	// default timeout value is 300
	return util.GetIntOrDefault(c.OdoSettings.BuildTimeout, DefaultBuildTimeout)
}

// GetPushTimeout gets the value set by PushTimeout
func (c *preferenceInfo) GetPushTimeout() int {
	// default timeout value is 1
	return util.GetIntOrDefault(c.OdoSettings.PushTimeout, DefaultPushTimeout)
}

// GetRegistryCacheTime gets the value set by RegistryCacheTime
func (c *preferenceInfo) GetRegistryCacheTime() int {
	return util.GetIntOrDefault(c.OdoSettings.RegistryCacheTime, DefaultRegistryCacheTime)
}

// GetUpdateNotification returns the value of UpdateNotification from preferences
// and if absent then returns default
func (c *preferenceInfo) GetUpdateNotification() bool {
	return util.GetBoolOrDefault(c.OdoSettings.UpdateNotification, true)
}

// GetEphemeralSourceVolume returns the value of ephemeral from preferences
// and if absent then returns default
func (c *preferenceInfo) GetEphemeralSourceVolume() bool {
	return util.GetBoolOrDefault(c.OdoSettings.Ephemeral, DefaultEphemeralSettings)
}

// GetNamePrefix returns the value of Prefix from preferences
// and if absent then returns default
func (c *preferenceInfo) GetNamePrefix() string {
	return util.GetStringOrEmpty(c.OdoSettings.NamePrefix)
}

// GetConsentTelemetry returns the value of ConsentTelemetry from preferences
// and if absent then returns default
// default value: false, consent telemetry is disabled by default
func (c *preferenceInfo) GetConsentTelemetry() bool {
	return util.GetBoolOrDefault(c.OdoSettings.ConsentTelemetry, DefaultConsentTelemetrySetting)
}

func (c *preferenceInfo) UpdateNotification() *bool {
	return c.OdoSettings.UpdateNotification
}

func (c *preferenceInfo) NamePrefix() *string {
	return c.OdoSettings.NamePrefix
}

func (c *preferenceInfo) Timeout() *int {
	return c.OdoSettings.Timeout
}

func (c *preferenceInfo) BuildTimeout() *int {
	return c.OdoSettings.BuildTimeout
}

func (c *preferenceInfo) PushTimeout() *int {
	return c.OdoSettings.PushTimeout
}

func (c *preferenceInfo) EphemeralSourceVolume() *bool {
	return c.OdoSettings.Ephemeral
}

func (c *preferenceInfo) ConsentTelemetry() *bool {
	return c.OdoSettings.ConsentTelemetry
}

func (c *preferenceInfo) RegistryList() *[]Registry {
	return c.OdoSettings.RegistryList
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
