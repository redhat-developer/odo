package eventlisteners

import (
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Filters for interceptors
const (
	devCIBuildFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'GITHUB_REPO'"

	decCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'GITHUB_REPO') && body.ref.startsWith('refs/heads/master')"

	stageCIDryRunFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'GITHUB_STAGE_REPO'"

	stageCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'GITHUB_STAGE_REPO') && body.ref.startsWith('refs/heads/master')"
)

// GenerateEventListener will create the required eventlisteners
func GenerateEventListener() triggersv1.EventListener {
	return triggersv1.EventListener{
		TypeMeta:   createListenerTypeMeta(),
		ObjectMeta: createListenerObjectMeta("cicd-event-listener"),
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: "demo-sa",
			Triggers: []triggersv1.EventListenerTrigger{
				createListenerTrigger(
					"dev-ci-build-from-pr",
					devCIBuildFilters,
					"dev-ci-build-from-pr-binding",
					"dev-ci-build-from-pr-template",
				),
				createListenerTrigger(
					"dev-cd-deploy-from-master",
					decCDDeployFilters,
					"dev-cd-deploy-from-master-binding",
					"dev-cd-deploy-from-master-template",
				),
				createListenerTrigger(
					"stage-ci-dryrun-from-pr",
					stageCIDryRunFilters,
					"stage-ci-dryrun-from-pr-binding",
					"stage-ci-dryrun-from-pr-template",
				),
				createListenerTrigger(
					"stage-cd-deploy-from-push",
					stageCDDeployFilters,
					"stage-cd-deploy-from-push-binding",
					"stage-cd-deploy-from-push-template",
				),
			},
		},
	}
}

func createEventInterceptor(filter string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: filter,
		},
	}
}

func createListenerTrigger(name string, filter string, binding string, template string) triggersv1.EventListenerTrigger {
	return triggersv1.EventListenerTrigger{
		Name: name,
		Interceptors: []*triggersv1.EventInterceptor{
			createEventInterceptor(filter),
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

func createListenerObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}
