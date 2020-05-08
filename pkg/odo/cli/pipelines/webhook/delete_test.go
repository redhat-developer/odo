package webhook

import (
	"fmt"
	"testing"
)

func TestMissingRequiredFlagsForDelete(t *testing.T) {

	testcases := []struct {
		flags   []keyValuePair
		wantErr string
	}{
		{[]keyValuePair{flag("cicd", "true")},
			"Required flag(s) \"access-token\" have/has not been set",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			_, _, err := executeCommand(newCmdDelete("webhook", "odo pipelines webhook delete"), tt.flags...)

			if err != nil {
				if err.Error() != tt.wantErr {
					rt.Errorf("got %s, want %s", err, tt.wantErr)
				}
			} else {
				if tt.wantErr != "" {
					rt.Errorf("got %s, want %s", "", tt.wantErr)
				}
			}
		})
	}
}

func TestValidateForDelete(t *testing.T) {

	testcases := []struct {
		options *deleteOptions
		errMsg  string
	}{
		{
			&deleteOptions{
				options{isCICD: true, serviceName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: true, envName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: true, envName: "foo", serviceName: "bar"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false, serviceName: "foo"},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false, serviceName: "foo", envName: "gau"},
			},
			"",
		},
		{
			&deleteOptions{
				options{isCICD: true, serviceName: ""},
			},
			"",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			err := tt.options.Validate()
			if err != nil && tt.errMsg == "" {
				t.Errorf("Validate() got an unexpected error: %s", err)
			} else {
				if !matchError(t, tt.errMsg, err) {
					t.Errorf("Validate() failed to match error: got %s, want %s", err, tt.errMsg)
				}
			}
		})
	}
}
