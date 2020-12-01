package kclient

import (
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"testing"
)

func TestGetInputEnvVarsFromStrings(t *testing.T) {
	tests := []struct {
		name          string
		envVars       []string
		wantedEnvVars []corev1.EnvVar
		wantErr       bool
	}{
		{
			name:    "Test case 1: with valid two key value pairs",
			envVars: []string{"key=value", "key1=value1"},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "value1",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 2: one env var with missing value",
			envVars: []string{"key=value", "key1="},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 3: one env var with no value and no =",
			envVars: []string{"key=value", "key1"},
			wantErr: true,
		},
		{
			name:    "Test case 4: one env var with multiple values",
			envVars: []string{"key=value", "key1=value1=value2"},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "value1=value2",
				},
			},
			wantErr: false,
		},
		{
			name:    "Test case 5: two env var with same key",
			envVars: []string{"key=value", "key=value1"},
			wantErr: true,
		},
		{
			name:    "Test case 6: one env var with base64 encoded value",
			envVars: []string{"key=value", "key1=SSd2ZSBnb3QgYSBsb3ZlbHkgYnVuY2ggb2YgY29jb251dHMhCg=="},
			wantedEnvVars: []corev1.EnvVar{
				{
					Name:  "key",
					Value: "value",
				},
				{
					Name:  "key1",
					Value: "SSd2ZSBnb3QgYSBsb3ZlbHkgYnVuY2ggb2YgY29jb251dHMhCg==",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars, err := GetInputEnvVarsFromStrings(tt.envVars)

			if err == nil && !tt.wantErr {
				if !reflect.DeepEqual(tt.wantedEnvVars, envVars) {
					t.Errorf("corev1.Env values are not matching with expected values, expected: %v, got %v", tt.wantedEnvVars, envVars)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}
