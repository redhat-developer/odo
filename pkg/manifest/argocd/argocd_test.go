package argocd

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/manifest/config"
	"github.com/openshift/odo/pkg/manifest/meta"

	// This is a hack because ArgoCD doesn't support a compatible (code-wise)
	// version of k8s in common with odo.
	argoappv1 "github.com/openshift/odo/pkg/manifest/argocd/v1alpha1"
	res "github.com/openshift/odo/pkg/manifest/resources"
)

const testRepoURL = "https://github.com/rhd-example-gitops/example"

var (
	testApp = &config.Application{
		Name: "http-api",
	}
	configRepoApp = &config.Application{
		Name: "prod-api",
		ConfigRepo: &config.Repository{
			URL:            "https://github.com/rhd-example-gitops/other-repo",
			Path:           "deploys",
			TargetRevision: "master",
		},
	}

	testEnv = &config.Environment{
		Name: "test-dev",
		Apps: []*config.Application{
			testApp,
		},
	}
	testCICD = &config.Environment{
		Name:   "test-cicd",
		IsCICD: true,
	}
)

func TestBuildCreatesArgoCD(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testCICD,
			testEnv,
			&config.Environment{Name: "argocd", IsArgoCD: true},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"environments/argocd/config/test-dev-http-api-app.yaml": &argoappv1.Application{
			TypeMeta:   applicationTypeMeta,
			ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "test-dev-http-api")),
			Spec: argoappv1.ApplicationSpec{
				Source: makeSource(testEnv, testEnv.Apps[0], testRepoURL),
				Destination: argoappv1.ApplicationDestination{
					Server:    defaultServer,
					Namespace: "test-dev",
				},
				Project:    defaultProject,
				SyncPolicy: syncPolicy,
			},
		},
		"environments/argocd/config/gitops-app.yaml":    makeApplication(gitOpsApp, ArgoCDNamespace, defaultProject, "test-cicd", defaultServer, argoappv1.ApplicationSource{RepoURL: testRepoURL, Path: filepath.Join(config.PathForEnvironment(testCICD), "base")}),
		"environments/argocd/config/kustomization.yaml": &res.Kustomization{Resources: []string{"gitops-app.yaml", "test-dev-http-api-app.yaml"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildCreatesArgoCDWithMultipleApps(t *testing.T) {
	prodEnv := &config.Environment{
		Name: "test-production",
		Apps: []*config.Application{
			testApp,
		},
	}
	m := &config.Manifest{
		Environments: []*config.Environment{
			prodEnv,
			testEnv,
			&config.Environment{Name: "argocd", IsArgoCD: true},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("got %d files, want 3\n", len(files))
	}
	want := &res.Kustomization{Resources: []string{"test-dev-http-api-app.yaml", "test-production-http-api-app.yaml"}}
	if diff := cmp.Diff(want, files["environments/argocd/config/kustomization.yaml"]); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithNoRepoURL(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testEnv,
			&config.Environment{Name: "argocd", IsArgoCD: true},
		},
	}

	files, err := Build(ArgoCDNamespace, "", m)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithNoArgoCDEnv(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			testEnv,
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildWithRepoConfig(t *testing.T) {
	prodEnv := &config.Environment{
		Name: "test-production",
		Apps: []*config.Application{
			configRepoApp,
		},
	}

	m := &config.Manifest{
		Environments: []*config.Environment{
			prodEnv,
			&config.Environment{Name: "argocd", IsArgoCD: true},
		},
	}

	files, err := Build(ArgoCDNamespace, testRepoURL, m)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"environments/argocd/config/test-production-prod-api-app.yaml": &argoappv1.Application{
			TypeMeta:   applicationTypeMeta,
			ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ArgoCDNamespace, "test-production-prod-api")),
			Spec: argoappv1.ApplicationSpec{
				Source: makeSource(prodEnv, prodEnv.Apps[0], testRepoURL),
				Destination: argoappv1.ApplicationDestination{
					Server:    defaultServer,
					Namespace: "test-production",
				},
				Project:    defaultProject,
				SyncPolicy: syncPolicy,
			},
		},
		"environments/argocd/config/kustomization.yaml": &res.Kustomization{Resources: []string{"test-production-prod-api-app.yaml"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}
