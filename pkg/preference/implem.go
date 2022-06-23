package preference

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	kpointer "k8s.io/utils/pointer"
)

// odoSettings holds all odo specific configurations
// these configurations are applicable across the odo components
type odoSettings struct {
	// Controls if an update notification is shown or not
	UpdateNotification *bool `yaml:"UpdateNotification,omitempty"`

	// Timeout for server connection check
	Timeout *time.Duration `yaml:"Timeout,omitempty"`

	// PushTimeout for pod timeout check
	PushTimeout *time.Duration `yaml:"PushTimeout,omitempty"`

	// RegistryList for telling odo to connect to all the registries in the registry list
	RegistryList *[]Registry `yaml:"RegistryList,omitempty"`

	// RegistryCacheTime how long odo should cache information from registry
	RegistryCacheTime *time.Duration `yaml:"RegistryCacheTime,omitempty"`

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

// newPreferenceInfo gets the PreferenceInfo from preference file
// or returns default PreferenceInfo if preference file does not exist
func newPreferenceInfo() (*preferenceInfo, error) {
	preferenceFile, err := getPreferenceFile()
	klog.V(4).Infof("The path for preference file is %+v", preferenceFile)
	if err != nil {
		return nil, err
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

	// TODO: This code block about logging warnings should be removed once users completely shift to odo v3.
	// The warning will be printed more than once, and it can be annoying, but it should ensure that the user will change these values.
	var requiresChange []string
	if c.OdoSettings.Timeout != nil && *c.OdoSettings.Timeout < minimumDurationValue {
		requiresChange = append(requiresChange, TimeoutSetting)
	}
	if c.OdoSettings.PushTimeout != nil && *c.OdoSettings.PushTimeout < minimumDurationValue {
		requiresChange = append(requiresChange, PushTimeoutSetting)
	}
	if c.OdoSettings.RegistryCacheTime != nil && *c.OdoSettings.RegistryCacheTime < minimumDurationValue {
		requiresChange = append(requiresChange, RegistryCacheTimeSetting)
	}
	if len(requiresChange) != 0 {
		log.Warningf("Please change the preference value for %s, the value does not comply with the minimum value of %s; e.g. of acceptable formats: 4s, 5m, 1h", strings.Join(requiresChange, ", "), minimumDurationValue)
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

// RegistryHandler handles registry add, and remove operations
func (c *preferenceInfo) RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error {
	var registryList []Registry
	var err error
	var registryExist bool

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
				registryList, err = handleWithRegistryExist(index, registryList, operation, registryName, forceFlag)
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
	err = util.WriteToYAMLFile(&c.Preference, c.Filename)
	if err != nil {
		return fmt.Errorf("unable to write the configuration of %q operation to preference file", operation)
	}

	return nil
}

// handleWithoutRegistryExist is useful for performing 'add' operation on registry and ensure that it is only performed if the registry does not already exist
func handleWithoutRegistryExist(registryList []Registry, operation string, registryName string, registryURL string, isSecure bool) ([]Registry, error) {
	switch operation {

	case "add":
		registry := Registry{
			Name:   registryName,
			URL:    registryURL,
			Secure: isSecure,
		}
		registryList = append(registryList, registry)

	case "remove":
		return nil, fmt.Errorf("failed to %v registry: registry %q doesn't exist", operation, registryName)
	}

	return registryList, nil
}

// handleWithRegistryExist is useful for performing 'remove' operation on registry and ensure that it is only performed if the registry exists
func handleWithRegistryExist(index int, registryList []Registry, operation string, registryName string, forceFlag bool) ([]Registry, error) {
	switch operation {

	case "add":
		return nil, fmt.Errorf("failed to %s registry: registry %q already exists", operation, registryName)

	case "remove":
		if !forceFlag {
			if !ui.Proceed(fmt.Sprintf("Are you sure you want to %s registry %q", operation, registryName)) {
				log.Info("Aborted by the user")
				return registryList, nil
			}
		}

		copy(registryList[index:], registryList[index+1:])
		registryList[len(registryList)-1] = Registry{}
		registryList = registryList[:len(registryList)-1]
		log.Info("Successfully removed registry")
	}

	return registryList, nil
}

// SetConfiguration modifies odo preferences in the preference file
// TODO: Use reflect to set parameters
func (c *preferenceInfo) SetConfiguration(parameter string, value string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names
		switch p {

		case "timeout":
			typedval, err := parseDuration(value, parameter)
			if err != nil {
				return err
			}
			c.OdoSettings.Timeout = &typedval

		case "pushtimeout":
			typedval, err := parseDuration(value, parameter)
			if err != nil {
				return err
			}
			c.OdoSettings.PushTimeout = &typedval

		case "registrycachetime":
			typedval, err := parseDuration(value, parameter)
			if err != nil {
				return err
			}
			c.OdoSettings.RegistryCacheTime = &typedval

		case "updatenotification":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return fmt.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.UpdateNotification = &val

		case "ephemeral":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return fmt.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.Ephemeral = &val

		case "consenttelemetry":
			val, err := strconv.ParseBool(strings.ToLower(value))
			if err != nil {
				return fmt.Errorf("unable to set %q to %q, value must be a boolean", parameter, value)
			}
			c.OdoSettings.ConsentTelemetry = &val
		}
	} else {
		return fmt.Errorf("unknown parameter : %q is not a parameter in odo preference, run `odo preference -h` to see list of available parameters", parameter)
	}

	err := util.WriteToYAMLFile(&c.Preference, c.Filename)
	if err != nil {
		return fmt.Errorf("unable to set %q, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md", parameter)
	}
	return nil
}

// parseDuration parses the value set for a parameter;
// if the value is for e.g. "4m", it is parsed by the time pkg and converted to an appropriate time.Duration
// it returns an error if one occurred, or if the parsed value is less than minimumDurationValue
func parseDuration(value, parameter string) (time.Duration, error) {
	typedval, err := time.ParseDuration(value)
	if err != nil {
		return typedval, fmt.Errorf("unable to set %q to %q; cause: %w\n%s", parameter, value, err, NewMinimumDurationValueError().Error())
	}
	if typedval < minimumDurationValue {
		return typedval, fmt.Errorf("unable to set %q to %q; cause: %w", parameter, value, NewMinimumDurationValueError())
	}
	return typedval, nil
}

// DeleteConfiguration deletes odo preference from the odo preference file
func (c *preferenceInfo) DeleteConfiguration(parameter string) error {
	if p, ok := asSupportedParameter(parameter); ok {
		// processing values according to the parameter names

		if err := util.DeleteConfiguration(&c.OdoSettings, p); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown parameter :%q is not a parameter in the odo preference", parameter)
	}

	err := util.WriteToYAMLFile(&c.Preference, c.Filename)
	if err != nil {
		return fmt.Errorf("unable to set %q, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md", parameter)
	}
	return nil
}

// IsSet checks if the value is set in the preference
func (c *preferenceInfo) IsSet(parameter string) bool {
	return util.IsSet(c.OdoSettings, parameter)
}

// GetTimeout returns the value of Timeout from config
// and if absent then returns default
func (c *preferenceInfo) GetTimeout() time.Duration {
	// default timeout value is 1s
	return kpointer.DurationDeref(c.OdoSettings.Timeout, DefaultTimeout)
}

// GetPushTimeout gets the value set by PushTimeout
func (c *preferenceInfo) GetPushTimeout() time.Duration {
	// default timeout value is 240s
	return kpointer.DurationDeref(c.OdoSettings.PushTimeout, DefaultPushTimeout)
}

// GetRegistryCacheTime gets the value set by RegistryCacheTime
func (c *preferenceInfo) GetRegistryCacheTime() time.Duration {
	return kpointer.DurationDeref(c.OdoSettings.RegistryCacheTime, DefaultRegistryCacheTime)
}

// GetUpdateNotification returns the value of UpdateNotification from preferences
// and if absent then returns default
func (c *preferenceInfo) GetUpdateNotification() bool {
	return kpointer.BoolDeref(c.OdoSettings.UpdateNotification, true)
}

// GetEphemeralSourceVolume returns the value of ephemeral from preferences
// and if absent then returns default
func (c *preferenceInfo) GetEphemeralSourceVolume() bool {
	return kpointer.BoolDeref(c.OdoSettings.Ephemeral, DefaultEphemeralSetting)
}

// GetConsentTelemetry returns the value of ConsentTelemetry from preferences
// and if absent then returns default
// default value: false, consent telemetry is disabled by default
func (c *preferenceInfo) GetConsentTelemetry() bool {
	return kpointer.BoolDeref(c.OdoSettings.ConsentTelemetry, DefaultConsentTelemetrySetting)
}

// GetEphemeral returns the value of Ephemeral from preferences
// and if absent then returns default
// default value: true, ephemeral is enabled by default
func (c *preferenceInfo) GetEphemeral() bool {
	return kpointer.BoolDeref(c.OdoSettings.Ephemeral, DefaultEphemeralSetting)
}

func (c *preferenceInfo) UpdateNotification() *bool {
	return c.OdoSettings.UpdateNotification
}

func (c *preferenceInfo) Ephemeral() *bool {
	return c.OdoSettings.Ephemeral
}

func (c *preferenceInfo) Timeout() *time.Duration {
	return c.OdoSettings.Timeout
}

func (c *preferenceInfo) PushTimeout() *time.Duration {
	return c.OdoSettings.PushTimeout
}

func (c *preferenceInfo) RegistryCacheTime() *time.Duration {
	return c.OdoSettings.RegistryCacheTime
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

func (c *preferenceInfo) RegistryNameExists(name string) bool {
	for _, registry := range *c.RegistryList() {
		if registry.Name == name {
			return true
		}
	}
	return false
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
	return dfutil.GetSortedKeys(supportedParameterDescriptions)
}
