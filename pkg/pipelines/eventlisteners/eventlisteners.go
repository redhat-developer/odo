package eventlisteners

import (
	"fmt"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Filters for interceptors
const (
	devCIBuildFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == '%s'"

	devCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == '%s') && body.ref.startsWith('refs/heads/master')"

	stageCIDryRunFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == '%s'"

	stageCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == '%s') && body.ref.startsWith('refs/heads/master')"
)

// Generate will create the required eventlisteners.
func Generate(githubRepo, ns, saName string) triggersv1.EventListener {
	githubStageRepo := githubRepo + "-stage-config"
	return triggersv1.EventListener{
		TypeMeta:   createListenerTypeMeta(),
		ObjectMeta: createListenerObjectMeta("cicd-event-listener", ns),
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers: []triggersv1.EventListenerTrigger{
				createListenerTrigger(
					"dev-ci-build-from-pr",
					devCIBuildFilters,
					githubRepo,
					"github-pr-binding",
					"dev-ci-build-from-pr-template",
				),
				createListenerTrigger(
					"dev-cd-deploy-from-master",
					devCDDeployFilters,
					githubRepo,
					"github-push-binding",
					"dev-cd-deploy-from-master-template",
				),
				createListenerTrigger(
					"stage-ci-dryrun-from-pr",
					stageCIDryRunFilters,
					githubStageRepo,
					"github-pr-binding",
					"stage-ci-dryrun-from-pr-template",
				),
				createListenerTrigger(
					"stage-cd-deploy-from-push",
					stageCDDeployFilters,
					githubStageRepo,
					"github-push-binding",
					"stage-cd-deploy-from-push-template",
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

func createListenerTrigger(name string, filter string, repoName string, binding string, template string) triggersv1.EventListenerTrigger {
	return triggersv1.EventListenerTrigger{
		Name: name,
		Interceptors: []*triggersv1.EventInterceptor{
			createEventInterceptor(filter, repoName),
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

func createListenerTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "EventListener",
		APIVersion: "tekton.dev/v1alpha1",
	}
}

func createListenerObjectMeta(name, ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
}
