package scm

import (
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// Repository interface exposes generic functions that will be
// implemented by repositories (Github,Gitlab,Bitbucket,etc)
type Repository interface {

	// Get Pull Request TriggerBinding name for this repository provider
	PRBindingName() string

	// Get Push TriggerBinding name for this repository provider
	PushBindingName() string

	// Create a TriggerBinding for PullRequest hooks.
	CreatePRBinding(namespace string) (triggersv1.TriggerBinding, string)

	// Create a TriggerBinding for Push Request hooks
	CreatePushBinding(namespace string) (triggersv1.TriggerBinding, string)

	// Create a CI eventlistener trigger
	CreateCITrigger(name, secretName, secretNs, template string, bindings []string) triggersv1.EventListenerTrigger

	// Create a CD eventlistener trigger
	CreateCDTrigger(name, secretName, secretNs, template string, bindings []string) triggersv1.EventListenerTrigger

	// Git Repository URL
	URL() string
}
