package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
)

var testpipelineConfig = &config.PipelinesConfig{Name: "tst-cicd"}
var testArgoCDConfig = &config.ArgoCDConfig{Namespace: "tst-argocd"}
var Config = &config.Config{ArgoCD: testArgoCDConfig, Pipelines: testpipelineConfig}

func TestCreateManifest(t *testing.T) {
	repoURL := "https://github.com/foo/bar.git"
	want := &config.Manifest{
		GitOpsURL: repoURL,
		Config:    Config,
	}
	got := createManifest(repoURL, Config)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("pipelines didn't match: %s\n", diff)
	}
}

func TestInitialFiles(t *testing.T) {
	prefix := "tst-"
	gitOpsURL := "https://github.com/foo/test-repo"
	gitOpsWebhook := "123"
	defer stubDefaultPublicKeyFunc(t)()
	fakeFs := ioutils.NewMapFilesystem()
	repo, err := scm.NewRepository(gitOpsURL)
	assertNoError(t, err)
	got, err := createInitialFiles(fakeFs, repo, prefix, gitOpsWebhook, "", "test-ns")
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		pipelinesFile: createManifest(gitOpsURL, &config.Config{Pipelines: testpipelineConfig}),
	}
	resources, err := createCICDResources(fakeFs, repo, testpipelineConfig, gitOpsWebhook, "", "test-ns")
	if err != nil {
		t.Fatalf("CreatePipelineResources() failed due to :%s\n", err)
	}
	files := getResourceFiles(resources)

	want = res.Merge(addPrefixToResources("config/tst-cicd/base/pipelines", resources), want)
	want = res.Merge(addPrefixToResources("config/tst-cicd", getCICDKustomization(files)), want)

	if diff := cmp.Diff(want, got, cmpopts.IgnoreMapEntries(ignoreSecrets)); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func ignoreSecrets(k string, v interface{}) bool {
	return k == "config/tst-cicd/base/pipelines/03-secrets/gitops-webhook-secret.yaml"
}

func TestGetCICDKustomization(t *testing.T) {
	want := res.Resources{
		"base/kustomization.yaml": res.Kustomization{
			Bases: []string{"./pipelines"},
		},
		"overlays/kustomization.yaml": res.Kustomization{
			Bases: []string{"../base"},
		},
		"base/pipelines/kustomization.yaml": res.Kustomization{
			Resources: []string{"resource1", "resource2"},
		},
	}
	got := getCICDKustomization([]string{"resource1", "resource2"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("getCICDKustomization was not correct: %s\n", diff)
	}

}

func TestAddPrefixToResources(t *testing.T) {
	files := map[string]interface{}{
		"base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}

	want := map[string]interface{}{
		"test-prefix/base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"test-prefix/overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}
	if diff := cmp.Diff(want, addPrefixToResources("test-prefix", files)); diff != "" {
		t.Fatalf("addPrefixToResources failed, diff %s\n", diff)
	}
}

func TestMerge(t *testing.T) {
	map1 := map[string]interface{}{
		"test-1": "value-1",
	}
	map2 := map[string]interface{}{
		"test-1": "value-a",
		"test-2": "value-2",
	}
	map3 := map[string]interface{}{
		"test-1": "value-a",
		"test-2": "value-2",
	}

	want := res.Resources{
		"test-1": "value-1",
		"test-2": "value-2",
	}
	if diff := cmp.Diff(want, res.Merge(map1, map2)); diff != "" {
		t.Fatalf("merge failed: %s\n", diff)
	}
	if diff := cmp.Diff(map2, map3); diff != "" {
		t.Fatalf("original map changed %s\n", diff)
	}

}
