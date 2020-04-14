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

func TestCompleteInitParameters(t *testing.T) {
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
		o := InitParameters{prefix: tt.prefix}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
	}
}

func TestValidateInitParameters(t *testing.T) {
	optionTests := []struct {
		name    string
		gitRepo string
		errMsg  string
	}{
		{"invalid repo", "test", "repo must be org/repo"},
		{"valid repo", "test/repo", ""},
	}

	for _, tt := range optionTests {
		o := InitParameters{gitOpsRepo: tt.gitRepo, prefix: "test"}

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

func TestInitCommandWithMissingParams(t *testing.T) {
	cmdTests := []struct {
		desc    string
		flags   []keyValuePair
		wantErr string
	}{
		{"Missing gitops-repo flag",
			[]keyValuePair{flag("output", "~/output"),
				flag("gitops-webhook-secret", "123"), flag("skip-checks", "true")},
			`Required flag(s) "gitops-repo" have/has not been set`},
		{"Missing gitops-webhook-secret flag",
			[]keyValuePair{flag("gitops-repo", "org/sample"), flag("output", "~/output"),
				flag("skip-checks", "true")},
			`Required flag(s) "gitops-webhook-secret" have/has not been set`},
		{"Missing output flag",
			[]keyValuePair{flag("gitops-repo", "org/sample"), flag("gitops-webhook-secret", "123"),
				flag("skip-checks", "true")},
			`Required flag(s) "output" have/has not been set`},
	}
	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := executeCommand(NewCmdInit("init", "odo pipelines init"), tt.flags...)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

func TestBypassChecks(t *testing.T) {
	tests := []struct {
		description        string
		skipChecks         bool
		wantedBypassChecks bool
	}{
		{"bypass tekton installation checks", true, true},
		{"don't bypass tekton installation checks", false, false},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := InitParameters{skipChecks: test.skipChecks}

			err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

			if err != nil {
				t.Errorf("Complete() %#v failed: ", err)
			}

			if o.skipChecks != test.wantedBypassChecks {
				t.Errorf("Complete() %#v bypassChecks flag: got %v, want %v", test.description, o.skipChecks, test.wantedBypassChecks)
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
