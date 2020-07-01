package pipelines

import (
	"testing"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
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
		o := BootstrapParameters{&pipelines.BootstrapOptions{Prefix: tt.prefix}, &genericclioptions.Context{}}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.Prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.Prefix, tt.wantPrefix)
		}
	}
}

func TestAddSuffixWithBootstrap(t *testing.T) {
	gitOpsURL := "https://github.com/org/gitops"
	appURL := "https://github.com/org/app"
	tt := []struct {
		name           string
		gitOpsURL      string
		appURL         string
		validGitOpsURL string
		validAppURL    string
	}{
		{"empty string", "", "", "", ""},
		{"suffix already exists", gitOpsURL + ".git", appURL + ".git", gitOpsURL + ".git", appURL + ".git"},
		{"misssing suffix", gitOpsURL, appURL, gitOpsURL + ".git", appURL + ".git"},
	}

	for _, test := range tt {
		t.Run(test.name, func(rt *testing.T) {
			o := BootstrapParameters{&pipelines.BootstrapOptions{
				GitOpsRepoURL: test.gitOpsURL, ServiceRepoURL: test.appURL},
				&genericclioptions.Context{}}

			err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})
			if err != nil {
				t.Errorf("Complete() %#v failed: ", err)
			}

			if o.GitOpsRepoURL != test.validGitOpsURL {
				rt.Fatalf("URL mismatch: got %s, want %s", o.GitOpsRepoURL, test.validAppURL)
			}
			if o.ServiceRepoURL != test.validAppURL {
				rt.Fatalf("URL mismatch: got %s, want %s", o.GitOpsRepoURL, test.validAppURL)
			}
		})
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
		o := BootstrapParameters{
			&pipelines.BootstrapOptions{
				GitOpsRepoURL: tt.gitRepo,
				Prefix:        "test",
			},
			&genericclioptions.Context{},
		}
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
