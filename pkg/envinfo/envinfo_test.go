package envinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestSetEnvInfo(t *testing.T) {
	fs := filesystem.NewFakeFs()
	tempEnvFile, err := fs.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testDebugPort := 5005
	invalidParam := "invalidParameter"

	tests := []struct {
		name               string
		parameter          string
		value              interface{}
		existingEnvInfo    EnvInfo
		checkConfigSetting []string
		expectError        bool
	}{
		{
			name:      fmt.Sprintf("Case 1: %s to test", DebugPort),
			parameter: DebugPort,
			value:     strconv.Itoa(testDebugPort),
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{},
			},
			checkConfigSetting: []string{"debugport"},
			expectError:        false,
		},
		{
			name:      fmt.Sprintf("Case 2: %s to test", invalidParam),
			parameter: invalidParam,
			value:     strconv.Itoa(testDebugPort),
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{},
			},
			checkConfigSetting: []string{"debugport"},
			expectError:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi, err := NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			esi.EnvInfo = tt.existingEnvInfo
			err = esi.SetConfiguration(tt.parameter, tt.value)
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for SetConfiguration with %s: %v", tt.parameter, err)
			} else if !tt.expectError && err == nil {
				isSet := false
				for _, configSetting := range tt.checkConfigSetting {
					isSet = esi.IsSet(configSetting)
					if !isSet {
						t.Errorf("the setting '%s' is not set", configSetting)
					}
				}

			}

		})
	}
}

func TestUnsetEnvInfo(t *testing.T) {
	fs := filesystem.NewFakeFs()
	tempEnvFile, err := fs.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testDebugPort := 15005
	invalidParam := "invalidParameter"

	tests := []struct {
		name            string
		parameter       string
		existingEnvInfo EnvInfo
		expectError     bool
	}{
		{
			name:      fmt.Sprintf("Case 1: unset %s", DebugPort),
			parameter: DebugPort,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					DebugPort: &testDebugPort,
				},
			},
			expectError: false,
		},
		{
			name:      fmt.Sprintf("Case 2: unset %s", invalidParam),
			parameter: invalidParam,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					DebugPort: &testDebugPort,
				},
			},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi, err := NewEnvSpecificInfo("")
			if err != nil {
				t.Error(err)
			}
			esi.EnvInfo = tt.existingEnvInfo
			err = esi.DeleteConfiguration(tt.parameter)
			if err == nil && tt.expectError {
				t.Errorf("expected error for DeleteConfiguration with %s", tt.parameter)
			} else if !tt.expectError {
				if err != nil {
					t.Error(err)
				}
				isSet := esi.IsSet(tt.parameter)
				if isSet {
					t.Errorf("the '%v' is not set to nil", tt.parameter)
				}
			}

		})
	}
}

func TestEnvSpecificInfonitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, _, err := getEnvInfoFile("")
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewEnvSpecificInfo("")
	if err != nil {
		t.Errorf("error while creating envinfo %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local env.yaml shouldn't exist yet")
	}
}

func TestDeleteEnvDirIfEmpty(t *testing.T) {
	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a odo config directory on fake fs
	configDir, err := fs.TempDir(os.TempDir(), "odo")
	if err != nil {
		t.Error(err)
	}
	// create a mock env info from above fake fs & dir
	esi, err := mockEnvSpecificInfo(configDir, fs)
	if err != nil {
		t.Error(err)
	}

	envDir := filepath.Join(configDir, ".odo", "env")
	if _, err = fs.Stat(envDir); os.IsNotExist(err) {
		t.Error("config directory doesn't exist")
	}

	tests := []struct {
		name string
		// create indicates if a file is supposed to be created in the odo config dir
		create     bool
		setupEnv   func(create bool, fs filesystem.Filesystem, envDir string) error
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

		err := tt.setupEnv(tt.create, fs, envDir)
		if err != nil {
			t.Error(err)
		}

		err = esi.DeleteEnvDirIfEmpty()
		if err != nil {
			t.Error(err)
		}

		file, err := fs.Stat(envDir)
		if !tt.wantOdoDir && !os.IsNotExist(err) {
			// we don't want odo dir but odo dir exists
			fmt.Println(file.Size())
			t.Error("odo env directory exists even after deleting it")
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

func mockEnvSpecificInfo(configDir string, fs filesystem.Filesystem) (*EnvSpecificInfo, error) {

	esi := &EnvSpecificInfo{
		Filename: filepath.Join(configDir, ".odo", "env", "env.yaml"),
		fs:       fs,
	}
	err := fs.MkdirAll(filepath.Join(configDir, ".odo", "env"), os.ModePerm)
	if err != nil {
		return nil, err
	}

	return esi, nil

}
