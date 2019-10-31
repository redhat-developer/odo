package config

func GetOneExistingConfigInfo(componentName, applicationName, projectName string) LocalConfigInfo {
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

	urlValue := []ConfigURL{
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
				URL:            &urlValue,
				Project:        &projectName,
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
