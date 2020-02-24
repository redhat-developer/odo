package pipelines

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

func TestCompleteBootstrapOptions(t *testing.T) {
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
		o := BootstrapOptions{prefix: tt.prefix}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
	}
}

func TestValidateBootstrapOptions(t *testing.T) {
	optionTests := []struct {
		name    string
		gitRepo string
		errMsg  string
	}{
		{"invalid repo", "test", "repo must be org/repo"},
		{"valid repo", "test/repo", ""},
	}

	for _, tt := range optionTests {
		o := BootstrapOptions{quayUsername: "testing", gitRepo: tt.gitRepo, prefix: "test"}

		err := o.Validate()

		if err != nil && tt.errMsg == "" {
			t.Errorf("Validate() %#v got an unexpected error: %s", tt.name, err)
			continue
		}

		if !matchError(t, tt.errMsg, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsg)
		}
	}
}
func TestBootstrapCommandWithMissingParams(t *testing.T) {
	cmdTests := []struct {
		flags   []keyValuePair
		wantErr string
	}{
		{[]keyValuePair{flag("quay-username", "example"), flag("github-token", "abc123"), flag("dockerconfigjson", "~/")}, `Required flag(s) "git-repo" have/has not been set`},
		{[]keyValuePair{flag("quay-username", "example"), flag("github-token", "abc123"), flag("git-repo", "example/repo")}, `Required flag(s) "dockerconfigjson" have/has not been set`},
		{[]keyValuePair{flag("quay-username", "example"), flag("dockerconfigjson", "~/"), flag("git-repo", "example/repo")}, `Required flag(s) "github-token" have/has not been set`},
		{[]keyValuePair{flag("github-token", "abc123"), flag("dockerconfigjson", "~/"), flag("git-repo", "example/repo")}, `Required flag(s) "quay-username" have/has not been set`},
	}
	for _, tt := range cmdTests {
		_, _, err := executeCommand(NewCmdBootstrap("bootstrap", "odo pipelines bootstrap"), tt.flags...)
		if err.Error() != tt.wantErr {
			t.Errorf("got %s, want %s", err, tt.wantErr)
		}
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
