package config

import (
	"reflect"
	"testing"
)

func TestNewEnvVarFromString(t *testing.T) {
	cases := []struct {
		envStr   string
		expected EnvVar
		wantErr  bool
	}{
		{
			envStr: "foo=bar",
			expected: EnvVar{
				Name:  "foo",
				Value: "bar",
			},
		},
		{
			envStr:   "foo",
			expected: EnvVar{},
			wantErr:  true,
		},
		{
			envStr: " foo=bar ",
			expected: EnvVar{
				Name:  "foo",
				Value: "bar",
			},
		},
	}

	for _, testCase := range cases {
		envVar, err := NewEnvVarFromString(testCase.envStr)
		// expected an error
		if testCase.wantErr {
			emptyEnvVar := EnvVar{}
			if envVar != emptyEnvVar || err == nil {
				t.Errorf("expected error for %s", testCase.envStr)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(envVar, testCase.expected) {
				t.Errorf("the %+v and %+v are not equal", envVar, testCase.expected)
			}
		}
	}
}

func TestNewEnvVarListFromSlice(t *testing.T) {
	cases := []struct {
		envList  []string
		expected EnvVarList
		wantErr  bool
	}{
		{
			envList: []string{"foo=bar"},
			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
			},
		},
		{
			envList:  []string{"foo"},
			expected: nil,
			wantErr:  true,
		},
		{
			envList: []string{" foo=bar "},
			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
			},
		},
		{
			envList: []string{"foo=bar", "fizz=buzz"},

			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "fizz",
					Value: "buzz",
				},
			},
		},
		{
			envList: []string{"foo=bar", "fizz=buzz", "test"},

			expected: nil,
			wantErr:  true,
		},
		{
			envList: []string{"foo=bar="},
			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar=",
				},
			},
		},
	}

	for _, testCase := range cases {

		envVarList, err := NewEnvVarListFromSlice(testCase.envList)
		// expected an error
		if testCase.wantErr {
			if envVarList != nil || err == nil {
				t.Errorf("expected error for %s", testCase.envList)
			}
		} else {
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(envVarList, testCase.expected) {
				t.Errorf("the %+v and %+v are not equal", envVarList, testCase.expected)
			}
		}
	}
}

func TestRemoveEnvVarsFromList(t *testing.T) {
	cases := []struct {
		envVarList EnvVarList
		expected   EnvVarList
		keys       []string
	}{
		{
			envVarList: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "fizz",
					Value: "buzz",
				},
			},
			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
			},
			keys: []string{"fizz"},
		},
		{
			envVarList: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "fizz",
					Value: "buzz",
				},
			},
			expected: EnvVarList{},
			keys:     []string{"fizz", "foo"},
		},
		{
			envVarList: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "fizz",
					Value: "buzz",
				},
			},
			expected: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "fizz",
					Value: "buzz",
				},
			},
		},
		{
			envVarList: EnvVarList{
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
				EnvVar{
					Name:  "foo",
					Value: "bar",
				},
			},
			expected: EnvVarList{},
			keys:     []string{"foo"},
		},
	}

	for _, testCase := range cases {

		envVarList := RemoveEnvVarsFromList(testCase.envVarList, testCase.keys)
		// expected an error

		if !reflect.DeepEqual(envVarList, testCase.expected) {
			t.Errorf("the %+v and %+v are not equal", envVarList, testCase.expected)
		}
	}

}
