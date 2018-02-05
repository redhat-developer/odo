package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGetOcDevConfigFile(t *testing.T) {
	// TODO: implement this
}

func TestNew(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "ocdevconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(configEnvName, tempConfigFile.Name())

	tests := []struct {
		name    string
		output  *ConfigInfo
		success bool
	}{
		{
			name: "Test filename is being set",
			output: &ConfigInfo{
				Filename: tempConfigFile.Name(),
			},
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfi, err := New()
			switch test.success {
			case true:
				if err != nil {
					t.Errorf("Expected test to pass, but it failed with error: %v", err)
				}
			case false:
				if err == nil {
					t.Errorf("Expected test to fail, but it passed!")
				}
			}
			if !reflect.DeepEqual(test.output, cfi) {
				t.Errorf("Expected output: %#v", test.output)
				t.Errorf("Actual output: %#v", cfi)
			}
		})
	}
}

func TestSetActiveComponent(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "ocdevconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(configEnvName, tempConfigFile.Name())

	tests := []struct {
		name           string
		existingConfig Config
		setComponent   string
		project        string
	}{
		{
			name:           "activeComponents nil",
			existingConfig: Config{},
			setComponent:   "foo",
			project:        "bar",
		},
		{
			name: "activeComponents empty",
			existingConfig: Config{
				ActiveComponents: make(map[string]string),
			},
			setComponent: "foo",
			project:      "bar",
		},
		{
			name: "activeComponents existing",
			existingConfig: Config{
				ActiveComponents: map[string]string{
					"a": "b",
				},
			},
			setComponent: "foo",
			project:      "bar",
		},
		{
			name: "overwrite existing active component",
			existingConfig: Config{
				ActiveComponents: map[string]string{
					"foo": "foo",
				},
			},
			setComponent: "foo",
			project:      "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New()
			if err != nil {
				t.Error(err)
			}
			err = cfg.SetActiveComponent(tt.setComponent, tt.project)
			if err != nil {
				t.Error(err)
			}

			found := false
			for aproj, acomp := range cfg.ActiveComponents {
				if aproj == tt.project && acomp == tt.setComponent {
					found = true
				}
			}
			if !found {
				t.Errorf("component %s/%s was not set as current", tt.project, tt.setComponent)
			}

		})
	}
}

func TestGetActiveComponent(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "ocdevconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(configEnvName, tempConfigFile.Name())

	tests := []struct {
		name            string
		existingConfig  Config
		activeProject   string
		activeComponent string
	}{
		{
			name:            "no component active",
			existingConfig:  Config{},
			activeProject:   "test",
			activeComponent: "",
		},
		{
			name: "activeComponents empty",
			existingConfig: Config{
				ActiveComponents: make(map[string]string),
			},
			activeProject:   "test",
			activeComponent: "",
		},
		{
			name: "no activeComponet record for given project",
			existingConfig: Config{
				ActiveComponents: map[string]string{
					"a": "b",
				},
			},
			activeProject:   "test",
			activeComponent: "",
		},
		{
			name: "activeComponents for one project",
			existingConfig: Config{
				ActiveComponents: map[string]string{
					"a": "b",
				},
			},
			activeProject:   "a",
			activeComponent: "b",
		},
		{
			name: "multiple projects",
			existingConfig: Config{
				ActiveComponents: map[string]string{
					"foo": "foo",
					"a":   "b",
				},
			},
			activeProject:   "a",
			activeComponent: "b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ConfigInfo{
				Config: tt.existingConfig,
			}
			output := cfg.GetActiveComponent(tt.activeProject)

			if output != tt.activeComponent {
				t.Errorf("active component doesn't match expected \ngot: %s \nexpected: %s\n", output, tt.activeComponent)
			}

		})
	}
}

//
//func TestGet(t *testing.T) {
//
//}
//
//func TestSet(t *testing.T) {
//
//}
//
//func TestApplicationExists(t *testing.T) {
//
//}
//
//func TestAddApplication(t *testing.T) {
//
//}
