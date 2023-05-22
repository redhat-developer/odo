package preference

import (
	"time"

	"github.com/redhat-developer/odo/pkg/api"
)

type Client interface {
	IsSet(parameter string) bool
	SetConfiguration(parameter string, value string) error
	DeleteConfiguration(parameter string) error

	GetUpdateNotification() bool
	GetTimeout() time.Duration
	GetPushTimeout() time.Duration
	GetEphemeralSourceVolume() bool
	GetConsentTelemetry() bool
	GetRegistryCacheTime() time.Duration
	GetImageRegistry() string
	RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error

	UpdateNotification() *bool
	Timeout() *time.Duration
	PushTimeout() *time.Duration
	RegistryCacheTime() *time.Duration
	EphemeralSourceVolume() *bool
	ConsentTelemetry() *bool
	RegistryList() []api.Registry
	RegistryNameExists(name string) bool

	NewPreferenceList() api.PreferenceList
}
