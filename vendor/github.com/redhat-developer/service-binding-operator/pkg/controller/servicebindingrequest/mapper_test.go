package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func TestSBRRequestMapperMap(t *testing.T) {
	mapper := &SBRRequestMapper{}

	u := &unstructured.Unstructured{}
	u.SetNamespace("mapper-unit")
	u.SetName("mapper-unit")
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"})

	// not containing annotations, should return empty
	mapObj := handler.MapObject{Meta: u, Object: u.DeepCopyObject()}
	mappedRequests := mapper.Map(mapObj)
	require.Equal(t, 0, len(mappedRequests))

	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns", Name: "name"}}

	// with annotations in place it should return the actual values
	u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"})
	mapObj = handler.MapObject{Meta: u, Object: u.DeepCopyObject()}
	mappedRequests = mapper.Map(mapObj)
	require.Equal(t, 1, len(mappedRequests))
	require.Equal(t, request, mappedRequests[0])

	// it should also understand a actual SBR as well, so return not empty
	sbr := &unstructured.Unstructured{}
	sbr.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind))
	sbr.SetNamespace("ns")
	sbr.SetName("name")
	mapObj = handler.MapObject{Meta: u, Object: sbr.DeepCopyObject()}
	mappedRequests = mapper.Map(mapObj)
	require.Equal(t, 1, len(mappedRequests))
	require.Equal(t, request, mappedRequests[0])
}
