package preference

type Client interface {
	IsSet(parameter string) bool
	SetConfiguration(parameter string, value string) error
	DeleteConfiguration(parameter string) error

	GetUpdateNotification() bool
	GetNamePrefix() string
	GetTimeout() int
	GetBuildTimeout() int
	GetPushTimeout() int
	GetEphemeralSourceVolume() bool
	GetConsentTelemetry() bool
	GetRegistryCacheTime() int
	RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error

	UpdateNotification() *bool
	NamePrefix() *string
	Timeout() *int
	BuildTimeout() *int
	PushTimeout() *int
	EphemeralSourceVolume() *bool
	ConsentTelemetry() *bool
	RegistryList() *[]Registry

	NewPreferenceList() PreferenceList
}
