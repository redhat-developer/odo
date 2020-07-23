package servicebindingrequest

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestAnnotationsExtractNamespacedName(t *testing.T) {
	require.Equal(t, types.NamespacedName{}, extractSBRNamespacedName(map[string]string{}))

	data := map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"}
	require.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, extractSBRNamespacedName(data))
}

func TestAnnotationsGetSBRNamespacedNameFromObject(t *testing.T) {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"})

	// not containing annotations, should return empty
	t.Run("returns-empty", func(t *testing.T) {
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// with annotations in place it should return the actual values
	t.Run("from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, namespacedName)
	})

	// incomplete annotations, only name present
	t.Run("only-name-from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNameAnnotation: "name"})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// incomplete annotations, namespace annotation is present but an empty string
	t.Run("namespace-empty-from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNameAnnotation: "name", sbrNamespaceAnnotation: ""})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// incomplete annotations, only namespace present
	t.Run("only-namespace-from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "namespace"})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// incomplete annotations, name annotation is present but an empty string
	t.Run("name-empty-from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "namespace", sbrNameAnnotation: ""})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// it should also understand a actual SBR as well, so return not empty
	t.Run("actual-sbr-object", func(t *testing.T) {
		sbr := &unstructured.Unstructured{}
		sbr.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind))
		sbr.SetNamespace("ns")
		sbr.SetName("name")
		namespacedName, err := GetSBRNamespacedNameFromObject(sbr.DeepCopyObject())
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, namespacedName)
	})
}

func TestAnnotationsSetAndRemoveSBRAnnotations(t *testing.T) {
	ns := "annotations"
	f := mocks.NewFake(t, ns)

	matchLabels := map[string]string{}
	f.AddMockedUnstructuredDeployment(ns, matchLabels)

	client := f.FakeDynClient()
	namespacedName := types.NamespacedName{Namespace: ns, Name: ns}

	deploymentGVR := appsv1.SchemeGroupVersion.WithResource("deployments")
	deploymentResource := client.Resource(deploymentGVR).Namespace(ns)

	u, err := deploymentResource.Get(ns, metav1.GetOptions{})
	require.NoError(t, err)

	objs := []*unstructured.Unstructured{}
	objs = append(objs, u)

	t.Run("SetSBRAnnotations", func(t *testing.T) {
		originCopy := u.DeepCopy()
		newObj := SetSBRAnnotations(namespacedName, u)

		// we are not modifying the origin object
		equal, err := nestedMapComparison(u, originCopy)
		require.NoError(t, err)
		require.True(t, equal)

		equal, err = nestedMapComparison(u, newObj, []string{"metadata", "annotations"}...)
		require.NoError(t, err)
		require.False(t, equal)

		objNamespacedName, err := GetSBRNamespacedNameFromObject(newObj)
		require.NoError(t, err)
		require.Equal(t, namespacedName, objNamespacedName)

		// assert nothing else is changed
		newObj.SetAnnotations(nil)
		equal, err = nestedMapComparison(u, newObj)
		require.NoError(t, err)
		require.True(t, equal)
	})

	t.Run("RemoveSBRAnnotations", func(t *testing.T) {
		originCopy := u.DeepCopy()
		newObj := RemoveSBRAnnotations(u)

		// we are not modifying the origin object
		equal, err := nestedMapComparison(u, originCopy)
		require.NoError(t, err)
		require.True(t, equal)

		equal, err = nestedMapComparison(u, newObj, []string{"metadata", "annotations"}...)
		require.NoError(t, err)
		require.False(t, equal)

		objNamespacedName, err := GetSBRNamespacedNameFromObject(newObj)
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, objNamespacedName)

		// assert nothing else is changed
		newObj.SetAnnotations(u.GetAnnotations())
		equal, err = nestedMapComparison(u, newObj)
		require.NoError(t, err)
		require.True(t, equal)
	})

	t.Run("SetAndUpdateSBRAnnotations", func(t *testing.T) {
		err := SetAndUpdateSBRAnnotations(client, namespacedName, objs)
		require.NoError(t, err)

		u, err := deploymentResource.Get(ns, metav1.GetOptions{})
		require.NoError(t, err)

		objNamespacedName, err := GetSBRNamespacedNameFromObject(u)
		require.NoError(t, err)
		require.Equal(t, namespacedName, objNamespacedName)
	})

	t.Run("RemoveAndUpdateSBRAnnotations", func(t *testing.T) {
		err := RemoveAndUpdateSBRAnnotations(client, objs)
		require.NoError(t, err)

		u, err := deploymentResource.Get(ns, metav1.GetOptions{})
		require.NoError(t, err)

		objNamespacedName, err := GetSBRNamespacedNameFromObject(u)
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, objNamespacedName)
	})
}

// Test_extractSBRNamespacedName verifies whether extractSBRNamespacedName returns an empty
// NamespacedName where appropriate.
func Test_extractSBRNamespacedName(t *testing.T) {
	type args struct {
		data map[string]string
	}

	tests := []struct {
		name string
		args args
		want types.NamespacedName
	}{
		{
			name: "empty map",
			args: args{
				data: map[string]string{},
			},
			want: types.NamespacedName{},
		},
		{
			name: "all annotations present",
			args: args{
				data: map[string]string{
					sbrNameAnnotation:      "name",
					sbrNamespaceAnnotation: "namespace",
				},
			},
			want: types.NamespacedName{Name: "name", Namespace: "namespace"},
		},
		{
			name: "only name present",
			args: args{
				data: map[string]string{
					sbrNameAnnotation: "name",
				},
			},
			want: types.NamespacedName{},
		},
		{
			name: "namespace present but empty",
			args: args{
				data: map[string]string{
					sbrNameAnnotation:      "name",
					sbrNamespaceAnnotation: "",
				},
			},
			want: types.NamespacedName{},
		},
		{
			name: "only namespace present",
			args: args{
				data: map[string]string{
					sbrNamespaceAnnotation: "namespace",
				},
			},
			want: types.NamespacedName{},
		},
		{
			name: "name present but empty",
			args: args{
				data: map[string]string{
					sbrNameAnnotation:      "",
					sbrNamespaceAnnotation: "namespace",
				},
			},
			want: types.NamespacedName{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractSBRNamespacedName(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractSBRNamespacedName() = %v, want %v", got, tt.want)
			}
		})
	}
}
