package preference

import (
	"fmt"
	"os"
	"time"

	"github.com/redhat-developer/odo/pkg/util"
)

const (
	GlobalConfigEnvName  = "GLOBALODOCONFIG"
	configFileName       = "preference.yaml"
	preferenceKind       = "Preference"
	preferenceAPIVersion = "odo.dev/v1alpha1"

	// minimumDurationValue is the minimum acceptable value for preferences that accept time.Duration
	minimumDurationValue = 1 * time.Second

	// DefaultTimeout for cluster server connection check (in seconds)
	DefaultTimeout = 1 * time.Second

	// DefaultPushTimeout is the default timeout for pods (in seconds)
	DefaultPushTimeout = 240 * time.Second

	// UpdateNotificationSetting is the name of the setting controlling update notification
	UpdateNotificationSetting = "UpdateNotification"

	// UpdateNotificationSettingDescription is human-readable description for the update notification setting
	UpdateNotificationSettingDescription = "Flag to control if an update notification is shown or not (Default: true)"

	// TimeoutSetting is the name of the setting controlling timeout for connection check
	TimeoutSetting = "Timeout"

	// PushTimeoutSetting is the name of the setting controlling PushTimeout
	PushTimeoutSetting = "PushTimeout"

	// RegistryCacheTimeSetting is human-readable description for the registrycachetime setting
	RegistryCacheTimeSetting = "RegistryCacheTime"

	// ImageRegistrySetting is the name of the setting controlling ImageRegistry
	ImageRegistrySetting = "ImageRegistry"

	// DefaultDevfileRegistryName is the name of default devfile registry
	DefaultDevfileRegistryName = "DefaultDevfileRegistry"

	// DefaultDevfileRegistryURL is the URL of default devfile registry
	DefaultDevfileRegistryURL = "https://registry.devfile.io"

	// OldDefaultDevfileRegistryURL is the URL of old default devfile registry for registry migration purpose
	OldDefaultDevfileRegistryURL = "https://github.com/odo-devfiles/registry"

	// DefaultRegistryCacheTime is time (in minutes) for how long odo will cache information from Devfile registry
	DefaultRegistryCacheTime = 15 * time.Minute

	// EphemeralSetting specifies if ephemeral volumes needs to be used as source volume.
	EphemeralSetting = "Ephemeral"

	// DefaultEphemeralSetting is a default value for Ephemeral preference
	DefaultEphemeralSetting = false

	// ConsentTelemetrySettings specifies if the user consents to telemetry
	ConsentTelemetrySetting = "ConsentTelemetry"

	// DefaultConsentTelemetry is a default value for ConsentTelemetry preference
	DefaultConsentTelemetrySetting = false
)

// TimeoutSettingDescription is human-readable description for the timeout setting
var TimeoutSettingDescription = fmt.Sprintf("Timeout (in Duration) for cluster server connection check (Default: %s)", DefaultTimeout)

// PushTimeoutSettingDescription adds a description for PushTimeout
var PushTimeoutSettingDescription = fmt.Sprintf("PushTimeout (in Duration) for waiting for a Pod to come up (Default: %s)", DefaultPushTimeout)

// RegistryCacheTimeSettingDescription adds a description for RegistryCacheTime
var RegistryCacheTimeSettingDescription = fmt.Sprintf("For how long (in Duration) odo will cache information from the Devfile registry (Default: %s)", DefaultRegistryCacheTime)

// EphemeralSettingDescription adds a description for EphemeralSourceVolume
var EphemeralSettingDescription = fmt.Sprintf("If true, odo will create an emptyDir volume to store source code (Default: %t)", DefaultEphemeralSetting)

// ConsentTelemetrySettingDescription adds a description for TelemetryConsentSetting
var ConsentTelemetrySettingDescription = fmt.Sprintf("If true, odo will collect telemetry for the user's odo usage (Default: %t)\n\t\t    For more information: https://developers.redhat.com/article/tool-data-collection", DefaultConsentTelemetrySetting)

const ImageRegistrySettingDescription = "Image Registry to which relative image names in Devfile Image Components will be pushed to (Example: quay.io/my-user/)"

// This value can be provided to set a seperate directory for users 'homedir' resolution
// note for mocking purpose ONLY
var customHomeDir = os.Getenv("CUSTOM_HOMEDIR")

var (
	// records information on supported parameters
	supportedParameterDescriptions = map[string]string{
		UpdateNotificationSetting: UpdateNotificationSettingDescription,
		TimeoutSetting:            TimeoutSettingDescription,
		PushTimeoutSetting:        PushTimeoutSettingDescription,
		RegistryCacheTimeSetting:  RegistryCacheTimeSettingDescription,
		EphemeralSetting:          EphemeralSettingDescription,
		ConsentTelemetrySetting:   ConsentTelemetrySettingDescription,
		ImageRegistrySetting:      ImageRegistrySettingDescription,
	}

	// set-like map to quickly check if a parameter is supported
	lowerCaseParameters = util.GetLowerCaseParameters(GetSupportedParameters())
)
