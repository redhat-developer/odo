package pipelines

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

func TestBuildEventListener(t *testing.T) {
	m := &config.Manifest{
		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "test-cicd",
			},
		},
		Environments: []*config.Environment{
			testEnv(testService(), "dev"),
			testEnv(testService(), "staging"),
		},
	}
	cicdPath := filepath.Join("config", "test-cicd")
	gitOpsRepo := "http://github.com/org/gitops.git"
	got, err := buildEventListenerResources(gitOpsRepo, m)
	assertNoError(t, err)
	want := res.Resources{
		getEventListenerPath(cicdPath): eventlisteners.CreateELFromTriggers("test-cicd", saName, fakeTriggers(t, m, gitOpsRepo)),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("resources didn't match:%s\n", diff)
	}
}

func TestBuildEventListenerWithServiceWithNoURL(t *testing.T) {
	m := &config.Manifest{

		Config: &config.Config{
			Pipelines: &config.PipelinesConfig{
				Name: "test-cicd",
			},
		},
		Environments: []*config.Environment{
			testEnv(testService(), "dev"),
		},
	}
	cicdPath := filepath.Join("config", "test-cicd")
	gitOpsRepo := "http://github.com/org/gitops.git"
	got, err := buildEventListenerResources(gitOpsRepo, m)
	assertNoError(t, err)
	want := res.Resources{
		getEventListenerPath(cicdPath): eventlisteners.CreateELFromTriggers("test-cicd", saName, fakeTriggers(t, m, gitOpsRepo)),
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("resources didn't match:%s\n", diff)
	}
}

func TestBuildEventListenerWithNoGitOpsURL(t *testing.T) {
	m := &config.Manifest{
		Environments: []*config.Environment{
			{
				Name: "test-cicd",
			},
			testEnv(testService(), "dev"),
		},
	}
	got, err := buildEventListenerResources("", m)
	assertNoError(t, err)

	want := res.Resources{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("resources didn't match:%s\n", diff)
	}
}

func TestGetPipelines(t *testing.T) {
	tests := []struct {
		desc string
		env  *config.Environment
		svc  *config.Service
		want *config.Pipelines
	}{
		{
			"Pipelines are provided by environment",
			&config.Environment{
				Name:      "test-env",
				Pipelines: testPipelines("env"),
			},
			&config.Service{
				Name: "test-svc",
			},
			testPipelines("env"),
		},
		{
			"Pipelines are provided by service",
			&config.Environment{
				Name: "test-env",
			},
			&config.Service{
				Name:      "test-service",
				Pipelines: testPipelines("svc"),
			},
			testPipelines("svc"),
		},
		{
			"Default pipelines are used",
			&config.Environment{
				Name: "test-env",
			},
			&config.Service{
				Name: "test-service",
			},
			&config.Pipelines{
				Integration: &config.TemplateBinding{
					Template: "app-ci-template",
					Bindings: []string{"github-pr-binding"},
				},
			},
		},
		{
			"Only override the bindings in the service",
			&config.Environment{
				Name:      "test-env",
				Pipelines: testPipelines("env"),
			},
			&config.Service{
				Name: "test-service",
				Pipelines: &config.Pipelines{
					Integration: &config.TemplateBinding{
						Bindings: []string{"svc-ci-binding"},
					},
				},
			},
			&config.Pipelines{
				Integration: &config.TemplateBinding{
					Template: "env-ci-template",
					Bindings: []string{"svc-ci-binding"},
				},
			},
		},
		{
			"Only override the template in the service",
			&config.Environment{
				Name:      "test-env",
				Pipelines: testPipelines("env"),
			},
			&config.Service{
				Name: "test-service",
				Pipelines: &config.Pipelines{
					Integration: &config.TemplateBinding{
						Template: "svc-ci-template",
					},
				},
			},
			&config.Pipelines{
				Integration: &config.TemplateBinding{
					Template: "svc-ci-template",
					Bindings: []string{"env-ci-binding"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(rt *testing.T) {
			var envPipelines *config.Pipelines
			if test.env.Pipelines != nil {
				envPipelines = clonePipelines(test.env.Pipelines)
			}
			repo, _ := scm.NewRepository("https://github.com/foo/bar")
			got := getPipelines(test.env, test.svc, repo)
			if diff := cmp.Diff(test.want, got); diff != "" {
				rt.Errorf("getPipelines() failed:\n%v", diff)
			}
			if diff := cmp.Diff(envPipelines, test.env.Pipelines); diff != "" {
				rt.Errorf("environment pipelines overwritten: %s\n", diff)
			}
		})
	}
}

func fakeTriggers(t *testing.T, m *config.Manifest, gitOpsRepo string) []triggersv1.EventListenerTrigger {
	triggers := []triggersv1.EventListenerTrigger{}
	cfg := m.GetPipelinesConfig()
	cicdTriggers, err := createTriggersForCICD(gitOpsRepo, cfg)
	assertNoError(t, err)
	triggers = append(triggers, cicdTriggers...)
	for _, env := range m.Environments {
		svc := testService()
		repo, err := scm.NewRepository(svc.SourceURL)
		assertNoError(t, err)
		pipelines := getPipelines(env, svc, repo)
		devCITrigger := repo.CreateCITrigger(fmt.Sprintf("app-ci-build-from-pr-%s", svc.Name), svc.Webhook.Secret.Name, svc.Webhook.Secret.Namespace, pipelines.Integration.Template, pipelines.Integration.Bindings)
		triggers = append(triggers, devCITrigger)
	}

	return triggers
}

func testService() *config.Service {
	return &config.Service{
		Name:      "test-svc",
		SourceURL: "http://github.com/org/test.git",
		Webhook: &config.Webhook{
			Secret: &config.Secret{
				Name:      "webhook-secret",
				Namespace: "webhook-ns",
			},
		},
	}
}

func testEnv(svc *config.Service, name string) *config.Environment {
	return &config.Environment{
		Name:      "test-" + name,
		Pipelines: testPipelines("test"),
		Services: []*config.Service{
			svc,
		},
		Apps: []*config.Application{
			{
				Name: "test-" + name + "-app",
			},
		},
	}
}

func testPipelines(name string) *config.Pipelines {
	return &config.Pipelines{
		Integration: &config.TemplateBinding{
			Template: fmt.Sprintf("%s-ci-template", name),
			Bindings: []string{fmt.Sprintf("%s-ci-binding", name)},
		},
	}
}
