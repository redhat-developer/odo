package preference

type Client interface {
	IsSet(parameter string) bool
	SetConfiguration(parameter string, value string) error
	DeleteConfiguration(parameter string) error

	GetUpdateNotification() bool
	GetTimeout() int
	GetPushTimeout() int
	GetEphemeralSourceVolume() bool
	GetConsentTelemetry() bool
	GetRegistryCacheTime() int
	RegistryHandler(operation string, registryName string, registryURL string, forceFlag bool, isSecure bool) error

	UpdateNotification() *bool
	Timeout() *int
	PushTimeout() *int
	RegistryCacheTime() *int
	EphemeralSourceVolume() *bool
	ConsentTelemetry() *bool
	RegistryList() *[]Registry
	RegistryNameExists(name string) bool

	NewPreferenceList() PreferenceList
}
