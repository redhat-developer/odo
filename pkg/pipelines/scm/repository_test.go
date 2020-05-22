package scm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewRepository(t *testing.T) {
	githubURL := "http://github.com/org/test"
	got, err := NewRepository(githubURL)
	assertNoError(t, err)
	want, err := NewGitHubRepository(githubURL)
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmp.AllowUnexported(GitHubRepository{})); diff != "" {
		t.Fatalf("NewRepository() failed:\n%s", diff)
	}
}

func TestNewRepositoryForInvalidRepoType(t *testing.T) {
	githubURL := "http://test.com/org/test"
	repoType := "test"
	_, gotErr := NewRepository(githubURL)
	if gotErr == nil {
		t.Fatalf("NewRepository() returned an invalid repository of type: %s", repoType)
	}
	wantErr := invalidRepoTypeError(githubURL)
	if diff := cmp.Diff(wantErr.Error(), gotErr.Error()); diff != "" {
		t.Fatalf("Errors don't match: got %v want %v", gotErr, wantErr)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
