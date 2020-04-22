package manifest

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteBootstrapParameters(t *testing.T) {
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
		o := BootstrapParameters{prefix: tt.prefix}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
	}
}

func TestValidateBootstrapParameters(t *testing.T) {
	optionTests := []struct {
		name    string
		gitRepo string
		errMsg  string
	}{
		{"invalid repo", "test", "repo must be org/repo"},
		{"valid repo", "test/repo", ""},
	}

	for _, tt := range optionTests {
		o := BootstrapParameters{gitOpsRepo: tt.gitRepo, prefix: "test"}

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

func TestBypassBootstrapChecks(t *testing.T) {
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
			o := BootstrapParameters{skipChecks: test.skipChecks}

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
