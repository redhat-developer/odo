package kclient

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
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
				if diff := cmp.Diff(tt.wantedEnvVars, envVars); diff != "" {
					t.Errorf("GetInputEnvVarsFromStrings() wantedEnvVars mismatch (-want +got):\n%s", diff)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func Test_getErrorMessageFromEvents(t *testing.T) {
	type args struct {
		failedEvents map[string]corev1.Event
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "case 1: three failed events",
			args: args{
				failedEvents: map[string]corev1.Event{
					"pvc": {
						Reason:  "pvc not bound",
						Message: "pvc not bound",
						Count:   5,
					},
					"deployment": {
						Reason:  "deployment not running",
						Message: "deployment not running",
						Count:   3,
					},
					"pod": {
						Reason:  "pod not running",
						Message: "pod not running",
						Count:   8,
					},
				},
			},
			want: []string{"pvc not bound", "deployment not running", "pod not running", "8", "5", "3"},
		},
		{
			name: "case 2: no failed events",
			args: args{
				failedEvents: map[string]corev1.Event{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getErrorMessageFromEvents(tt.args.failedEvents)

			for _, wantString := range tt.want {
				if !strings.Contains(got.String(), wantString) {
					t.Errorf("getErrorMessageFromEvents() out: %s, doesn't contain: %s", got.String(), wantString)
				}
			}
		})
	}
}
