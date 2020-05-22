package pipelines

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

const (
	elPatchFile     = "eventlistener_patch.yaml"
	elPatchDir      = "eventlistener_patches"
	rolebindingFile = "edit-rolebinding.yaml"
)

type patchStringValue struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type tektonBuilder struct {
	files      res.Resources
	gitOpsRepo string
	triggers   []v1alpha1.EventListenerTrigger
}

func buildEventListenerResources(gitOpsRepo string, m *config.Manifest) (res.Resources, error) {
	if gitOpsRepo == "" {
		return res.Resources{}, nil
	}
	cicd, err := m.GetCICDEnvironment()
	if err != nil {
		return nil, err
	}
	if cicd == nil {
		return nil, nil
	}
	files := make(res.Resources)
	tb := &tektonBuilder{files: files, gitOpsRepo: gitOpsRepo}
	err = m.Walk(tb)
	return tb.files, err
}

func (tk *tektonBuilder) Service(env *config.Environment, svc *config.Service) error {
	if svc.SourceURL == "" {
		return nil
	}
	repo, err := scm.NewRepository(svc.SourceURL)
	if err != nil {
		return err
	}
	pipelines := getPipelines(env, svc)
	ciTrigger := repo.CreateCITrigger(triggerName(svc.Name), svc.Webhook.Secret.Name, svc.Webhook.Secret.Namespace, pipelines.Integration.Template, pipelines.Integration.Bindings)
	tk.triggers = append(tk.triggers, ciTrigger)
	return nil
}

func (tk *tektonBuilder) Environment(env *config.Environment) error {
	if env.IsCICD {
		triggers, err := createTriggersForCICD(tk.gitOpsRepo, env)
		if err != nil {
			return err
		}
		tk.triggers = append(tk.triggers, triggers...)
		cicdPath := config.PathForEnvironment(env)
		tk.files[getEventListenerPath(cicdPath)] = eventlisteners.CreateELFromTriggers(env.Name, saName, tk.triggers)
	}
	return nil
}

func getEventListenerPath(cicdPath string) string {
	return filepath.Join(cicdPath, "base", "pipelines", eventListenerPath)
}

func createTriggersForCICD(gitOpsRepo string, env *config.Environment) ([]v1alpha1.EventListenerTrigger, error) {
	triggers := []v1alpha1.EventListenerTrigger{}
	repo, err := scm.NewRepository(gitOpsRepo)
	if err != nil {
		return []v1alpha1.EventListenerTrigger{}, err
	}
	_, prBindingName := repo.CreatePRBinding(env.Name)
	ciTrigger := repo.CreateCITrigger("ci-dryrun-from-pr", eventlisteners.GitOpsWebhookSecret, env.Name, "ci-dry-run-from-pr-template", []string{prBindingName})
	_, pushBindingName := repo.CreatePushBinding(env.Name)
	cdTrigger := repo.CreateCDTrigger("cd-deploy-from-push", eventlisteners.GitOpsWebhookSecret, env.Name, "cd-deploy-from-push-template", []string{pushBindingName})
	triggers = append(triggers, ciTrigger, cdTrigger)
	return triggers, nil
}

func getPipelines(env *config.Environment, svc *config.Service) *config.Pipelines {
	pipelines := clonePipelines(defaultPipelines)
	if env.Pipelines != nil {
		pipelines = clonePipelines(env.Pipelines)
	}
	if svc.Pipelines != nil {
		if len(svc.Pipelines.Integration.Bindings) > 0 {
			pipelines.Integration.Bindings = svc.Pipelines.Integration.Bindings[:]
		}
		if svc.Pipelines.Integration.Template != "" {
			pipelines.Integration.Template = svc.Pipelines.Integration.Template
		}
	}
	return pipelines
}

func clonePipelines(p *config.Pipelines) *config.Pipelines {
	return &config.Pipelines{
		Integration: &config.TemplateBinding{
			Bindings: p.Integration.Bindings[:],
			Template: p.Integration.Template,
		},
	}
}

func extractRepo(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	parts := strings.Split(parsed.Path, "/")
	return fmt.Sprintf("%s/%s", parts[1], strings.TrimSuffix(parts[2], ".git")), nil
}

func triggerName(svc string) string {
	return fmt.Sprintf("app-ci-build-from-pr-%s", svc)
}
