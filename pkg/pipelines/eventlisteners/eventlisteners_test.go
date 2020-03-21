package eventlisteners

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	sName = "pipeline"
)

func TestGenerateEventListener(t *testing.T) {
	validEventListener := triggersv1.EventListener{
		TypeMeta: eventListenerTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cicd-event-listener",
			Namespace: "testing",
		},
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: "pipeline",
			Triggers: []triggersv1.EventListenerTrigger{
				triggersv1.EventListenerTrigger{
					Name: "dev-ci-build-from-pr",
					Interceptors: []*triggersv1.EventInterceptor{
						&triggersv1.EventInterceptor{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'sample'",
							},
						},
						&triggersv1.EventInterceptor{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: GithubWebHookSecret,
									SecretKey:  WebhookSecretKey,
								},
								EventTypes: []string{
									"pull_request",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "github-pr-binding",
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
								Filter: "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'sample') && body.ref.startsWith('refs/heads/master')",
							},
						},
						&triggersv1.EventInterceptor{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: GithubWebHookSecret,
									SecretKey:  WebhookSecretKey,
								},
								EventTypes: []string{
									"push",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "github-push-binding",
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
								Filter: "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'sample-stage-config'",
							},
						},
						&triggersv1.EventInterceptor{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: GithubWebHookSecret,
									SecretKey:  WebhookSecretKey,
								},
								EventTypes: []string{
									"pull_request",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "github-pr-binding",
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
								Filter: "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'sample-stage-config') && body.ref.startsWith('refs/heads/master')",
							},
						},
						&triggersv1.EventInterceptor{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: GithubWebHookSecret,
									SecretKey:  WebhookSecretKey,
								},
								EventTypes: []string{
									"push",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						&triggersv1.EventListenerBinding{
							Name: "github-push-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "stage-cd-deploy-from-push-template",
					},
				},
			},
		},
	}

	eventListener := Generate("sample", "testing", sName)
	if diff := cmp.Diff(validEventListener, eventListener); diff != "" {
		t.Fatalf("Generate() failed:\n%s", diff)
	}
}

func TestCreateListenerObjectMeta(t *testing.T) {
	validObjectMeta := metav1.ObjectMeta{
		Name:      "sample",
		Namespace: "testing",
	}
	objectMeta := createListenerObjectMeta("sample", "testing")
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
					Filter: "sampleFilter sample",
				},
			},
			&triggersv1.EventInterceptor{
				GitHub: &triggersv1.GitHubInterceptor{
					SecretRef: &triggersv1.SecretRef{
						SecretName: GithubWebHookSecret,
						SecretKey:  WebhookSecretKey,
					},
					EventTypes: []string{
						"pull_request",
					},
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
	listenerTrigger := createListenerTrigger("sampleName", "sampleFilter %s", "sample", "sampleBindingName", "sampleTemplateName", "pull_request")
	if diff := cmp.Diff(validListenerTrigger, listenerTrigger); diff != "" {
		t.Fatalf("createListenerTrigger() failed:\n%s", diff)
	}
}

func TestCreateEventInterceptor(t *testing.T) {
	validEventInterceptor := triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: "sampleFilter sample",
		},
	}
	eventInterceptor := createEventInterceptor("sampleFilter %s", "sample")
	if diff := cmp.Diff(validEventInterceptor, *eventInterceptor); diff != "" {
		t.Fatalf("createEventInterceptor() failed:\n%s", diff)
	}
}

func TestCreateGithubInterceptor(t *testing.T) {
	validGithubInterceptor := triggersv1.EventInterceptor{
		GitHub: &triggersv1.GitHubInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: GithubWebHookSecret,
				SecretKey:  WebhookSecretKey,
			},
			EventTypes: []string{
				"pull_request",
			},
		},
	}
	githubInterceptor := createGithubInterceptor("pull_request")
	if diff := cmp.Diff(validGithubInterceptor, *githubInterceptor); diff != "" {
		t.Fatalf("createEventInterceptor() failed:\n%s", diff)
	}
}
