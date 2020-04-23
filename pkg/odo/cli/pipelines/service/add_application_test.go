package service

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
)

type keyValuePair struct {
	key   string
	value string
}

func TestCompleteAddParameters(t *testing.T) {
	completeTests := []struct {
		name       string
		prefix     string
		wantPrefix string
	}{
		{"no prefix", "", ""},
		{"prefix with hyphen", "test-", "test-"},
		{"prefix without hyphen", "test", "test-"},
	}

	for _, tt := range completeTests {
		o := AddParameters{prefix: tt.prefix}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
	}
}

func TestValidateAddParameters(t *testing.T) {
	optionTests := []struct {
		name        string
		errMsgGit   string
		serviceRepo string
		errMsgApp   string
	}{
		{"invalid repo", "repo must be org/repo", "test2", "repo must be org/repo"},
		{"valid repo", "", "test2/repo", ""},
	}

	for _, tt := range optionTests {
		o := AddParameters{prefix: "test", serviceGitRepo: tt.serviceRepo}

		err := o.Validate()

		if err != nil && tt.errMsgGit == "" {
			t.Errorf("Validate() %#v got an unexpected error: %s", tt.name, err)
			continue
		}

		if !matchError(t, tt.errMsgGit, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsgGit)
		}
		if !matchError(t, tt.errMsgApp, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsgApp)
		}
	}
}

func TestAddCommandWithMissingParams(t *testing.T) {
	cmdTests := []struct {
		desc    string
		flags   []keyValuePair
		wantErr string
	}{
		{"Missing app-name flag",
			[]keyValuePair{flag("gitops-repo", "example/repo"),
				flag("service-git-repo", "example/repo"), flag("service-webhook-secret", "abc123"), flag("env-name", "sample"), flag("skip-checks", "true"), flag("output", "./")},
			`Required flag(s) "app-name" have/has not been set`},
		{"Missing service-git-repo flag",
			[]keyValuePair{flag("app-name", "app"),
				flag("gitops-repo", "example/repo"), flag("service-webhook-secret", "abc123"), flag("env-name", "sample"), flag("skip-checks", "true"), flag("output", "./")},
			`Required flag(s) "service-git-repo" have/has not been set`},
		{"Missing service-webhook-secret flag",
			[]keyValuePair{flag("app-name", "app"),
				flag("gitops-repo", "example/repo"), flag("service-git-repo", "sample/repo"), flag("env-name", "sample"), flag("skip-checks", "true"), flag("output", "./")},
			`Required flag(s) "service-webhook-secret" have/has not been set`},
		{"Missing env-name flag",
			[]keyValuePair{flag("app-name", "app"),
				flag("gitops-repo", "example/repo"), flag("service-git-repo", "sample/repo"), flag("service-webhook-secret", "abc123"), flag("skip-checks", "true"), flag("output", "./")},
			`Required flag(s) "env-name" have/has not been set`},
		{"Missing output flag",
			[]keyValuePair{flag("output", "app"),
				flag("gitops-repo", "example/repo"), flag("service-git-repo", "sample/repo"), flag("service-webhook-secret", "abc123"), flag("skip-checks", "true"), flag("app-name", "example-1")},
			`Required flag(s) "env-name" have/has not been set`},
	}
	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := executeCommand(NewCmdAddService("add", "odo pipelines serviec"), tt.flags...)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
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

func flag(k, v string) keyValuePair {
	return keyValuePair{
		key:   k,
		value: v,
	}
}
