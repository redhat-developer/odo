package webhook

import (
	"fmt"
	"testing"
)

func TestMissingRequiredFlagsForList(t *testing.T) {

	testcases := []struct {
		flags   []keyValuePair
		wantErr string
	}{
		{[]keyValuePair{flag("cicd", "true")},
			`required flag(s) "access-token" not set`,
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			_, _, err := executeCommand(newCmdList("webhook", "odo pipelines webhook create"), tt.flags...)

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

func TestValidateForList(t *testing.T) {

	testcases := []struct {
		options *listOptions
		errMsg  string
	}{
		{
			&listOptions{
				options{isCICD: true, serviceName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&listOptions{
				options{isCICD: true, envName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&listOptions{
				options{isCICD: true, envName: "foo", serviceName: "bar"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&listOptions{
				options{isCICD: false},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&listOptions{
				options{isCICD: false, serviceName: "foo"},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&listOptions{
				options{isCICD: false, serviceName: "foo", envName: "gau"},
			},
			"",
		},
		{
			&listOptions{
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
