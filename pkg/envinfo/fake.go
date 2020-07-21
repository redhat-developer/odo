package envinfo

// GetFakeEnvInfo gets a fake envInfo using the given componentSettings
func GetFakeEnvInfo(settings ComponentSettings) *EnvInfo {
	return &EnvInfo{
		componentSettings: settings,
	}
}
