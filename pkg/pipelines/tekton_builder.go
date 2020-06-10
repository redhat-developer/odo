package pipelines

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/pipelines/config"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	res "github.com/openshift/odo/pkg/pipelines/resources"
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

type tektonBuilder struct {
	files      res.Resources
	gitOpsRepo string
	triggers   []v1alpha1.EventListenerTrigger
}

func buildEventListenerResources(gitOpsRepo string, m *config.Manifest) (res.Resources, error) {
	if gitOpsRepo == "" {
		return res.Resources{}, nil
	}
	cfg := m.GetPipelinesConfig()
	if cfg == nil {
		return nil, nil
	}
	files := make(res.Resources)
	tb := &tektonBuilder{files: files, gitOpsRepo: gitOpsRepo}
	triggers, err := createTriggersForCICD(tb.gitOpsRepo, cfg)
	if err != nil {
		return nil, err
	}
	tb.triggers = append(tb.triggers, triggers...)
	err = m.Walk(tb)
	if err != nil {
		return nil, err
	}
	cicdPath := config.PathForPipelines(cfg)
	files[getEventListenerPath(cicdPath)] = eventlisteners.CreateELFromTriggers(cfg.Name, saName, tb.triggers)
	return files, nil
}

func (tb *tektonBuilder) Service(env *config.Environment, svc *config.Service) error {
	if svc.SourceURL == "" {
		return nil
	}
	repo, err := scm.NewRepository(svc.SourceURL)
	if err != nil {
		return err
	}
	pipelines := getPipelines(env, svc, repo)
	ciTrigger := repo.CreateCITrigger(triggerName(svc.Name), svc.Webhook.Secret.Name, svc.Webhook.Secret.Namespace, pipelines.Integration.Template, pipelines.Integration.Bindings)
	tb.triggers = append(tb.triggers, ciTrigger)
	return nil
}

func getEventListenerPath(cicdPath string) string {
	return filepath.Join(cicdPath, "base", "pipelines", eventListenerPath)
}

func createTriggersForCICD(gitOpsRepo string, cfg *config.PipelinesConfig) ([]v1alpha1.EventListenerTrigger, error) {
	triggers := []v1alpha1.EventListenerTrigger{}
	repo, err := scm.NewRepository(gitOpsRepo)
	if err != nil {
		return []v1alpha1.EventListenerTrigger{}, err
	}
	ciTrigger := repo.CreateCITrigger("ci-dryrun-from-pr", eventlisteners.GitOpsWebhookSecret, cfg.Name, "ci-dryrun-from-pr-template", []string{repo.PRBindingName()})
	cdTrigger := repo.CreateCDTrigger("cd-deploy-from-push", eventlisteners.GitOpsWebhookSecret, cfg.Name, "cd-deploy-from-push-template", []string{repo.PushBindingName()})
	triggers = append(triggers, ciTrigger, cdTrigger)
	return triggers, nil
}

func getPipelines(env *config.Environment, svc *config.Service, r scm.Repository) *config.Pipelines {
	pipelines := defaultPipelines(r)
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

func triggerName(svc string) string {
	return fmt.Sprintf("app-ci-build-from-pr-%s", svc)
}
