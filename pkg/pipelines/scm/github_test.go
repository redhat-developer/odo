package scm

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ Repository = (*GitHubRepository)(nil)

func TestCreatePRBindingForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "github-pr-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.sha)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "fullname",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.full_name)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreatePRBinding("testns")
	if name != githubPRBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", githubPRBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("createPRBinding() failed:\n%s", diff)
	}
}

func TestCreatePushBindingForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "github-push-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.head_commit.id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	got, name := repo.CreatePushBinding("testns")
	if name != githubPushBindingName {
		t.Fatalf("CreatePushBinding() returned a wrong binding: want %v got %v", githubPushBindingName, name)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreatePushBinding() failed:\n%s", diff)
	}
}

func TestCreateCITriggerForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.EventListenerTrigger{
		Name: "test",
		Bindings: []*triggersv1.EventListenerBinding{
			&triggersv1.EventListenerBinding{Name: "test-binding"},
		},
		Template: triggersv1.EventListenerTemplate{Name: "test-template"},
		Interceptors: []*triggersv1.EventInterceptor{
			&triggersv1.EventInterceptor{
				CEL: &triggersv1.CELInterceptor{
					Filter: fmt.Sprintf(githubCIDryRunFilters, "org/test"),
				},
			},
			&triggersv1.EventInterceptor{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{SecretKey: "webhook-secret-key", SecretName: "secret", Namespace: "ns"},
				},
			},
		},
	}
	got := repo.CreateCITrigger("test", "secret", "ns", "test-template", []string{"test-binding"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreateCITrigger() failed:\n%s", diff)
	}
}

func TestCreateCDTriggersForGithub(t *testing.T) {
	repo, err := NewGitHubRepository("http://github.com/org/test")
	assertNoError(t, err)
	want := triggersv1.EventListenerTrigger{
		Name: "test",
		Bindings: []*triggersv1.EventListenerBinding{
			&triggersv1.EventListenerBinding{Name: "test-binding"},
		},
		Template: triggersv1.EventListenerTemplate{Name: "test-template"},
		Interceptors: []*triggersv1.EventInterceptor{
			&triggersv1.EventInterceptor{
				CEL: &triggersv1.CELInterceptor{
					Filter: fmt.Sprintf(githubCDDeployFilters, "org/test"),
				},
			},
			&triggersv1.EventInterceptor{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{SecretKey: "webhook-secret-key", SecretName: "secret", Namespace: "ns"},
				},
			},
		},
	}
	got := repo.CreateCDTrigger("test", "secret", "ns", "test-template", []string{"test-binding"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("CreateCDTrigger() failed:\n%s", diff)
	}
}

func TestNewGitHubRepository(t *testing.T) {
	tests := []struct {
		url      string
		repoPath string
		errMsg   string
	}{
		{
			"http://github.org",
			"",
			"unable to determine repo path from: http://github.org",
		},
		{
			"http://github.com/",
			"",
			"unable to determine repo path from: http://github.com/",
		},
		{
			"http://github.com/foo/bar",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar.git",
			"foo/bar",
			"",
		},
		{
			"https://githuB.com/foo/bar/test.git",
			"foo/bar",
			"",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test %d", i), func(rt *testing.T) {
			repo, err := NewGitHubRepository(tt.url)
			if err != nil {
				if diff := cmp.Diff(tt.errMsg, err.Error()); diff != "" {
					rt.Fatalf("repo path errMsg mismatch: \n%s", diff)
				}
			}
			if repo != nil {
				if diff := cmp.Diff(tt.repoPath, repo.path); diff != "" {
					rt.Fatalf("repo path mismatch: got\n%s", diff)
				}
			}
		})
	}
}
