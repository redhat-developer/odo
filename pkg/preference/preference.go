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

	"github.com/openshift/odo/pkg/util"
)

const (
	GlobalConfigEnvName  = "GLOBALODOCONFIG"
	configFileName       = "preference.yaml"
	preferenceKind       = "Preference"
	preferenceAPIVersion = "odo.openshift.io/v1alpha1"

	//DefaultTimeout for openshift server connection check (in seconds)
	DefaultTimeout = 1

	// DefaultPushTimeout is the default timeout for pods (in seconds)
	DefaultPushTimeout = 240

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

	// Constants for PushTarget values

	// DockerPushTarget represents the value of the push target when it's set to Docker
	DockerPushTarget = "docker"

	// KubePushTarget represents the value of the push target when it's set to Kube
	KubePushTarget = "kube"
)

// TimeoutSettingDescription is human-readable description for the timeout setting
var TimeoutSettingDescription = fmt.Sprintf("Timeout (in seconds) for OpenShift server connection check (Default: %d)", DefaultTimeout)

// PushTimeoutSettingDescription adds a description for PushTimeout
var PushTimeoutSettingDescription = fmt.Sprintf("PushTimeout (in seconds) for waiting for a Pod to come up (Default: %d)", DefaultPushTimeout)

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

var (
	// records information on supported parameters
	supportedParameterDescriptions = map[string]string{
		UpdateNotificationSetting: UpdateNotificationSettingDescription,
		NamePrefixSetting:         NamePrefixSettingDescription,
		TimeoutSetting:            TimeoutSettingDescription,
		PushTimeoutSetting:        PushTimeoutSettingDescription,
		ExperimentalSetting:       ExperimentalDescription,
		PushTargetSetting:         PushTargetDescription,
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

	// PushTimeout for OpenShift pod timeout check
	PushTimeout *int `yaml:"PushTimeout,omitempty"`

	// Experimental for exposing features in development/experimental mode
	Experimental *bool `yaml:"Experimental,omitempty"`

	// PushTarget for telling odo which platform to push to (either kube or docker)
	PushTarget *string `yaml:"PushTarget,omitempty"`
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
	// if the preference file doesn't exist then we dont worry about it and return
	if _, err = os.Stat(preferenceFile); os.IsNotExist(err) {
		return &c, nil
	}

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

		case "pushtimeout":
			typedval, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrapf(err, "unable to set %s to %s", parameter, value)
			}
			if typedval < 0 {
				return errors.Errorf("cannot set timeout to less than 0")
			}
			c.OdoSettings.PushTimeout = &typedval

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
	if c.OdoSettings.Timeout == nil {
		return DefaultTimeout
	}
	return *c.OdoSettings.Timeout
}

// GetPushTimeout gets the value set by PushTimeout
func (c *PreferenceInfo) GetPushTimeout() int {
	// default timeout value is 1
	if c.OdoSettings.PushTimeout == nil {
		return DefaultPushTimeout
	}
	return *c.OdoSettings.PushTimeout
}

// GetUpdateNotification returns the value of UpdateNotification from preferences
// and if absent then returns default
func (c *PreferenceInfo) GetUpdateNotification() bool {
	if c.OdoSettings.UpdateNotification == nil {
		return true
	}
	return *c.OdoSettings.UpdateNotification
}

// GetNamePrefix returns the value of Prefix from preferences
// and if absent then returns default
func (c *PreferenceInfo) GetNamePrefix() string {
	if c.OdoSettings.NamePrefix == nil {
		return ""
	}
	return *c.OdoSettings.NamePrefix
}

// GetExperimental returns the value of Experimental from preferences
// and if absent then returns default
// default value: false, experimental mode is disabled by default
func (c *PreferenceInfo) GetExperimental() bool {
	if c.OdoSettings.Experimental == nil {
		return false
	}
	return *c.OdoSettings.Experimental
}

// GetPushTarget returns the value of PushTarget from preferences
// and if absent then returns defualt
// default value: kube, docker push target needs to be manually enabled
func (c *PreferenceInfo) GetPushTarget() string {
	if c.OdoSettings.PushTarget == nil {
		return KubePushTarget
	}
	return *c.OdoSettings.PushTarget
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
