package webhook

import (
	"bytes"
	"fmt"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
)

type keyValuePair struct {
	key   string
	value string
}

func TestMissingRequiredFlagsForCreate(t *testing.T) {
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
			_, _, err := executeCommand(newCmdCreate("webhook", "odo pipelines webhook create"), tt.flags...)

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

func TestValidateForCreate(t *testing.T) {
	testcases := []struct {
		options *createOptions
		errMsg  string
	}{
		{
			&createOptions{
				options{isCICD: true, serviceName: "foo"},
			},
			"Only one of 'cicd' or 'app-name/env-name/service-name' can be specified",
		},
		{
			&createOptions{
				options{isCICD: true, appName: "foo"},
			},
			"Only one of 'cicd' or 'app-name/env-name/service-name' can be specified",
		},
		{
			&createOptions{
				options{isCICD: true, envName: "foo"},
			},
			"Only one of 'cicd' or 'app-name/env-name/service-name' can be specified",
		},
		{
			&createOptions{
				options{isCICD: true, envName: "foo", serviceName: "bar", appName: "gau"},
			},
			"Only one of 'cicd' or 'app-name/env-name/service-name' can be specified",
		},
		{
			&createOptions{
				options{isCICD: false},
			},
			"One of 'cicd' or 'app-name/env-name/service-name' must be specified",
		},
		{
			&createOptions{
				options{isCICD: false, serviceName: "foo"},
			},
			"One of 'cicd' or 'app-name/env-name/service-name' must be specified",
		},
		{
			&createOptions{
				options{isCICD: false, serviceName: "foo", appName: "bar"},
			},
			"One of 'cicd' or 'app-name/env-name/service-name' must be specified",
		},
		{
			&createOptions{
				options{isCICD: false, serviceName: "foo", appName: "bar", envName: "gau"},
			},
			"",
		},
		{
			&createOptions{
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

func executeCommand(cmd *cobra.Command, flags ...keyValuePair) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)
	for _, flag := range flags {
		cmd.Flags().Set(flag.key, flag.value)
	}
	c, err = cmd.ExecuteC()
	return c, buf.String(), err
}

func flag(k, v string) keyValuePair {
	return keyValuePair{
		key:   k,
		value: v,
	}
}

func matchError(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}
