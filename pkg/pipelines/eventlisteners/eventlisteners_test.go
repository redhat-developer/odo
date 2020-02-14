package eventlisteners

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateEventListener(t *testing.T) {
	validEventListener := triggersv1.EventListener{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventListener",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "cicd-event-listener",
		},
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: "demo-sa",
			Triggers: []triggersv1.EventListenerTrigger{
				triggersv1.EventListenerTrigger{
					Name: "dev-ci-build-from-pr",
					Interceptors: []*triggersv1.EventInterceptor{
						&triggersv1.EventInterceptor{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'GITHUB_REPO'",
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "dev-ci-build-from-pr-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "dev-ci-build-from-pr-template",
					},
				},
				triggersv1.EventListenerTrigger{
					Name: "dev-cd-deploy-from-master",
					Interceptors: []*triggersv1.EventInterceptor{
						&triggersv1.EventInterceptor{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'GITHUB_REPO') && body.ref.startsWith('refs/heads/master')",
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "dev-cd-deploy-from-master-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "dev-cd-deploy-from-master-template",
					},
				},
				triggersv1.EventListenerTrigger{
					Name: "stage-ci-dryrun-from-pr",
					Interceptors: []*triggersv1.EventInterceptor{
						&triggersv1.EventInterceptor{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'GITHUB_STAGE_REPO'",
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "stage-ci-dryrun-from-pr-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "stage-ci-dryrun-from-pr-template",
					},
				},
				triggersv1.EventListenerTrigger{
					Name: "stage-cd-deploy-from-push",
					Interceptors: []*triggersv1.EventInterceptor{
						&triggersv1.EventInterceptor{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'GITHUB_STAGE_REPO') && body.ref.startsWith('refs/heads/master')",
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "stage-cd-deploy-from-push-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "stage-cd-deploy-from-push-template",
					},
				},
			},
		},
	}

	eventListener := GenerateEventListener()
	if diff := cmp.Diff(validEventListener, eventListener); diff != "" {
		t.Fatalf("GenerateEventListener() failed:\n%s", diff)
	}
}

func TestCreateListenerObjectMeta(t *testing.T) {
	validObjectMeta := metav1.ObjectMeta{
		Name: "sample",
	}
	objectMeta := createListenerObjectMeta("sample")
	if diff := cmp.Diff(validObjectMeta, objectMeta); diff != "" {
		t.Fatalf("createListenerObjectMeta() failed:\n%s", diff)
	}
}

func TestCreateListenerBinding(t *testing.T) {
	validListenerBinding := triggersv1.EventListenerBinding{
		Name: "sample",
	}
	listenerBinding := createListenerBinding("sample")
	if diff := cmp.Diff(validListenerBinding, *listenerBinding); diff != "" {
		t.Fatalf("createListenerBinding() failed:\n%s", diff)
	}
}

func TestCreateListenerTemplate(t *testing.T) {
	validListenerTemplate := triggersv1.EventListenerTemplate{
		Name: "sample",
	}
	listenerTemplate := createListenerTemplate("sample")
	if diff := cmp.Diff(validListenerTemplate, listenerTemplate); diff != "" {
		t.Fatalf("createListenerTemplate() failed:\n%s", diff)
	}
}

func TestCreateListenerTrigger(t *testing.T) {
	validListenerTrigger := triggersv1.EventListenerTrigger{
		Name: "sampleName",
		Interceptors: []*triggersv1.EventInterceptor{
			&triggersv1.EventInterceptor{
				CEL: &triggersv1.CELInterceptor{
					Filter: "sampleFilter",
				},
			},
		},
		Bindings: []*triggersv1.EventListenerBinding{
			&triggersv1.EventListenerBinding{
				Name: "sampleBindingName",
			},
		},
		Template: triggersv1.EventListenerTemplate{
			Name: "sampleTemplateName",
		},
	}
	listenerTrigger := createListenerTrigger("sampleName", "sampleFilter", "sampleBindingName", "sampleTemplateName")
	if diff := cmp.Diff(validListenerTrigger, listenerTrigger); diff != "" {
		t.Fatalf("createListenerTrigger() failed:\n%s", diff)
	}
}

func TestCreateEventInterceptor(t *testing.T) {
	validEventInterceptor := triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: "sampleFilter",
		},
	}
	eventInterceptor := createEventInterceptor("sampleFilter")
	if diff := cmp.Diff(validEventInterceptor, *eventInterceptor); diff != "" {
		t.Fatalf("createEventInterceptor() failed:\n%s", diff)
	}
}
