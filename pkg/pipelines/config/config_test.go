package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestManifestWalk(t *testing.T) {
	m := &Manifest{
		Config: &Config{
			Pipelines: &PipelinesConfig{
				Name: "cicd",
			},
			ArgoCD: &ArgoCDConfig{
				Namespace: "argocd",
			},
		},
		Environments: []*Environment{
			{
				Name: "development",
				Services: []*Service{
					{Name: "app-1-service-http"},
					{Name: "app-1-service-test"},
					{Name: "app-2-service"},
				},
				Apps: []*Application{
					{
						Name: "my-app-1",
						ServiceRefs: []string{
							"app-1-service-http",
							"app-1-service-test",
						},
					},
					{
						Name: "my-app-2",
						ServiceRefs: []string{
							"app-2-service",
						},
					},
				},
			},
			{
				Name: "staging",
				Services: []*Service{
					{Name: "app-1-service-user"},
				},
				Apps: []*Application{
					{Name: "my-app-1",
						ServiceRefs: []string{
							"app-1-service-user",
						},
					},
				},
			},
		},
	}

	v := &testVisitor{paths: []string{}}
	err := m.Walk(v)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(v.paths)

	want := []string{
		"development/app-1-service-http",
		"development/app-1-service-test",
		"development/app-2-service",
		"development/my-app-1",
		"development/my-app-2",
		"envs/development",
		"envs/staging",
		"staging/app-1-service-user",
		"staging/my-app-1",
	}

	if diff := cmp.Diff(want, v.paths); diff != "" {
		t.Fatalf("tree files: %s", diff)
	}
}

func TestManifestWalkCalls(t *testing.T) {
	m := &Manifest{
		Environments: []*Environment{
			{
				Name: "development",

				Services: []*Service{
					{Name: "app-1-service-http"},
					{Name: "app-1-service-test"},
					{Name: "app-2-service"},
				},
				Apps: []*Application{
					{
						Name: "my-app-1",
						ServiceRefs: []string{
							"app-1-service-http",
							"app-1-service-test",
						},
					},
					{
						Name:        "my-app-2",
						ServiceRefs: []string{"app-2-service"},
					},
				},
			},
			{
				Name: "staging",
				Services: []*Service{
					{Name: "app-1-service-user"},
				},
				Apps: []*Application{
					{Name: "my-app-1",
						ServiceRefs: []string{
							"app-1-service-user",
						},
					},
				},
			},
		},
	}

	v := &testVisitor{paths: []string{}}
	err := m.Walk(v)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"development/app-1-service-http",
		"development/app-1-service-test",
		"development/app-2-service",
		"development/my-app-1",
		"development/my-app-2",
		"envs/development",
		"staging/app-1-service-user",
		"staging/my-app-1",
		"envs/staging",
	}

	if diff := cmp.Diff(want, v.paths); diff != "" {
		t.Fatalf("tree files: %s", diff)
	}
}

func TestGetPipelinesConfig(t *testing.T) {
	cfg := &Config{
		Pipelines: &PipelinesConfig{
			Name: "cicd",
		},
	}

	envTests := []struct {
		name     string
		manifest *Manifest
		want     *PipelinesConfig
	}{
		{
			name:     "manifest with configuration",
			manifest: &Manifest{Config: cfg},
			want:     cfg.Pipelines,
		},
		{
			name:     "manifest with no configuration",
			manifest: &Manifest{},
			want:     nil,
		},
	}

	for i, tt := range envTests {
		t.Run(fmt.Sprintf("test %d", i), func(rt *testing.T) {
			m := tt.manifest
			got := m.GetPipelinesConfig()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("%s: configuration did not match:\n%s", tt.name, diff)
			}
		})
	}
}
func TestGetArgoCDConfig(t *testing.T) {
	cfg := &Config{
		ArgoCD: &ArgoCDConfig{
			Namespace: "argocd",
		},
	}

	envTests := []struct {
		name     string
		manifest *Manifest
		want     *ArgoCDConfig
	}{
		{
			name:     "manifest with configuration",
			manifest: &Manifest{Config: cfg},
			want:     cfg.ArgoCD,
		},
		{
			name:     "manifest with no configuration",
			manifest: &Manifest{},
			want:     nil,
		},
	}

	for i, tt := range envTests {
		t.Run(fmt.Sprintf("test %d", i), func(rt *testing.T) {
			m := tt.manifest
			got := m.GetArgoCDConfig()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("%s: configuration did not match:\n%s", tt.name, diff)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	m := &Manifest{Environments: makeEnvs([]testEnv{{name: "prod"}, {name: "testing"}})}
	env := m.GetEnvironment("prod")
	if env.Name != "prod" {
		t.Fatalf("got the wrong environment back: %#v", env)
	}

	unknown := m.GetEnvironment("unknown")
	if unknown != nil {
		t.Fatalf("found an unknown env: %#v", unknown)
	}
}

func makeEnvs(ns []testEnv) []*Environment {
	n := make([]*Environment, len(ns))
	for i, v := range ns {
		n[i] = &Environment{Name: v.name}
	}
	return n

}

type testEnv struct {
	name string
}

type testVisitor struct {
	pipelineServices []string
	paths            []string
}

func (v *testVisitor) Service(env *Environment, svc *Service) error {
	v.paths = append(v.paths, filepath.Join(env.Name, svc.Name))
	v.pipelineServices = append(v.pipelineServices, filepath.Join("cicd", env.Name, svc.Name))
	return nil
}

func (v *testVisitor) Application(env *Environment, app *Application) error {
	v.paths = append(v.paths, filepath.Join(env.Name, app.Name))
	return nil
}

func (v *testVisitor) Environment(env *Environment) error {
	if env.Name == "cicd" {
		v.paths = append(v.paths, v.pipelineServices...)
	}
	v.paths = append(v.paths, filepath.Join("envs", env.Name))
	return nil
}
