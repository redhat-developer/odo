package preference

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNew(t *testing.T) {

	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())

	tests := []struct {
		name    string
		output  *PreferenceInfo
		success bool
	}{
		{
			name: "Test filename is being set",
			output: &PreferenceInfo{
				Filename: tempConfigFile.Name(),
				Preference: Preference{
					TypeMeta: metav1.TypeMeta{
						Kind:       preferenceKind,
						APIVersion: preferenceAPIVersion,
					},
					OdoSettings: OdoSettings{
						RegistryList: &[]Registry{
							{
								Name:   DefaultDevfileRegistryName,
								URL:    DefaultDevfileRegistryURL,
								Secure: false,
							},
						},
					},
				},
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

func TestGetBuildTimeout(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
	zeroValue := 0
	nonzeroValue := 5
	tests := []struct {
		name           string
		existingConfig Preference
		want           int
	}{
		{
			name:           "Case 1: Validating default value from test case",
			existingConfig: Preference{},
			want:           300,
		},

		{
			name: "Case 2: Validating value 0 from configuration",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					BuildTimeout: &zeroValue,
				},
			},
			want: 0,
		},

		{
			name: "Case 3: Validating value 5 from configuration",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					BuildTimeout: &nonzeroValue,
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

			output := cfg.GetBuildTimeout()
			if output != tt.want {
				t.Errorf("GetBuildTimeout returned unexpected value\ngot: %d \nexpected: %d\n", output, tt.want)
			}
		})
	}
}

