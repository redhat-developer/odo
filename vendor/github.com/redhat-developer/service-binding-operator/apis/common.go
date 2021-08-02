package apis

import (
	"errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const finalizerName = "finalizer.servicebinding.openshift.io"

func MaybeAddFinalizer(obj Object) bool {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizerName {
			return false
		}
	}
	obj.SetFinalizers(append(finalizers, finalizerName))
	return true
}

func MaybeRemoveFinalizer(obj Object) bool {
	finalizers := obj.GetFinalizers()
	for i, f := range finalizers {
		if f == finalizerName {
			obj.SetFinalizers(append(finalizers[:i], finalizers[i+1:]...))
			return true
		}
	}
	return false
}

type Object interface {
	runtime.Object
	GetFinalizers() []string
	SetFinalizers([]string)
	HasDeletionTimestamp() bool
	StatusConditions() []metav1.Condition
}

func CanUpdateBinding(obj Object) error {
	if obj.HasDeletionTimestamp() {
		return nil
	}
	if meta.IsStatusConditionTrue(obj.StatusConditions(), BindingReady) {
		return errors.New("cannot update Service Binding if 'Ready' condition is True. If you want to rebind to another service/application, remove this binding and create a new one.")
	}

	return nil
}
