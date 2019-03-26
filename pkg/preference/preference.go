package preference

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/odo/pkg/util"
)

const (
	globalConfigEnvName  = "GLOBALODOCONFIG"
	configFileName       = "config.yaml"
	preferenceKind       = "Preference"
	preferenceAPIVersion = "odo.openshift.io/v1alpha1"

	//DefaultTimeout for openshift server connection check
	DefaultTimeout = 1

	// UpdateNotificationSetting is the name of the setting controlling update notification
	UpdateNotificationSetting = "UpdateNotification"
	// UpdateNotificationSettingDescription is human-readable description for the update notification setting
	UpdateNotificationSettingDescription = "Controls if an update notification is shown or not (true or false)"
	// NamePrefixSetting is the name of the setting controlling name prefix
	NamePrefixSetting = "NamePrefix"
	// NamePrefixSettingDescription is human-readable description for the name prefix setting
	NamePrefixSettingDescription = "Default prefix is the current directory name. Use this value to set a default name prefix"
	// TimeoutSetting is the name of the setting controlling timeout for connection check
	TimeoutSetting = "Timeout"
	// TimeoutSettingDescription is human-readable description for the timeout setting
	TimeoutSettingDescription = "Timeout (in seconds) for OpenShift server connection check"
)

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

var (
	// records information on supported parameters
	supportedParameterDescriptions = map[string]string{
		UpdateNotificationSetting: UpdateNotificationSettingDescription,
		NamePrefixSetting:         NamePrefixSettingDescription,
		TimeoutSetting:            TimeoutSettingDescription,
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
	// Timeout for openshift server connection check
	Timeout *int `yaml:"Timeout,omitempty"`
}

// Preference stores all the preferences related to odo
type Preference struct {
	metav1.TypeMeta `yaml:",inline"`

	// Odo settings holds the odo specific global settings
	OdoSettings OdoSettings `yaml:"OdoSettings,omitempty"`
}

func getPreferenceFile() (string, error) {
	if env, ok := os.LookupEnv(globalConfigEnvName); ok {
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
	configFile, err := getPreferenceFile()
	glog.V(4).Infof("The configFile is %+v", configFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo config file")
	}
	// Check whether directory and file are not present if they aren't then create them
	if err = util.CreateIfNotExists(configFile); err != nil {
		return nil, err
	}
	c := PreferenceInfo{
		Preference: NewPreference(),
	}
	c.Filename = configFile
	err = util.GetFromFile(&c.Preference, c.Filename)
	if err != nil {
		return nil, err
	}
	return &c, nil
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
	if c.OdoSettings.Timeout == nil {
		return DefaultTimeout
	}
	return *c.OdoSettings.Timeout
}

// GetUpdateNotification returns the value of UpdateNotification from config
// and if absent then returns default
func (c *PreferenceInfo) GetUpdateNotification() bool {
	if c.OdoSettings.UpdateNotification == nil {
		return true
	}
	return *c.OdoSettings.UpdateNotification
}

// GetNamePrefix returns the value of Prefix from config
// and if absent then returns default
func (c *PreferenceInfo) GetNamePrefix() string {
	if c.OdoSettings.NamePrefix == nil {
		return ""
	}
	return *c.OdoSettings.NamePrefix
}

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
