package eventlisteners

import (
	"fmt"

	"github.com/openshift/odo/pkg/pipelines/meta"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Filters for interceptors
const (
	stageCIDryRunFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == '%s'"

	stageCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == '%s') && body.ref.startsWith('refs/heads/master')"

	GitOpsWebhookSecret = "gitops-webhook-secret"

	WebhookSecretKey = "webhook-secret-key"
)

var (
	eventListenerTypeMeta = meta.TypeMeta("EventListener", "tekton.dev/v1alpha1")
)

// Generate will create the required eventlisteners.
func Generate(githubRepo, ns, saName string) triggersv1.EventListener {
	return triggersv1.EventListener{
		TypeMeta:   eventListenerTypeMeta,
		ObjectMeta: createListenerObjectMeta("cicd-event-listener", ns),
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers: []triggersv1.EventListenerTrigger{
				createListenerTrigger(
					"ci-dryrun-from-pr",
					stageCIDryRunFilters,
					githubRepo,
					"github-pr-binding",
					"ci-dryrun-from-pr-template",
					"pull_request",
				),
				createListenerTrigger(
					"cd-deploy-from-push",
					stageCDDeployFilters,
					githubRepo,
					"github-push-binding",
					"cd-deploy-from-push-template",
					"push",
				),
			},
		},
	}
}

func createEventInterceptor(filter string, repoName string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: fmt.Sprintf(filter, repoName),
		},
	}
}

func createGithubInterceptor(eventType string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		GitHub: &triggersv1.GitHubInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: GitOpsWebhookSecret,
				SecretKey:  WebhookSecretKey,
			},
			EventTypes: []string{
				eventType,
			},
		},
	}
}

func createListenerTrigger(name string, filter string, repoName string, binding string, template string, eventType string) triggersv1.EventListenerTrigger {
	return triggersv1.EventListenerTrigger{
		Name: name,
		Interceptors: []*triggersv1.EventInterceptor{
			createEventInterceptor(filter, repoName),
			createGithubInterceptor(eventType),
		},
		Bindings: []*triggersv1.EventListenerBinding{
			createListenerBinding(binding),
		},
		Template: createListenerTemplate(template),
	}
}

func createListenerTemplate(name string) triggersv1.EventListenerTemplate {
	return triggersv1.EventListenerTemplate{
		Name: name,
	}
}

func createListenerBinding(name string) *triggersv1.EventListenerBinding {
	return &triggersv1.EventListenerBinding{
		Name: name,
	}
}

func createListenerObjectMeta(name, ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
}
