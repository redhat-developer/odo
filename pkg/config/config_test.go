package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestSetLocalConfiguration(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())
	minCPUValue := "0.5"
	maxCPUValue := "2"
	minMemValue := "500M"
	maxMemValue := "1000M"
	testValue := "test"
	portsValue := "8080/TCP,45/UDP"
	typeValue := "nodejs"
	applicationValue := "odotestapp"
	projectValue := "odotestproject"
	sourceTypeValue := "git"
	sourceLocationValue := "https://github.com/sclorg/nodejs-ex"
	refValue := "develop"

	tests := []struct {
		name           string
		parameter      string
		value          string
		existingConfig LocalConfig
	}{
		// update notification
		{
			name:      fmt.Sprintf("Case 1: %s set nil to true", Ignore),
			parameter: Ignore,
			value:     "true",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 2: %s set true to false", Ignore),
			parameter: Ignore,
			value:     "false",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 3: %s to test", Name),
			parameter: Name,
			value:     testValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 5: %s set to %s from 0", MaxCPU, maxCPUValue),
			parameter: MaxCPU,
			value:     maxCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 6: %s set to %s", MinCPU, minCPUValue),
			parameter: MinCPU,
			value:     minCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 6: %s set to %s", MinMemory, minMemValue),
			parameter: MinMemory,
			value:     minMemValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 7: %s set to %s", MaxMemory, maxCPUValue),
			parameter: MaxMemory,
			value:     maxMemValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 8: %s set to %s", Ports, portsValue),
			parameter: Ports,
			value:     portsValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 9: %s set to %s", Type, typeValue),
			parameter: Type,
			value:     typeValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 10: %s set to %s", Application, applicationValue),
			parameter: Application,
			value:     applicationValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 11: %s set to %s", Project, projectValue),
			parameter: Project,
			value:     projectValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 12: %s set to %s", SourceType, sourceTypeValue),
			parameter: SourceType,
			value:     sourceTypeValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 12: %s set to %s", SourceLocation, sourceLocationValue),
			parameter: SourceLocation,
			value:     sourceLocationValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
			},
		},
		{
			name:      fmt.Sprintf("Case 13: %s set to %s", Ref, refValue),
			parameter: Ref,
			value:     refValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{},
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

			err = cfg.SetConfiguration(tt.parameter, tt.value)
			if err != nil {
				t.Error(err)
			}

			isSet := cfg.IsSet(tt.parameter)

			if !isSet {
				t.Errorf("the '%v' is not set", tt.parameter)
			}

		})
	}
}

func TestLocalUnsetConfiguration(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())
	trueValue := true
	minCPUValue := "0.5"
	maxCPUValue := "2"
	minMemValue := "500M"
	testValue := "test"

	tests := []struct {
		name           string
		parameter      string
		value          string
		existingConfig LocalConfig
	}{
		// update notification
		{
			name:      fmt.Sprintf("Case 1: unset %s", Ignore),
			parameter: Ignore,
			value:     "true",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Ignore: &trueValue,
				},
			},
		},
		{
			name:      fmt.Sprintf("Case 3: unset %s", Name),
			parameter: Name,
			value:     testValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Name: &testValue,
				},
			},
		},
		{
			name:      fmt.Sprintf("Case 5: unset %s", MaxCPU),
			parameter: MaxCPU,
			value:     maxCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MaxCPU: &maxCPUValue,
				},
			},
		},
		{
			name:      fmt.Sprintf("Case 6: unset %s", MinCPU),
			parameter: MinCPU,
			value:     minCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MinCPU: &minCPUValue,
				},
			},
		},
		{
			name:      fmt.Sprintf("Case 6: unset %s", MinMemory),
			parameter: MinMemory,
			value:     minMemValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MinMemory: &minMemValue,
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

			err = cfg.SetConfiguration(tt.parameter, tt.value)
			if err != nil {
				t.Error(err)
			}
			isSet := cfg.IsSet(tt.parameter)
			if !isSet {
				t.Errorf("the '%v' was not set", tt.parameter)
			}

			err = cfg.DeleteConfiguration(tt.parameter)

			if err != nil {
				t.Error(err)
			}
			isSet = cfg.IsSet(tt.parameter)
			if isSet {
				t.Errorf("the '%v' is not set to nil", tt.parameter)
			}

		})
	}
}

func TestLocalConfigInitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, err := getLocalConfigFile("")
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewLocalConfigInfo("")
	if err != nil {
		t.Errorf("error while creating local config %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local config.yaml shouldn't exist yet")
	}
}

