package scm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewRepositoryGitHub(t *testing.T) {
	githubURL := "http://github.com/org/test"
	got, err := NewRepository(githubURL)
	assertNoError(t, err)
	want, err := newGitHub(githubURL)
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmp.AllowUnexported(githubSpec{}, repository{})); diff != "" {
		t.Fatalf("NewRepository() failed:\n%s", diff)
	}
}

func TestNewRepositoryGitLab(t *testing.T) {
	gitlabURL := "http://gitlab.com/org/test"
	got, err := NewRepository(gitlabURL)
	assertNoError(t, err)
	want, err := newGitLab(gitlabURL)
	assertNoError(t, err)
	if diff := cmp.Diff(got, want, cmp.AllowUnexported(gitlabSpec{}, repository{})); diff != "" {
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
	wantErr := unsupportedGitTypeError(repoType)
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
