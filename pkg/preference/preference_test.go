package preference

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/openshift/odo/pkg/util"
)

func TestNew(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name    string
		output  *PreferenceInfo
		success bool
	}{
		{
			name: "Test filename is being set",
			output: &PreferenceInfo{
				Filename:   tempConfigFile.Name(),
				Preference: NewPreference(),
			},
			success: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfi, err := NewPreferenceInfo()
			switch test.success {
			case true:
				if err != nil {
					t.Errorf("expected test to pass, but it failed with error: %v", err)
				}
			case false:
				if err == nil {
					t.Errorf("expected test to fail, but it passed!")
				}
			}
			if !reflect.DeepEqual(test.output, cfi) {
				t.Errorf("expected output: %#v", test.output)
				t.Errorf("actual output: %#v", cfi)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	zeroValue := 0
	nonzeroValue := 5
	tests := []struct {
		name           string
		existingConfig Preference
		want           int
	}{
		{
			name:           "Case 1: validating value 1 from config in default case",
			existingConfig: Preference{},
			want:           1,
		},

		{
			name: "Case 2: validating value 0 from config",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					Timeout: &zeroValue,
				},
			},
			want: 0,
		},

		{
			name: "Case 3: validating value 5 from config",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					Timeout: &nonzeroValue,
				},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewPreferenceInfo()
			if err != nil {
				t.Error(err)
			}
			cfg.Preference = tt.existingConfig

			output := cfg.GetTimeout()
			if output != tt.want {
				t.Errorf("GetTimeout returned unexpeced value expected \ngot: %d \nexpected: %d\n", output, tt.want)
			}
		})
	}
}

func TestSetConfiguration(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false
	zeroValue := 0

	tests := []struct {
		name           string
		parameter      string
		value          string
		existingConfig Preference
		wantErr        bool
		want           interface{}
	}{
		// update notification
		{
			name:           fmt.Sprintf("Case 1: %s set nil to true", UpdateNotificationSetting),
			parameter:      UpdateNotificationSetting,
			value:          "true",
			existingConfig: Preference{},
			want:           true,
			wantErr:        false,
		},
		{
			name:      fmt.Sprintf("Case 2: %s set true to false", UpdateNotificationSetting),
			parameter: UpdateNotificationSetting,
			value:     "false",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					UpdateNotification: &trueValue,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name:      fmt.Sprintf("Case 3: %s set false to true", UpdateNotificationSetting),
			parameter: UpdateNotificationSetting,
			value:     "true",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					UpdateNotification: &falseValue,
				},
			},
			want:    true,
			wantErr: false,
		},

		{
			name:           fmt.Sprintf("Case 4: %s invalid value", UpdateNotificationSetting),
			parameter:      UpdateNotificationSetting,
			value:          "invalid_value",
			existingConfig: Preference{},
			wantErr:        true,
		},
		// time out
		{
			name:      fmt.Sprintf("Case 5: %s set to 5 from 0", TimeoutSetting),
			parameter: TimeoutSetting,
			value:     "5",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					Timeout: &zeroValue,
				},
			},
			want:    5,
			wantErr: false,
		},
		{
			name:           fmt.Sprintf("Case 6: %s set to 300", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "300",
			existingConfig: Preference{},
			want:           300,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 7: %s set to 0", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "0",
			existingConfig: Preference{},
			want:           0,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 8: %s set to -1", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "-1",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 9: %s invalid value", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "this",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 10: %s set to 300 with mixed case in parameter name", TimeoutSetting),
			parameter:      "TimeOut",
			value:          "300",
			existingConfig: Preference{},
			want:           300,
			wantErr:        false,
		},
		// invalid parameter
		{
			name:           "Case 11: invalid parameter",
			parameter:      "invalid_parameter",
			existingConfig: Preference{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewPreferenceInfo()
			if err != nil {
				t.Error(err)
			}
			cfg.Preference = tt.existingConfig

			err = cfg.SetConfiguration(tt.parameter, tt.value)

			if !tt.wantErr && err == nil {
				// validating the value after executing Serconfiguration
				// according to component in positive cases
				switch tt.parameter {
				case "updatenotification":
					if *cfg.OdoSettings.UpdateNotification != tt.want {
						t.Errorf("unexpeced value after execution of SetConfiguration \ngot: %t \nexpected: %t\n", *cfg.OdoSettings.UpdateNotification, tt.want)
					}
				case "timeout":
					if *cfg.OdoSettings.Timeout != tt.want {
						t.Errorf("unexpeced value after execution of SetConfiguration \ngot: %v \nexpected: %d\n", cfg.OdoSettings.Timeout, tt.want)
					}
				}
			} else if tt.wantErr && err != nil {
				// negative cases
				switch tt.parameter {
				case "updatenotification":
				case "timeout":
					typedval, err := strconv.Atoi(tt.value)
					// if err is found in cases other than value <0 or !ok
					if !(typedval < 0 || err != nil) {
						t.Error(err)
					}
				}
			} else {
				t.Error(err)
			}

		})
	}
}

