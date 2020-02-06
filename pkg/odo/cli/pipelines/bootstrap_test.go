package pipelines

import (
	"regexp"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateBootstrapOptions(t *testing.T) {
	optionTests := []struct {
		name     string
		baseRepo string
		errMsg   string
	}{
		{"invalid repo", "test", "repo must be org/repo"},
		{"valid repo", "test/repo", ""},
	}

	for _, tt := range optionTests {
		o := BootstrapOptions{quayUsername: "testing", baseRepo: tt.baseRepo, prefix: "test"}

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

// TODO: set up for complete BootstrapOptions instead of just prefix.
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
			t.Fatalf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
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
