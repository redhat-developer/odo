package eventlisteners

import (
	"github.com/openshift/odo/pkg/pipelines/scm"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

// Filters for interceptors
const (
	GitOpsWebhookSecret = "gitops-webhook-secret"

	WebhookSecretKey = "webhook-secret-key"
)

var (
	eventListenerTypeMeta = meta.TypeMeta("EventListener", "triggers.tekton.dev/v1alpha1")
)

// Generate will create the required eventlisteners.
func Generate(repo scm.Repository, ns, saName, secretName string) triggersv1.EventListener {
	return triggersv1.EventListener{
		TypeMeta:   eventListenerTypeMeta,
		ObjectMeta: createListenerObjectMeta("cicd-event-listener", ns),
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers: []triggersv1.EventListenerTrigger{
				repo.CreateCITrigger("ci-dryrun-from-pr", secretName, ns, "ci-dryrun-from-pr-template",
					[]string{"github-pr-binding"}),
				repo.CreateCDTrigger("cd-deploy-from-push", secretName, ns, "cd-deploy-from-push-template", []string{"github-push-binding"}),
			},
		},
	}
}

func CreateELFromTriggers(cicdNS, saName string, triggers []triggersv1.EventListenerTrigger) *triggersv1.EventListener {
	return &v1alpha1.EventListener{
		TypeMeta: eventListenerTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cicd-event-listener",
			Namespace: cicdNS,
		},
		Spec: triggersv1.EventListenerSpec{
			ServiceAccountName: saName,
			Triggers:           triggers,
		},
	}
}

func createListenerObjectMeta(name, ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: ns,
	}
}
