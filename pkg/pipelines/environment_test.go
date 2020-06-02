package pipelines

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/afero"
)

func TestAddEnv(t *testing.T) {
	fakeFs := ioutils.NewMapFilesystem()
	gitopsPath := afero.GetTempDir(fakeFs, "test")

	manifestFile := filepath.Join(gitopsPath, pipelinesFile)
	envParameters := EnvParameters{
		ManifestFilename: manifestFile,
		EnvName:          "dev",
	}
	afero.WriteFile(fakeFs, manifestFile, []byte("environments:"), 0644)

	if err := AddEnv(&envParameters, fakeFs); err != nil {
		t.Fatalf("AddEnv() failed :%s", err)
	}

	wantedPaths := []string{
		"environments/dev/env/base/kustomization.yaml",
		"environments/dev/env/base/dev-environment.yaml",
		"environments/dev/env/overlays/kustomization.yaml",
	}
	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			assertFileExists(rt, fakeFs, filepath.Join(gitopsPath, path))
		})
	}
}

func TestAddEnvWithExistingName(t *testing.T) {
	fakeFs := ioutils.NewMapFilesystem()
	gitopsPath := afero.GetTempDir(fakeFs, "test")

	manifestFile := filepath.Join(gitopsPath, pipelinesFile)
	envParameters := EnvParameters{
		ManifestFilename: manifestFile,
		EnvName:          "dev",
	}
	afero.WriteFile(fakeFs, manifestFile, []byte("environments:\n - name: dev\n"), 0644)

	if err := AddEnv(&envParameters, fakeFs); err == nil {
		t.Fatal("AddEnv() did not fail with duplicate environment")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertFileExists(t *testing.T, testFs afero.Fs, path string) {
	t.Helper()
	exists, err := afero.Exists(testFs, path)
	assertNoError(t, err)
	if !exists {
		t.Fatalf("unable to find file %q", path)
	}
	isDir, err := afero.DirExists(testFs, path)
	assertNoError(t, err)
	if isDir {
		t.Fatalf("%q is a directory", path)
	}
}

func TestNewEnvironment(t *testing.T) {

	tests := []struct {
		m      *config.Manifest
		name   string
		errMsg string
		want   *config.Environment
	}{
		{
			m: &config.Manifest{
				GitOpsURL: "https://github.com/foo/bar",
				Environments: []*config.Environment{
					{
						IsCICD: true,
						Name:   "my-cicd",
					},
				},
			},
			name:   "test-env",
			errMsg: "",
			want: &config.Environment{
				Name: "test-env",
				Pipelines: &config.Pipelines{
					Integration: &config.TemplateBinding{
						Template: appCITemplateName,
						Bindings: []string{"github-pr-binding"},
					},
				},
			},
		},
		{
			m: &config.Manifest{
				GitOpsURL: "https://gitlab.com/foo/bar",
				Environments: []*config.Environment{
					{
						IsCICD: true,
						Name:   "my-cicd",
					},
				},
			},
			name:   "test-env",
			errMsg: "",
			want: &config.Environment{
				Name: "test-env",
				Pipelines: &config.Pipelines{
					Integration: &config.TemplateBinding{
						Template: appCITemplateName,
						Bindings: []string{"gitlab-pr-binding"},
					},
				},
			},
		},
		{
			m: &config.Manifest{
				// no GitOpsURL -> no Pipelines
				Environments: []*config.Environment{
					{
						IsCICD: true,
						Name:   "my-cicd",
					},
				},
			},
			name:   "test-env",
			errMsg: "",
			want: &config.Environment{
				Name: "test-env",
			},
		},
		{
			m: &config.Manifest{
				GitOpsURL: "https://gitlab.com/foo/bar",
				Environments: []*config.Environment{
					{
						// no CICD -> no Pipelines
						IsCICD: false,
						Name:   "my-cicd",
					},
				},
			},
			name:   "test-env",
			errMsg: "",
			want: &config.Environment{
				Name: "test-env",
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test_%d", i), func(tt *testing.T) {
			got, err := newEnvironment(test.m, test.name)
			gotError := ""
			if err != nil {
				gotError = err.Error()
			}
			if diff := cmp.Diff(test.errMsg, gotError); diff != "" {
				tt.Errorf("errMsg mismatch: \n%s", diff)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				tt.Errorf("env mismatch: \n%s", diff)
			}
		})
	}
}