func TestMetaTypePopulatedInLocalConfig(t *testing.T) {
	ci, err := NewLocalConfigInfo("")

	if err != nil {
		t.Error(err)
	}
	if ci.typeMeta.APIVersion != localConfigAPIVersion || ci.typeMeta.Kind != localConfigKind {
		t.Error("the api version and kind in local config are incorrect")
	}
}

// TODO: Write Windows tests for below
func TestGetOSSourcePath(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())

	binarySourceType := BINARY
	localSourceType := LOCAL
	gitSourceType := GIT

	tests := []struct {
		name           string
		parameter      string
		value          string
		wantErr        bool
		existingConfig LocalConfig
	}{
		{
			name:      "Case 1: Valid location (even though it shows c:/)",
			parameter: SourceLocation,
			value:     "file://c:/foo/bar",
			wantErr:   false,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &binarySourceType,
				},
			},
		},
		{
			name:      "Case 2: Error if passing in blank",
			parameter: SourceLocation,
			value:     "",
			wantErr:   true,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &localSourceType,
				},
			},
		},
		{
			name:      "Case 3: Error if we're passing in git source type...",
			parameter: SourceLocation,
			value:     "",
			wantErr:   true,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &gitSourceType,
				},
			},
		},
		{
			name:      "Case 4: Error if passing in just a url but using local",
			parameter: SourceLocation,
			value:     "https://redhat.com",
			wantErr:   true,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &localSourceType,
				},
			},
		},
		{
			name:      "Case 5: Valid path",
			parameter: SourceLocation,
			value:     "/var/foo/bar",
			wantErr:   false,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &localSourceType,
				},
			},
		},
		{
			name:      "Case 6: Error if URL escapes were passed in..",
			parameter: SourceLocation,
			value:     "%a",
			wantErr:   true,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &localSourceType,
				},
			},
		},
		{
			name:      "Case 7: Valid binary path",
			parameter: SourceLocation,
			value:     "/var/foo/bar",
			wantErr:   false,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					SourceType: &binarySourceType,
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

			err = cfg.SetConfiguration(tt.parameter, tt.value)
			if err != nil {
				t.Error(err)
			}
			isSet := cfg.IsSet(tt.parameter)
			if !isSet {
				t.Errorf("the '%v' was not set", tt.parameter)
			}

			_, err = cfg.GetOSSourcePath()
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %s source path", tt.value)
			} else if !tt.wantErr && err != nil {
				t.Error(err)
			}

		})
	}
}

func TestDeleteConfigDirIfEmpty(t *testing.T) {
	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a odo config directory on fake fs
	configDir, err := fs.TempDir(os.TempDir(), "odo")
	if err != nil {
		t.Error(err)
	}
	// create a mock local configuration from above fake fs & dir
	lci, err := mockLocalConfigInfo(configDir, fs)
	if err != nil {
		t.Error(err)
	}

	odoDir := filepath.Join(configDir, ".odo")
	if _, err = fs.Stat(odoDir); os.IsNotExist(err) {
		t.Error("config directory doesn't exist")
	}

	tests := []struct {
		name string
		// create indicates if a file is supposed to be created in the odo config dir
		create     bool
		setupEnv   func(create bool, fs filesystem.Filesystem, odoDir string) error
		wantOdoDir bool
		wantErr    bool
	}{
		{
			name:       "Case 1: Empty config dir",
			create:     false,
			setupEnv:   createDirectoryAndFile,
			wantOdoDir: false,
		},
		{
			name:       "Case 2: Config dir with test file",
			create:     true,
			setupEnv:   createDirectoryAndFile,
			wantOdoDir: true,
		},
	}

	for _, tt := range tests {

		err := tt.setupEnv(tt.create, fs, odoDir)
		if err != nil {
			t.Error(err)
		}

		err = lci.DeleteConfigDirIfEmpty()
		if err != nil {
			t.Error(err)
		}

		file, err := fs.Stat(odoDir)
		if !tt.wantOdoDir && !os.IsNotExist(err) {
			// we don't want odo dir but odo dir exists
			fmt.Println(file.Size())
			t.Error("odo config directory exists even after deleting it")
			t.Errorf("Error in test %q", tt.name)
		} else if tt.wantOdoDir && os.IsNotExist(err) {
			// we want odo dir to exist after odo delete --all but it does not exist
			t.Error("wanted odo directory to exist after odo delete --all")
			t.Errorf("Error in test %q", tt.name)
		}
	}
}

func createDirectoryAndFile(create bool, fs filesystem.Filesystem, odoDir string) error {
	if !create {
		return nil
	}

	file, err := fs.Create(filepath.Join(odoDir, "testfile"))
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("hello world"))
	if err != nil {
		return err
	}

	file.Close()
	if err != nil {
		return err
	}
	return nil
}
