package preference

import (
	"fmt"
	"os"

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
