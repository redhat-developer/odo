package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/redhat-developer/odo/pkg/util"
)

func TestSetLocalConfiguration(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(localConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false
	minCPUValue := "0.5"
	maxCPUValue := "2"
	minMemValue := "500M"
	testValue := "test"

	tests := []struct {
		name           string
		parameter      string
		value          string
		existingConfig LocalConfig
		want           interface{}
	}{
		// update notification
		{
			name:      fmt.Sprintf("Case 1: %s set nil to true", Ignore),
			parameter: Ignore,
			value:     "true",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Ignore: &trueValue,
				},
			},
			want: true,
		},
		{
			name:      fmt.Sprintf("Case 2: %s set true to false", Ignore),
			parameter: Ignore,
			value:     "false",
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					Ignore: &falseValue,
				},
			},
			want: false,
		},
		{
			name:      fmt.Sprintf("Case 3: %s to test", ComponentName),
			parameter: ComponentName,
			value:     testValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					ComponentName: &testValue,
				},
			},
			want: testValue,
		},
		{
			name:      fmt.Sprintf("Case 5: %s set to %s from 0", MaxCPU, maxCPUValue),
			parameter: MaxCPU,
			value:     maxCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MaxCPU: &maxCPUValue,
				},
			},
			want: maxCPUValue,
		},
		{
			name:      fmt.Sprintf("Case 6: %s set to %s", MinCPU, minCPUValue),
			parameter: MinCPU,
			value:     minCPUValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MinCPU: &minCPUValue,
				},
			},
			want: minCPUValue,
		},
		{
			name:      fmt.Sprintf("Case 6: %s set to %s", MinMemory, minMemValue),
			parameter: MinMemory,
			value:     minMemValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					MinMemory: &minMemValue,
				},
			},
			want: minMemValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewLocalConfigInfo()
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			err = cfg.SetConfiguration(tt.parameter, tt.value)
			if err != nil {
				t.Error(err)
			}

			idata, _ := cfg.GetConfiguration(tt.parameter)

			if idata != tt.want {
				t.Errorf("the '%v' is not set to '%v' instead its '%v'", tt.parameter, tt.want, idata)
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
			name:      fmt.Sprintf("Case 3: unset %s", ComponentName),
			parameter: ComponentName,
			value:     testValue,
			existingConfig: LocalConfig{
				componentSettings: ComponentSettings{
					ComponentName: &testValue,
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
			cfg, err := NewLocalConfigInfo()
			if err != nil {
				t.Error(err)
			}
			cfg.LocalConfig = tt.existingConfig

			err = cfg.SetConfiguration(tt.parameter, tt.value)
			if err != nil {
				t.Error(err)
			}
			idata, ok := cfg.GetConfiguration(tt.parameter)
			if (idata == nil) || !ok {
				t.Errorf("the '%v' was not set", tt.parameter)
			}

			err = cfg.DeleteConfiguration(tt.parameter)

			if err != nil {
				t.Error(err)
			}
			idata, ok = cfg.GetConfiguration(tt.parameter)
			if (idata != nil) || !ok {
				t.Errorf("the '%v' is not set to nil", tt.parameter)
			}

		})
	}
}

func TestLowerCaseParameterForLocalParameters(t *testing.T) {
	expected := map[string]bool{"componentname": true, "minmemory": true, "ignore": true, "project": true,
		"app": true, "componenttype": true, "ref": true, "mincpu": true, "cpu": true, "ports": true, "maxmemory": true,
		"maxcpu": true, "type": true, "path": true, "memory": true}
	actual := util.GetLowerCaseParameters(GetLocallySupportedParameters())
	if !boolMapEqual(expected, actual) {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func boolMapEqual(m1, m2 map[string]bool) bool {
	for key, value := range m1 {
		if val2, ok := m2[key]; ok {
			if val2 != value {
				return false
			}
		}
	}
	return len(m1) == len(m2)
}

func TestLocalConfigInitDoesntCreateLocalOdoFolder(t *testing.T) {
	// cleaning up old odo files if any
	filename, err := getLocalConfigFile()
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewLocalConfigInfo()
	if err != nil {
		t.Errorf("error while creating local config %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("local config.yaml shouldn't exist yet")
	}
}

func TestMetaTypePopulatedInLocalConfig(t *testing.T) {
	ci, err := NewLocalConfigInfo()

	if err != nil {
		t.Error(err)
	}
	if ci.APIVersion != localConfigAPIVersion || ci.Kind != localConfigKind {
		t.Error("the api version and kind in local config are incorrect")
	}
}
