package pipelines

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/secrets"
)

var testCICDEnv = &config.Environment{Name: "tst-cicd", IsCICD: true}

func TestCreateManifest(t *testing.T) {
	want := &config.Manifest{
		GitOpsURL: "https://github.com/foo/bar.git",
		Environments: []*config.Environment{
			testCICDEnv,
		},
	}
	got := createManifest("https://github.com/foo/bar.git", testCICDEnv)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("pipelines didn't match: %s\n", diff)
	}
}

func TestInitialFiles(t *testing.T) {
	prefix := "tst-"
	gitOpsURL := "https://gibhub.com/foo/test-repo"
	gitOpsWebhook := "123"
	defer func(f secrets.PublicKeyFunc) {
		secrets.DefaultPublicKeyFunc = f
	}(secrets.DefaultPublicKeyFunc)

	secrets.DefaultPublicKeyFunc = func() (*rsa.PublicKey, error) {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatalf("failed to generate a private RSA key: %s", err)
		}
		return &key.PublicKey, nil
	}
	fakeFs := ioutils.NewMapFilesystem()

	got, err := createInitialFiles(fakeFs, prefix, gitOpsURL, gitOpsWebhook, "")
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		pipelinesFile: createManifest(gitOpsURL, testCICDEnv),
	}
	gitOpsRepo, err := orgRepoFromURL(gitOpsURL)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := createCICDResources(fakeFs, testCICDEnv, gitOpsRepo, gitOpsWebhook, "")
	if err != nil {
		t.Fatalf("CreatePipelineResources() failed due to :%s\n", err)
	}
	files := getResourceFiles(resources)

	want = res.Merge(addPrefixToResources("environments/tst-cicd/base/pipelines", resources), want)
	want = res.Merge(addPrefixToResources("environments/tst-cicd", getCICDKustomization(files)), want)

	if diff := cmp.Diff(want, got, cmpopts.IgnoreMapEntries(ignoreSecrets)); diff != "" {
		t.Fatalf("outputs didn't match: %s\n", diff)
	}
}

func ignoreSecrets(k string, v interface{}) bool {
	if k == "environments/tst-cicd/base/pipelines/03-secrets/gitops-webhook-secret.yaml" {
		return true
	}
	return false
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

func TestValidateImageRepo(t *testing.T) {

	errorMsg := "failed to parse image repo:%s, expected image repository in the form <registry>/<username>/<repository> or <project>/<app> for internal registry"

	tests := []struct {
		description                string
		options                    InitParameters
		expectedError              string
		expectedIsInternalRegistry bool
		expectedImageRepo          string
	}{
		{
			"Valid image regsitry URL",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user/sample-repo",
			},
			"",
			false,
			"quay.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL random registry",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "random.io/sample-user/sample-repo",
			},
			"",
			false,
			"random.io/sample-user/sample-repo",
		},
		{
			"Valid image regsitry URL docker.io",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/sample-user/sample-repo",
			},
			"",
			false,
			"docker.io/sample-user/sample-repo",
		},
		{
			"Invalid image registry URL with missing repo name",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user",
			},
			fmt.Sprintf(errorMsg, "quay.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with missing repo name docker.io",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/sample-user",
			},
			fmt.Sprintf(errorMsg, "docker.io/sample-user"),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sample-user/ ",
			},
			fmt.Sprintf(errorMsg, "quay.io/sample-user/ "),
			false,
			"",
		},
		{
			"Invalid image registry URL with whitespaces in between",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/sam\tple-user/",
			},
			fmt.Sprintf(errorMsg, "quay.io/sam\tple-user/"),
			false,
			"",
		},
		{
			"Invalid image registry URL with leading whitespaces",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "quay.io/ sample-user/",
			},
			fmt.Sprintf(errorMsg, "quay.io/ sample-user/"),
			false,
			"",
		},
		{
			"Valid internal registry URL",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "image-registry.openshift-image-registry.svc:5000/project/app",
			},
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid internal registry URL implicit starts with '/'",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "/project/app",
			},
			fmt.Sprintf(errorMsg, "/project/app"),
			false,
			"",
		},
		{
			"Valid internal registry URL implicit",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "project/app",
			},
			"",
			true,
			"image-registry.openshift-image-registry.svc:5000/project/app",
		},
		{
			"Invalid too many URL components docker",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io/foo/project/app",
			},
			fmt.Sprintf(errorMsg, "docker.io/foo/project/app"),
			false,
			"",
		},
		{
			"Invalid too many URL components internal",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "image-registry.openshift-image-registry.svc:5000/project/app/foo",
			},
			fmt.Sprintf(errorMsg, "image-registry.openshift-image-registry.svc:5000/project/app/foo"),
			false,
			"",
		},
		{
			"Invalid not enough URL components, no slash",
			InitParameters{
				InternalRegistryHostname: "image-registry.openshift-image-registry.svc:5000",
				ImageRepo:                "docker.io",
			},
			fmt.Sprintf(errorMsg, "docker.io"),
			false,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			isInternalRegistry, imageRepo, error := validateImageRepo(test.options.ImageRepo,
				test.options.InternalRegistryHostname,
			)
			if diff := cmp.Diff(isInternalRegistry, test.expectedIsInternalRegistry); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
			if diff := cmp.Diff(imageRepo, test.expectedImageRepo); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
			errorString := ""
			if error != nil {
				errorString = error.Error()
			}
			if diff := cmp.Diff(errorString, test.expectedError); diff != "" {
				t.Errorf("validateImageRepo() failed:\n%s", diff)
			}
		})
	}
}
