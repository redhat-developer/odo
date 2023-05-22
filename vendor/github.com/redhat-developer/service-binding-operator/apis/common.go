package apis

import (
	"encoding/json"
	"errors"
	"k8s.io/api/authentication/v1"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finalizerName          = "finalizer.servicebinding.openshift.io"
	requesterAnnotationKey = "servicebinding.io/requester"
)

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
	client.Object
	HasDeletionTimestamp() bool
	StatusConditions() []metav1.Condition
	GetSpec() interface{}
}

func CanUpdateBinding(obj Object, oldObj Object) error {
	if meta.IsStatusConditionTrue(obj.StatusConditions(), BindingReady) && !reflect.DeepEqual(obj.GetSpec(), oldObj.GetSpec()) {
		return errors.New("cannot update Service Binding if 'Ready' condition is True. If you want to rebind to another service/application, remove this binding and create a new one.")
	}

	return nil
}

func SetRequester(obj *unstructured.Unstructured, userInfo v1.UserInfo) {
	jsonContent, _ := json.Marshal(userInfo)
	anns := obj.GetAnnotations()
	if anns == nil {
		anns = make(map[string]string)
	}
	anns[requesterAnnotationKey] = string(jsonContent)
	obj.SetAnnotations(anns)
}

// Return username of requester who submitted the service binding
func Requester(objMeta metav1.ObjectMeta) *authv1.UserInfo {
	req, found := objMeta.Annotations[requesterAnnotationKey]
	if found {
		userInfo := &authv1.UserInfo{}
		err := json.Unmarshal([]byte(req), userInfo)
		if err != nil {
			return nil
		}
		return userInfo
	}
	return nil
}
