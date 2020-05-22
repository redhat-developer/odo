package eventlisteners

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/scm"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				{
					Name: "ci-dryrun-from-pr",
					Interceptors: []*triggersv1.EventInterceptor{
						{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == 'org/test'",
							},
						},
						{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: "test",
									SecretKey:  WebhookSecretKey,
									Namespace:  "testing",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						{
							Name: "github-pr-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "ci-dryrun-from-pr-template",
					},
				},
				{
					Name: "cd-deploy-from-push",
					Interceptors: []*triggersv1.EventInterceptor{
						{
							CEL: &triggersv1.CELInterceptor{
								Filter: "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == 'org/test') && body.ref.startsWith('refs/heads/master')",
							},
						},
						{
							GitHub: &triggersv1.GitHubInterceptor{
								SecretRef: &triggersv1.SecretRef{
									SecretName: "test",
									SecretKey:  WebhookSecretKey,
									Namespace:  "testing",
								},
							},
						},
					},
					Bindings: []*triggersv1.EventListenerBinding{
						{
							Name: "github-push-binding",
						},
					},
					Template: triggersv1.EventListenerTemplate{
						Name: "cd-deploy-from-push-template",
					},
				},
			},
		},
	}
	repo, err := scm.NewRepository("http://github.com/org/test")
	if err != nil {
		t.Fatal(err)
	}
	eventListener := Generate(repo, "testing", "pipeline", "test")
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