func TestGetPushTimeout(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
	zeroValue := 0
	nonzeroValue := 5
	tests := []struct {
		name           string
		existingConfig Preference
		want           int
	}{
		{
			name:           "Case 1: Validating default value from test case",
			existingConfig: Preference{},
			want:           240,
		},

		{
			name: "Case 2: Validating value 0 from configuration",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					PushTimeout: &zeroValue,
				},
			},
			want: 0,
		},

		{
			name: "Case 3: Validating value 5 from configuration",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					PushTimeout: &nonzeroValue,
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

			output := cfg.GetPushTimeout()
			if output != tt.want {
				t.Errorf("GetPushTimeout returned unexpected value\ngot: %d \nexpected: %d\n", output, tt.want)
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
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
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
				t.Errorf("GetTimeout returned unexpected value\ngot: %d \nexpected: %d\n", output, tt.want)
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
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
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
		{
			name:           fmt.Sprintf("Case 12: %s set to 50 with mixed case in parameter name", TimeoutSetting),
			parameter:      "BuildTimeout",
			value:          "50",
			existingConfig: Preference{},
			want:           50,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 13: %s set to 0", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "0",
			existingConfig: Preference{},
			want:           0,
			wantErr:        false,
		},
		{
			name:           fmt.Sprintf("Case 14: %s set to -1 with mixed case in parameter name", TimeoutSetting),
			parameter:      "BuildTimeout",
			value:          "-1",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 15: %s invalid value", TimeoutSetting),
			parameter:      TimeoutSetting,
			value:          "invalid",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 16: %s set to 99 with mixed case in parameter name", TimeoutSetting),
			parameter:      "PushTimeout",
			value:          "99",
			existingConfig: Preference{},
			want:           99,
			wantErr:        false,
		},
		{
			name:           "Case 17: set RegistryCacheTime to non int value",
			parameter:      "RegistryCacheTime",
			value:          "a",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 18: set %s to non bool value", ConsentTelemetrySetting),
			parameter:      ConsentTelemetrySetting,
			value:          "123",
			existingConfig: Preference{},
			wantErr:        true,
		},
		{
			name:           fmt.Sprintf("Case 19: set %s from nil to true", ConsentTelemetrySetting),
			parameter:      ConsentTelemetrySetting,
			value:          "true",
			existingConfig: Preference{},
			wantErr:        false,
			want:           true,
		},
		{
			name:      fmt.Sprintf("Case 20: set %s from true to false", ConsentTelemetrySetting),
			parameter: ConsentTelemetrySetting,
			value:     "false",
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					ConsentTelemetry: &trueValue,
				},
			},
			wantErr: false,
			want:    false,
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
						t.Errorf("unexpected value after execution of SetConfiguration\ngot: %t \nexpected: %t\n", *cfg.OdoSettings.UpdateNotification, tt.want)
					}
				case "timeout":
					if *cfg.OdoSettings.Timeout != tt.want {
						t.Errorf("unexpected value after execution of SetConfiguration\ngot: %v \nexpected: %d\n", cfg.OdoSettings.Timeout, tt.want)
					}
				case "registrycachetime":
					if *cfg.OdoSettings.RegistryCacheTime != tt.want {
						t.Errorf("unexpected value after execution of SetConfiguration\ngot: %v \nexpected: %d\n", *cfg.OdoSettings.RegistryCacheTime, tt.want)
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

func TestConsentTelemetry(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false

	tests := []struct {
		name           string
		existingConfig Preference
		want           bool
	}{
		{
			name:           fmt.Sprintf("Case 1: %s nil", ConsentTelemetrySetting),
			existingConfig: Preference{},
			want:           false,
		},
		{
			name: fmt.Sprintf("Case 2: %s true", ConsentTelemetrySetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					ConsentTelemetry: &trueValue,
				},
			},
			want: true,
		},
		{
			name: fmt.Sprintf("Case 3: %s false", ConsentTelemetrySetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					ConsentTelemetry: &falseValue,
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
			output := cfg.GetConsentTelemetry()

			if output != tt.want {
				t.Errorf("ConsentTelemetry returned unexpected value\ngot: %t \nexpected: %t\n", output, tt.want)
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
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
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
				t.Errorf("GetUpdateNotification returned unexpected value\ngot: %t \nexpected: %t\n", output, tt.want)
			}

		})
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

func TestPreferenceIsntCreatedWhenOdoIsUsed(t *testing.T) {
	// cleaning up old odo files if any
	filename, err := getPreferenceFile()
	if err != nil {
		t.Error(err)
	}
	os.RemoveAll(filename)

	conf, err := NewPreferenceInfo()
	if err != nil {
		t.Errorf("error while creating global preference %v", err)
	}
	if _, err = os.Stat(conf.Filename); !os.IsNotExist(err) {
		t.Errorf("preference file shouldn't exist yet")
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

func TestHandleWithoutRegistryExist(t *testing.T) {
	tests := []struct {
		name         string
		registryList []Registry
		operation    string
		registryName string
		registryURL  string
		want         []Registry
	}{
		{
			name:         "Case 1: Add registry",
			registryList: []Registry{},
			operation:    "add",
			registryName: "testName",
			registryURL:  "testURL",
			want: []Registry{
				{
					Name: "testName",
					URL:  "testURL",
				},
			},
		},
		{
			name:         "Case 2: Update registry",
			registryList: []Registry{},
			operation:    "update",
			registryName: "testName",
			registryURL:  "testURL",
			want:         nil,
		},
		{
			name:         "Case 3: Delete registry",
			registryList: []Registry{},
			operation:    "delete",
			registryName: "testName",
			registryURL:  "testURL",
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handleWithoutRegistryExist(tt.registryList, tt.operation, tt.registryName, tt.registryURL, false)
			if err != nil {
				t.Logf("Error message is %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got: %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleWithRegistryExist(t *testing.T) {
	tests := []struct {
		name         string
		index        int
		registryList []Registry
		operation    string
		registryName string
		registryURL  string
		forceFlag    bool
		want         []Registry
	}{
		{
			name:  "Case 1: Add registry",
			index: 0,
			registryList: []Registry{
				{
					Name: "testName",
					URL:  "testURL",
				},
			},
			operation:    "add",
			registryName: "testName",
			registryURL:  "addURL",
			forceFlag:    false,
			want:         nil,
		},
		{
			name:  "Case 2: update registry",
			index: 0,
			registryList: []Registry{
				{
					Name: "testName",
					URL:  "testURL",
				},
			},
			operation:    "update",
			registryName: "testName",
			registryURL:  "updateURL",
			forceFlag:    true,
			want: []Registry{
				{
					Name: "testName",
					URL:  "updateURL",
				},
			},
		},
		{
			name:  "Case 3: Delete registry",
			index: 0,
			registryList: []Registry{
				{
					Name: "testName",
					URL:  "testURL",
				},
			},
			operation:    "delete",
			registryName: "testName",
			registryURL:  "",
			forceFlag:    true,
			want:         []Registry{},
		},
	}

	for _, tt := range tests {
		got, err := handleWithRegistryExist(tt.index, tt.registryList, tt.operation, tt.registryName, tt.registryURL, tt.forceFlag, false)
		if err != nil {
			t.Logf("Error message is %v", err)
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Got: %v, want: %v", got, tt.want)
		}
	}
}

func TestGetConsentTelemetry(t *testing.T) {
	tempConfigFile, err := ioutil.TempFile("", "odoconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer tempConfigFile.Close()
	os.Setenv(GlobalConfigEnvName, tempConfigFile.Name())
	trueValue := true
	falseValue := false

	tests := []struct {
		name           string
		existingConfig Preference
		want           bool
	}{{
		name:           fmt.Sprintf("Case 1: %s nil", ConsentTelemetrySetting),
		existingConfig: Preference{},
		want:           false,
	},
		{
			name: fmt.Sprintf("Case 2: %s true", ConsentTelemetrySetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					ConsentTelemetry: &trueValue,
				},
			},
			want: true,
		},
		{
			name: fmt.Sprintf("Case 3: %s false", ConsentTelemetrySetting),
			existingConfig: Preference{
				OdoSettings: OdoSettings{
					ConsentTelemetry: &falseValue,
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
			output := cfg.GetConsentTelemetry()

			if output != tt.want {
				t.Errorf("GetConsentTelemetry returned unexpected value\ngot: %t \nexpected: %t\n", output, tt.want)
			}

		})
	}
}
