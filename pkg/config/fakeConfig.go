package config

func GetOneExistingConfigInfo(componentName, applicationName, projectName string) LocalConfigInfo {
	componentType := "nodejs"
	sourceLocation := "./"

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

	localVar := LOCAL

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
				SourceType:     &localVar,
			},
		},
	}
}

func GetOneExistingConfigInfoStorage(componentName, applicationName, projectName, storeName, storeSize, storePath string) LocalConfigInfo {
	componentType := "nodejs"
	sourceLocation := "./"

	storageValue := []ComponentStorageSettings{
		{
			Name: storeName,
			Size: storeSize,
			Path: storePath,
		},
	}

	localVar := LOCAL

	return LocalConfigInfo{
		configFileExists: true,
		LocalConfig: LocalConfig{
			componentSettings: ComponentSettings{
				Name:           &componentName,
				Application:    &applicationName,
				Type:           &componentType,
				SourceLocation: &sourceLocation,
				Storage:        &storageValue,
				Project:        &projectName,
				SourceType:     &localVar,
			},
		},
	}
}

func GetOneGitExistingConfigInfo(componentName, applicationName, projectName string) LocalConfigInfo {
	localConfigInfo := GetOneExistingConfigInfo(componentName, applicationName, projectName)
	git := GIT
	location := "https://example.com"
	localConfigInfo.LocalConfig.componentSettings.SourceType = &git
	localConfigInfo.LocalConfig.componentSettings.SourceLocation = &location
	return localConfigInfo
}

func GetOneNonExistingConfigInfo() LocalConfigInfo {
	return LocalConfigInfo{
		configFileExists: false,
		LocalConfig:      LocalConfig{},
	}
}
