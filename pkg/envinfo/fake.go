package envinfo

func GetFakeEnvInfo(settings ComponentSettings) *EnvInfo {
	return &EnvInfo{
		componentSettings: settings,
	}
}
