package servicebindingrequest

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// ServiceBindingRequestResource the name of ServiceBindingRequest resource.
	ServiceBindingRequestResource = "servicebindingrequests"
	// ServiceBindingRequestKind defines the name of the CRD kind.
	ServiceBindingRequestKind = "ServiceBindingRequest"
	// DeploymentConfigKind defines the name of DeploymentConfig kind.
	DeploymentConfigKind = "DeploymentConfig"
	// ClusterServiceVersionKind the name of ClusterServiceVersion kind.
	ClusterServiceVersionKind = "ClusterServiceVersion"
	// SecretResource defines the resource name for Secrets.
	SecretResource = "secrets"
	// SecretKind defines the name of Secret kind.
	SecretKind = "Secret"
)

// RequeueOnNotFound inspect error, if not-found then returns Requeue, otherwise expose the error.
func RequeueOnNotFound(err error, requeueAfter int64) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return Requeue(nil, requeueAfter)
	}
	return NoRequeue(err)
}

// RequeueOnConflict in case of conflict error, returning the error with requeue, otherwise Done.
func RequeueOnConflict(err error) (reconcile.Result, error) {
	if errors.IsConflict(err) {
		return RequeueError(err)
	}
	return Done()
}

// RequeueError simply requeue exposing the error.
func RequeueError(err error) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

// Requeue based on empty result and no error informed upstream, request will be requeued.
func Requeue(err error, requeueAfter int64) (reconcile.Result, error) {
	return reconcile.Result{
		RequeueAfter: time.Duration(requeueAfter) * time.Second,
		Requeue:      true,
	}, err
}

// Done when no error is informed and request is not set for requeue.
func Done() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// DoneOnNotFound will return done when error is not-found, otherwise it calls out NoRequeue.
func DoneOnNotFound(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return Done()
	}
	return NoRequeue(err)
}

// NoRequeue returns error without requeue flag.
func NoRequeue(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}

// containsStringSlice given a string slice and a string, returns boolean when is contained.
func containsStringSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// removeStringSlice given a string slice and a string, returns a new slice without given string.
func removeStringSlice(slice []string, str string) []string {
	var cleanSlice []string
	for _, s := range slice {
		if str != s {
			cleanSlice = append(cleanSlice, s)
		}
	}
	return cleanSlice
}
