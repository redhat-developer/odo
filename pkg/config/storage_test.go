package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestLocalConfigInfo_StorageCreate(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		storageName    string
		storageSize    string
		storagePath    string
		existingConfig LocalConfig
	}{
		{
			name:        "case 1: no other storage present",
			storageName: "example-storage-0",
			storageSize: "100M",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:        "case 2: one other storage present",
			storageName: "example-storage-1",
			storageSize: "100M",
			storagePath: "/data-1",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data",
							Size: "100M",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			_, err = cfg.StorageCreate(tt.storageName, tt.storageSize, tt.storagePath)
			if err != nil {
				t.Error(err)
			}

			found := false
			for _, storage := range *cfg.componentSettings.Storage {
				if storage.Name == tt.storageName && storage.Size == tt.storageSize && storage.Path == tt.storagePath {
					found = true
				}
			}
			if !found {
				t.Errorf("the storage '%v' is not set properly in the config", tt)
			}
		})
	}
}

func TestLocalConfigInfo_StorageExists(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		storageName    string
		existingConfig LocalConfig
		storageExists  bool
	}{
		{
			name:        "case 1: storage present",
			storageName: "example-storage-1",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-1",
						},
					},
				},
			},
			storageExists: true,
		},
		{
			name:        "case 2: storage present",
			storageName: "example-storage-1",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
						},
					},
				},
			},
			storageExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			exists := cfg.StorageExists(tt.storageName)
			if exists != tt.storageExists {
				t.Errorf("wrong value of exists, expected: %v, unexpected: %v", tt.storageExists, exists)
			}
		})
	}
}

func TestLocalConfigInfo_StorageList(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		existingConfig LocalConfig
	}{
		{
			name: "case 1: one storage exists",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data-0",
							Size: "100M",
						},
					},
				},
			},
		},
		{
			name: "case 2: more than one storage exists",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data-0",
							Size: "100M",
						},
						{
							Name: "example-storage-1",
							Path: "/data-1",
							Size: "100M",
						},
					},
				},
			},
		},
		{
			name: "case 3: no storage exists",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			storageList, err := cfg.StorageList()
			if err != nil {
				t.Error(err)
			}

			if len(*tt.existingConfig.componentSettings.Storage) != len(storageList) {
				t.Errorf("length mismatch, expected: %v, unexpected: %v", len(*tt.existingConfig.componentSettings.Storage), len(storageList))
			}

			for _, storageConfig := range *tt.existingConfig.componentSettings.Storage {
				found := false

				for _, storageResult := range storageList {
					if reflect.DeepEqual(storageResult, storageConfig) {
						found = true
					}
				}

				if !found {
					t.Errorf("storage %v not found while listing", storageConfig)
				}
			}
		})
	}
}

func TestLocalConfigInfo_ValidateStorage(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		storageName    string
		storagePath    string
		existingConfig LocalConfig
		wantError      bool
	}{
		{
			name:        "case 1: no storage present in config",
			storageName: "example-storage-0",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{},
				},
			},
			wantError: false,
		},
		{
			name:        "case 2: storage present in config with no conflict",
			storageName: "example-storage-0",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-1",
							Path: "/data-1",
							Size: "100M",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name:        "case 3: storage present in config and with path conflict",
			storageName: "example-storage-0",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-1",
							Path: "/data",
							Size: "100M",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name:        "case 4: storage present in config and with name conflict",
			storageName: "example-storage-0",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data-1",
							Size: "100M",
						},
					},
				},
			},
			wantError: true,
		},
		{
			name:        "case 5: storage present in config and with name and path conflicts",
			storageName: "example-storage-0",
			storagePath: "/data",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data",
							Size: "100M",
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			err = cfg.ValidateStorage(tt.storageName, tt.storagePath)

			if !tt.wantError && err != nil {
				t.Errorf("no error expected,but got error: %v", err)
			}

			if tt.wantError && err == nil {
				t.Errorf("error expected,but got no error")
			}
		})
	}
}

func TestLocalConfigInfo_StorageDelete(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		storageName    string
		existingConfig LocalConfig
		wantError      bool
	}{
		{
			name:        "case 1: storage does exist",
			storageName: "example-storage-0",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
						},
					},
				},
			},
			wantError: false,
		},
		{
			name:        "case 2: storage doesn't exist",
			storageName: "example-storage-0",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-1",
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			err = cfg.StorageDelete(tt.storageName)

			if !tt.wantError && err != nil {
				t.Errorf("no error expected,but got error: %v", err)
			}

			if tt.wantError && err == nil {
				t.Errorf("error expected,but got no error")
			}

			found := false
			for _, storage := range *cfg.componentSettings.Storage {
				if storage.Name == tt.storageName {
					found = true
				}
			}
			if found {
				t.Errorf("the storage '%v' is not deleted properly from the config", tt.storageName)
			}
		})
	}
}

func TestLocalConfigInfo_GetMountPath(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		storageName    string
		existingConfig LocalConfig
		wantPath       string
	}{
		{
			name:        "case 1: no storage exists",
			storageName: "example-storage-0",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{},
				},
			},
			wantPath: "",
		},
		{
			name:        "case 2: storage exists and one storage exists in config",
			storageName: "example-storage-0",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data",
						},
					},
				},
			},
			wantPath: "/data",
		},
		{
			name:        "case 3: storage exists and two storage exists in config",
			storageName: "example-storage-1",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Storage: &[]ComponentStorageSettings{
						{
							Name: "example-storage-0",
							Path: "/data",
						},
						{
							Name: "example-storage-1",
							Path: "/data-1",
						},
					},
				},
			},
			wantPath: "/data-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo("")
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			path := cfg.GetMountPath(tt.storageName)

			if path != tt.wantPath {
				t.Errorf("the value of returned path is different, expected: %v, got: %v", tt.wantPath, path)
			}
		})
	}
}
