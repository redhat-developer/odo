package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestManifestWalk(t *testing.T) {
	m := &Manifest{
		Environments: []*Environment{
			{
				Name: "development",
				Apps: []*Application{
					{
						Name: "my-app-1",
						Services: []*Service{
							{Name: "app-1-service-http"},
							{Name: "app-1-service-test"},
						},
					},
					{
						Name: "my-app-2",
						Services: []*Service{
							{Name: "app-2-service"},
						},
					},
				},
			},
			{
				Name: "staging",
				Apps: []*Application{
					{Name: "my-app-1",
						Services: []*Service{
							{Name: "app-1-service-user"},
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
		"development/my-app-1",
		"development/my-app-1/app-1-service-http",
		"development/my-app-1/app-1-service-test",
		"development/my-app-2",
		"development/my-app-2/app-2-service",
		"envs/development",
		"envs/staging",
		"staging/my-app-1",
		"staging/my-app-1/app-1-service-user",
	}

	if diff := cmp.Diff(want, v.paths); diff != "" {
		t.Fatalf("tree files: %s", diff)
	}
}

func TestManifestWalkCallsCICDEnvironmentLast(t *testing.T) {
	m := &Manifest{
		Environments: []*Environment{
			{
				Name:   "cicd",
				IsCICD: true,
			},
			{
				Name: "development",
				Apps: []*Application{
					{
						Name: "my-app-1",
						Services: []*Service{
							{Name: "app-1-service-http"},
							{Name: "app-1-service-test"},
						},
					},
					{
						Name: "my-app-2",
						Services: []*Service{
							{Name: "app-2-service"},
						},
					},
				},
			},
			{
				Name: "staging",
				Apps: []*Application{
					{Name: "my-app-1",
						Services: []*Service{
							{Name: "app-1-service-user"},
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
		"development/my-app-1/app-1-service-http",
		"development/my-app-1/app-1-service-test",
		"development/my-app-1",
		"development/my-app-2/app-2-service",
		"development/my-app-2",
		"envs/development",
		"staging/my-app-1/app-1-service-user",
		"staging/my-app-1",
		"envs/staging",
		"cicd/development/app-1-service-http",
		"cicd/development/app-1-service-test",
		"cicd/development/app-2-service",
		"cicd/staging/app-1-service-user",
		"envs/cicd",
	}

	if diff := cmp.Diff(want, v.paths); diff != "" {
		t.Fatalf("tree files: %s", diff)
	}
}

func TestEnviromentSorting(t *testing.T) {
	envNames := func(envs []*Environment) []string {
		n := make([]string, len(envs))
		for i, v := range envs {
			n[i] = v.Name
		}
		return n
	}
	envTests := []struct {
		names []testEnv
		want  []string
	}{
		{[]testEnv{{"prod", false, false}, {"staging", false, false}, {"dev", false, false}}, []string{"dev", "prod", "staging"}},
		{[]testEnv{{"cicd", true, false}, {"staging", false, false}, {"dev", false, false}}, []string{"dev", "staging", "cicd"}},
		{[]testEnv{{"m-cicd", true, false}, {"staging", false, false}, {"dev", false, false}}, []string{"dev", "staging", "m-cicd"}},
		{[]testEnv{{"m-cicd", true, false}, {"argo", false, true}, {"dev", false, false}}, []string{"dev", "m-cicd", "argo"}},
		{[]testEnv{{"m-argo", false, true}, {"testing", false, false}, {"dev", false, false}}, []string{"dev", "testing", "m-argo"}},
	}

	for _, tt := range envTests {
		envs := makeEnvs(tt.names)
		sort.Sort(ByName(envs))
		if diff := cmp.Diff(tt.want, envNames(envs)); diff != "" {
			t.Errorf("sort(%#v): %s", envs, diff)
		}
	}
}

func TestFindCICDEnviroment(t *testing.T) {
	envTests := []struct {
		names []testEnv
		want  string
		err   string
	}{
		{[]testEnv{{"prod", false, false}, {"staging", false, false}, {"dev", false, false}}, "", ""},
		{[]testEnv{{"test-cicd", true, false}, {"staging", false, false}, {"dev", false, false}}, "test-cicd", ""},
		{[]testEnv{{"test-cicd", true, false}, {"oc-cicd", true, false}, {"dev", false, false}}, "", "found multiple CI/CD environments"},
	}

	for i, tt := range envTests {
		t.Run(fmt.Sprintf("test %d", i), func(rt *testing.T) {
			m := &Manifest{Environments: makeEnvs(tt.names)}
			env, err := m.GetCICDEnvironment()
			if !matchErrorString(t, tt.err, err) {
				rt.Errorf("did not match error, got %s, want %s", err, tt.err)
				return
			}

			if tt.want != "" && (env.Name != tt.want) {
				rt.Errorf("found incorrect CICD environment, got %s, want %s", env.Name, tt.want)
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
		n[i] = &Environment{Name: v.name, IsCICD: v.cicd, IsArgoCD: v.argocd}
	}
	return n

}

type testEnv struct {
	name   string
	cicd   bool
	argocd bool
}

type testVisitor struct {
	pipelineServices []string
	paths            []string
}

func (v *testVisitor) Service(env *Environment, app *Application, svc *Service) error {
	v.paths = append(v.paths, filepath.Join(env.Name, app.Name, svc.Name))
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

// MatchErrorString takes a string and matches on the error and returns true if
// the
// string matches the error.
//
// This is useful in table tests.
//
// If the string can't be compiled as an regexp, then this will fail with a
// Fatal error.
func matchErrorString(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}
