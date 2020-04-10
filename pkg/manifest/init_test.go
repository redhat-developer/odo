package manifest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/manifest/config"
)

func TestCreateManifest(t *testing.T) {
	want := &config.Manifest{
		Environments: []*config.Environment{
			{
				Name:   "tst-cicd",
				IsCICD: true,
			},
		},
	}
	got := createManifest("tst-")

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("manifest didn't match: %s\n", diff)
	}
}

func TestInitialFiles(t *testing.T) {
	prefix := "tst-"
	got, err := createInitialFiles(prefix)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]interface{}{
		"manifest.yaml": createManifest(prefix),
	}
	want = merge(addPrefixToResources("environments/tst-cicd", getCICDKustomization()), want)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func TestGetCICDKustomization(t *testing.T) {
	want := resources{
		"base/kustomization.yaml": map[string]interface{}{
			"resources": []string{},
		},
		"overlays/kustomization.yaml": map[string]interface{}{
			"bases": []string{"../base"},
		},
	}
	got := getCICDKustomization()
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

	want := resources{
		"test-1": "value-1",
		"test-2": "value-2",
	}
	if diff := cmp.Diff(want, merge(map1, map2)); diff != "" {
		t.Fatalf("merge failed: %s\n", diff)
	}
	if diff := cmp.Diff(map2, map3); diff != "" {
		t.Fatalf("original map changed %s\n", diff)
	}

}