func TestGetupdateNotification(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(globalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false

	tests := []struct {
		name           string
		existingConfig Preference
		want           bool
	}{
		{
			name:           fmt.Sprintf("Case 1: %s nil", UpdateNotificationSetting),
			existingConfig: Preference{},
			want:           true,
		},
		{
			name: fmt.Sprintf("Case 2: %s true", UpdateNotificationSetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					UpdateNotification: &trueValue,
				},
			},
			want: true,
		},
		{
			name: fmt.Sprintf("Case 3: %s false", UpdateNotificationSetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					UpdateNotification: &falseValue,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PreferenceInfo{
				Preference: tt.existingConfig,
			}
			output := cfg.GetUpdateNotification()

			if output != tt.want {
				t.Errorf("GetUpdateNotification returned unexpeced value expected \ngot: %t \nexpected: %t\n", output, tt.want)
			}

		})
	}
}

func TestFormatSupportedParameters(t *testing.T) {
	expected := `
Available Parameters:
%s - %s
%s - %s
%s - %s
`
	expected = fmt.Sprintf(expected,
		NamePrefixSetting, NamePrefixSettingDescription,
		TimeoutSetting, TimeoutSettingDescription,
		UpdateNotificationSetting, UpdateNotificationSettingDescription)
	actual := FormatSupportedParameters()
	if expected != actual {
		t.Errorf("expected '%s', got '%s'", expected, actual)
	}
}

func TestLowerCaseParameters(t *testing.T) {
	expected := map[string]bool{"nameprefix": true, "timeout": true, "updatenotification": true}
	actual := util.GetLowerCaseParameters(GetSupportedParameters())
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected '%v', got '%v'", expected, actual)
	}
}

func TestIsSupportedParameter(t *testing.T) {
	tests := []struct {
		testName      string
		param         string
		expectedLower string
		expected      bool
	}{
		{
			testName:      "existing, lower case",
			param:         "timeout",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName:      "existing, from description",
			param:         "Timeout",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName:      "existing, mixed case",
			param:         "TimeOut",
			expectedLower: "timeout",
			expected:      true,
		},
		{
			testName: "empty",
			param:    "",
			expected: false,
		},
		{
			testName: "unexisting",
			param:    "foo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Log("Running test: ", tt.testName)
		t.Run(tt.testName, func(t *testing.T) {
			actual, ok := asSupportedParameter(tt.param)
			if tt.expected != ok && tt.expectedLower != actual {
				t.Fail()
			}
		})
	}
}

func TestMetaTypePopulatedInPreference(t *testing.T) {
	pi, err := NewPreferenceInfo()

	if err != nil {
		t.Error(err)
	}
	if pi.APIVersion != preferenceAPIVersion || pi.Kind != preferenceKind {
		t.Error("the api version and kind in preference are incorrect")
	}
}
