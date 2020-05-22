package scm

import (
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// Repository interface exposes generic functions that will be
// implemented by repositories (Github,Gitlab,Bitbucket,etc)
type Repository interface {
	CreatePRBinding(namespace string) (triggersv1.TriggerBinding, string)
	CreatePushBinding(namespace string) (triggersv1.TriggerBinding, string)
	CreateInterceptor(secretName, secretNs string) *triggersv1.EventInterceptor
	CreateCITrigger(name, secretName, secretNs, template string, bindings []string) v1alpha1.EventListenerTrigger
	CreateCDTrigger(name, secretName, secretNs, template string, bindings []string) v1alpha1.EventListenerTrigger
	URL() string
}
