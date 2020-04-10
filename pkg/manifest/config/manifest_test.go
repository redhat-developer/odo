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
			&Environment{
				Name: "development",
				Apps: []*Application{
					&Application{
						Name: "my-app-1",
						Services: []*Service{
							&Service{Name: "app-1-service-http"},
							&Service{Name: "app-1-service-test"},
						},
					},
					&Application{
						Name: "my-app-2",
						Services: []*Service{
							&Service{Name: "app-2-service"},
						},
					},
				},
			},
			&Environment{
				Name: "staging",
				Apps: []*Application{
					&Application{Name: "my-app-1",
						Services: []*Service{
							&Service{Name: "app-1-service-user"},
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
			&Environment{
				Name:   "cicd",
				IsCICD: true,
			},
			&Environment{
				Name: "development",
				Apps: []*Application{
					&Application{
						Name: "my-app-1",
						Services: []*Service{
							&Service{Name: "app-1-service-http"},
							&Service{Name: "app-1-service-test"},
						},
					},
					&Application{
						Name: "my-app-2",
						Services: []*Service{
							&Service{Name: "app-2-service"},
						},
					},
				},
			},
			&Environment{
				Name: "staging",
				Apps: []*Application{
					&Application{Name: "my-app-1",
						Services: []*Service{
							&Service{Name: "app-1-service-user"},
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
		{[]testEnv{{"prod", false}, {"staging", false}, {"dev", false}}, []string{"dev", "prod", "staging"}},
		{[]testEnv{{"cicd", true}, {"staging", false}, {"dev", false}}, []string{"dev", "staging", "cicd"}},
		{[]testEnv{{"m-cicd", true}, {"staging", false}, {"dev", false}}, []string{"dev", "staging", "m-cicd"}},
	}

	for _, tt := range envTests {
		envs := makeEnvs(tt.names)
		sort.Sort(byName(envs))
		if diff := cmp.Diff(envNames(envs), tt.want); diff != "" {
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
		{[]testEnv{{"prod", false}, {"staging", false}, {"dev", false}}, "", "could not find CI/CD environment"},
		{[]testEnv{{"test-cicd", true}, {"staging", false}, {"dev", false}}, "test-cicd", ""},
		{[]testEnv{{"test-cicd", true}, {"oc-cicd", true}, {"dev", false}}, "", "found multiple CI/CD environments"},
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

func makeEnvs(ns []testEnv) []*Environment {
	n := make([]*Environment, len(ns))
	for i, v := range ns {
		n[i] = &Environment{Name: v.name, IsCICD: v.cicd}
	}
	return n

}

type testEnv struct {
	name string
	cicd bool
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
