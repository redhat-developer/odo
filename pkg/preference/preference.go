package preference

import (
	"time"
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
	RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error

	UpdateNotification() *bool
	Timeout() *time.Duration
	PushTimeout() *time.Duration
	RegistryCacheTime() *time.Duration
	EphemeralSourceVolume() *bool
	ConsentTelemetry() *bool
	RegistryList() *[]Registry
	RegistryNameExists(name string) bool

	NewPreferenceList() PreferenceList
}
