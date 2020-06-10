package environments

import (
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/openshift/odo/pkg/pipelines/namespaces"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/spf13/afero"
)

func TestBuildEnvironmentFilesWithAppsToEnvironment(t *testing.T) {
	var appFs = ioutils.NewMapFilesystem()
	m := buildManifest(true)

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{Bases: []string{
			"../../../services/service-http",
			"../../../services/service-metrics",
			"../../../env/base"}},
		"environments/test-dev/apps/my-app-1/kustomization.yaml":                     &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml":            &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":                   namespaces.Create("test-dev"),
		"environments/test-dev/env/base/test-dev-rolebinding.yaml":                   createRoleBinding(m.Environments[0], "environments/test-dev/env/base", "cicd", "pipelines"),
		"environments/test-dev/env/base/kustomization.yaml":                          &res.Kustomization{Resources: []string{"test-dev-environment.yaml", "test-dev-rolebinding.yaml"}},
		"environments/test-dev/env/overlays/kustomization.yaml":                      &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentFilesWithEnvironmentsToApps(t *testing.T) {
	var appFs = ioutils.NewMapFilesystem()
	m := buildManifest(true)

	files, err := Build(appFs, m, "pipelines", EnvironmentsToApps)
	if err != nil {
		t.Fatal(err)
	}
	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{Bases: []string{
			"../../../services/service-http",
			"../../../services/service-metrics"}},
		"environments/test-dev/apps/my-app-1/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":        namespaces.Create("test-dev"),
		"environments/test-dev/env/base/test-dev-rolebinding.yaml":        createRoleBinding(m.Environments[0], "environments/test-dev/env/base", "cicd", "pipelines"),
		"environments/test-dev/env/base/kustomization.yaml": &res.Kustomization{
			Resources: []string{"test-dev-environment.yaml", "test-dev-rolebinding.yaml"},
			Bases:     []string{"../../apps/my-app-1/overlays"},
		},
		"environments/test-dev/env/overlays/kustomization.yaml":                      &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentsDoesNotOutputCIorArgo(t *testing.T) {
	var appFs = ioutils.NewMapFilesystem()
	m := &config.Manifest{
		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "cicd",
			},
			ArgoCD: &config.ArgoCDConfig{
				Namespace: "argocd",
			},
		},
	}

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{}
	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func mustWriteFile(t *testing.T, fs afero.Fs, path string, data []byte, perm os.FileMode) {
	t.Helper()
	err := afero.WriteFile(fs, path, data, perm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildEnvironmentsAddsKustomizedFiles(t *testing.T) {
	var appFs = ioutils.NewMapFilesystem()
	err := appFs.MkdirAll("environments/test-dev/base", 0755)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, appFs, "environments/test-dev/base/volume.yaml", []byte(`this is a file`), 0644)
	mustWriteFile(t, appFs, "environments/test-dev/base/test-dev-environment.yaml", []byte(`this is a file`), 0644)
	mustWriteFile(t, appFs, "environments/test-dev/base/routes/01-route.yaml", []byte(`this is a file`), 0644)

	m := &config.Manifest{
		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "cicd",
			},
		},
		Environments: []*config.Environment{
			{Name: "test-dev"},
		},
	}

	resources, err := Build(appFs, m, "pipelines", EnvironmentsToApps)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"environments/test-dev/env/base/kustomization.yaml",
		"environments/test-dev/env/base/test-dev-environment.yaml",
		"environments/test-dev/env/overlays/kustomization.yaml",
	}
	sort.Strings(want)

	if diff := cmp.Diff(want, filesFromResources(resources)); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func TestBuildEnvironmentFilesWithNoCICDEnv(t *testing.T) {
	var appFs = ioutils.NewMapFilesystem()
	m := buildManifest(false)

	files, err := Build(appFs, m, "pipelines", AppsToEnvironments)
	if err != nil {
		t.Fatal(err)
	}

	want := res.Resources{
		"environments/test-dev/apps/my-app-1/base/kustomization.yaml": &res.Kustomization{Bases: []string{
			"../../../services/service-http",
			"../../../services/service-metrics",
			"../../../env/base"}},
		"environments/test-dev/apps/my-app-1/kustomization.yaml":                     &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/apps/my-app-1/overlays/kustomization.yaml":            &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/env/base/test-dev-environment.yaml":                   namespaces.Create("test-dev"),
		"environments/test-dev/env/base/kustomization.yaml":                          &res.Kustomization{Resources: []string{"test-dev-environment.yaml"}},
		"environments/test-dev/env/overlays/kustomization.yaml":                      &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-http/kustomization.yaml":             &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-http/base/kustomization.yaml":        &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-http/overlays/kustomization.yaml":    &res.Kustomization{Bases: []string{"../base"}},
		"environments/test-dev/services/service-metrics/kustomization.yaml":          &res.Kustomization{Bases: []string{"overlays"}},
		"environments/test-dev/services/service-metrics/base/kustomization.yaml":     &res.Kustomization{Bases: []string{"./config"}},
		"environments/test-dev/services/service-metrics/overlays/kustomization.yaml": &res.Kustomization{Bases: []string{"../base"}},
	}

	if diff := cmp.Diff(want, files); diff != "" {
		t.Fatalf("files didn't match: %s\n", diff)
	}
}

func filesFromResources(r res.Resources) []string {
	names := []string{}
	for k := range r {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func buildManifest(withCICD bool) *config.Manifest {
	if withCICD {
		return &config.Manifest{
			Config: &config.Config{
				Pipelines: &config.PipelinesConfig{
					Name: "cicd",
				},
			},
			Environments: createEnvironment(),
		}
	}
	return &config.Manifest{
		Environments: createEnvironment(),
	}

}

func createEnvironment() []*config.Environment {
	return []*config.Environment{
		{
			Name: "test-dev",
			Apps: []*config.Application{
				{
					Name: "my-app-1",
					ServiceRefs: []string{
						"service-http",
						"service-metrics",
					},
				},
			},
			Services: []*config.Service{
				{
					Name:      "service-http",
					SourceURL: "https://github.com/myproject/myservice.git",
				},
				{
					Name: "service-metrics",
				},
			},
		},
	}
}
