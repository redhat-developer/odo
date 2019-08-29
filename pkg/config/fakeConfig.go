package config

func GetOneExistingConfigInfo() LocalConfigInfo {
	componentName := "comp-test-name"
	applicationName := "app-test-name"
	componentType := "nodejs"
	sourceLocation := "github.com/example"

	storageValue := []ComponentStorageSettings{
		{
			Name: "example-storage-0",
		},
		{
			Name: "example-storage-1",
		},
	}

	portsValue := []string{"8080/TCP,45/UDP"}

	urlValue := []ConfigUrl{
		{
			Name: "example-url-0",
		},
		{
			Name: "example-url-1",
		},
	}

	envVars := EnvVarList{
		EnvVar{Name: "env-0", Value: "value-0"},
		EnvVar{Name: "env-1", Value: "value-1"},
	}

	return LocalConfigInfo{
		configFileExists: true,
		LocalConfig: LocalConfig{
			componentSettings: ComponentSettings{
				Name:           &componentName,
				Application:    &applicationName,
				Type:           &componentType,
				SourceLocation: &sourceLocation,
				Storage:        &storageValue,
				Envs:           envVars,
				Ports:          &portsValue,
				Url:            &urlValue,
			},
		},
	}
}

func GetOneNonExistingConfigInfo() LocalConfigInfo {
	return LocalConfigInfo{
		configFileExists: false,
		LocalConfig:      LocalConfig{},
	}
}
