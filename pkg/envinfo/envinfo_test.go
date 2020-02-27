package envinfo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/testingutil/filesystem"

	"github.com/openshift/odo/pkg/util"
)

func TestSetEnvInfo(t *testing.T) {

	tempEnvFile, err := ioutil.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testURL := ConfigURL{Name: "testURL", ClusterHost: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}

	tests := []struct {
		name            string
		parameter       string
		value           interface{}
		existingEnvInfo EnvInfo
	}{
		{
			name:      fmt.Sprintf("Case 1: %s to test", URL),
			parameter: URL,
			value:     testURL,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{},
			},
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
			if err != nil {
				t.Error(err)
			}

			isSet := esi.IsSet(tt.parameter)

			if !isSet {
				t.Errorf("the '%v' is not set", tt.parameter)
			}

		})
	}
}

func TestLocalUnsetConfiguration(t *testing.T) {
	tempEnvFile, err := ioutil.TempFile("", "odoenvinfo")
	if err != nil {
		t.Fatal(err)
	}
	defer tempEnvFile.Close()
	os.Setenv(envInfoEnvName, tempEnvFile.Name())
	testURL := ConfigURL{Name: "testURL", ClusterHost: "1.2.3.4.nip.io", TLSSecret: "testTLSSecret"}

	tests := []struct {
		name            string
		parameter       string
		value           interface{}
		existingEnvInfo EnvInfo
	}{
		{
			name:      fmt.Sprintf("Case 1: unset %s", URL),
			parameter: URL,
			value:     testURL,
			existingEnvInfo: EnvInfo{
				componentSettings: ComponentSettings{
					URL: &[]ConfigURL{testURL},
				},
			},
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
			if err != nil {
				t.Error(err)
			}
			isSet := esi.IsSet(tt.parameter)
			if !isSet {
				t.Errorf("the '%v' was not set", tt.parameter)
			}

			err = esi.DeleteEnvSpecificInfo(tt.parameter)

			if err != nil {
				t.Error(err)
			}
			isSet = esi.IsSet(tt.parameter)
			if isSet {
				t.Errorf("the '%v' is not set to nil", tt.parameter)
			}

		})
	}
}

func TestLowerCaseParameterForLocalParameters(t *testing.T) {
	expected := map[string]bool{"url": true}
	actual := util.GetLowerCaseParameters(GetLocallySupportedParameters())
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func TestLocalConfigInitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, err := getEnvInfoFile("")
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

func TestDeleteConfigDirIfEmpty(t *testing.T) {
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
