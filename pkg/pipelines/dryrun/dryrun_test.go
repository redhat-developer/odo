package dryrun

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/tasks"
	"github.com/openshift/odo/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

var (
	logsWithArgoCD = strings.Join([]string{
		"Apply argocd applications",
		"Apply cicd environment",
		"Apply taxi application",
		"Apply go-app application\n",
	}, "\n")

	logsWithoutArgoCD = strings.Join([]string{
		"Apply cicd environment",
		"Apply dev environment",
		"Apply stage environment\n",
	}, "\n")
)

func TestMakeScriptWithArgo(t *testing.T) {
	tempDir, cleanup := tempDir(t)
	defer cleanup()

	fs := ioutils.NewFilesystem()
	setupGitOpsTree(t, fs, tempDir, true)
	s, err := MakeScript("", "cicd")
	assertNoError(t, err)

	want := logsWithArgoCD
	got := executeScript(t, fs, tempDir, s)
	if got != want {
		t.Fatalf("makeScript() failed: got \n%s want: \n%s", got, want)
	}
}

func TestMakeScriptWithoutArgo(t *testing.T) {
	tempDir, cleanup := tempDir(t)
	defer cleanup()

	fs := ioutils.NewFilesystem()
	setupGitOpsTree(t, fs, tempDir, false)
	s, err := MakeScript("", "cicd")
	assertNoError(t, err)

	want := logsWithoutArgoCD
	got := executeScript(t, fs, tempDir, s)
	if got != want {
		t.Fatalf("makeScript() failed: got \n%s want: \n%s", got, want)
	}
}

func setupGitOpsTree(t *testing.T, fs afero.Fs, base string, withArgoCD bool) {
	t.Helper()
	// minimal resources to have a valid GitOps tree
	files := res.Resources{
		"environments/dev/env/overlays/kustomization.yaml":   res.Kustomization{Bases: []string{"../base"}},
		"environments/dev/apps/taxi/kustomization.yaml":      res.Kustomization{Bases: []string{"../overlays"}},
		"environments/stage/env/overlays/kustomization.yaml": res.Kustomization{Bases: []string{"../base"}},
		"environments/stage/apps/go-app/kustomization.yaml":  res.Kustomization{Bases: []string{"../overlays"}},
		"config/cicd/base/kustomization.yaml":                res.Kustomization{Resources: []string{"task.yaml"}},
		"config/cicd/overlays/kustomization.yaml":            res.Kustomization{Bases: []string{"../base"}},
		"config/cicd/base/task.yaml":                         tasks.CreateDeployUsingKubectlTask("cicd"),
	}
	if withArgoCD {
		argoDir := res.Resources{
			"config/argocd/config/kustomization.yaml": res.Kustomization{Resources: []string{"argo.yaml"}},
			"config/argocd/config/argo.yaml":          namespaces.Create("argo"),
		}
		files = res.Merge(argoDir, files)
	}
	_, err := yaml.WriteResources(fs, base, files)
	assertNoError(t, err)
}

func executeScript(t *testing.T, fs afero.Fs, baseDir, script string) string {
	t.Helper()
	scriptPath := filepath.Join(baseDir, "dryrun_script.sh")
	err := afero.WriteFile(fs, scriptPath, []byte(script), 0777)
	assertNoError(t, err)
	cmd := exec.Command(scriptPath)
	cmd.Dir = baseDir
	out, err := cmd.CombinedOutput()
	assertNoError(t, err)
	return string(out)
}

func tempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "gnome")
	assertNoError(t, err)
	return dir, func() {
		assertNoError(t, os.RemoveAll(dir))
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
