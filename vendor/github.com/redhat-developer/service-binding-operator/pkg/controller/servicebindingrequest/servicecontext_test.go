package servicebindingrequest

import (
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestBuildServiceContexts(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		*sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredSecret("db-credentials")

	restMapper := testutils.BuildTestRESTMapper()

	t.Run("existing selectors", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(
			f.FakeDynClient(), ns, extractServiceSelectors(sbr), false, restMapper)
		require.NoError(t, err)
		require.NotEmpty(t, serviceCtxs)
	})

	t.Run("empty selectors", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(f.FakeDynClient(), ns, nil, false, restMapper)
		require.NoError(t, err)
		require.Empty(t, serviceCtxs)
	})

	t.Run("services in different namespace", func(t *testing.T) {
		serviceCtxs, err := buildServiceContexts(
			f.FakeDynClient(), ns, extractServiceSelectors(sbr), false, restMapper)
		require.NoError(t, err)
		require.NotEmpty(t, serviceCtxs)
	})
}

var trueBool = true

func TestFindOwnedResourcesCtxs_ConfigMap(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		*sbr.Spec.BackingServiceSelector,
	}
	sbr.Spec.DetectBindingResources = trueBool

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredSecret("db-credentials")

	cr := mocks.DatabaseCRMock("test", "test")
	reference := metav1.OwnerReference{
		APIVersion:         cr.APIVersion,
		Kind:               cr.Kind,
		Name:               cr.Name,
		UID:                cr.UID,
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	configMap := mocks.ConfigMapMock("test", "test_database")
	us := &unstructured.Unstructured{}
	uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	require.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]metav1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	require.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]metav1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	restMapper := testutils.BuildTestRESTMapper()

	t.Run("existing selectors", func(t *testing.T) {
		got, err := findOwnedResourcesCtxs(
			f.FakeDynClient(),
			cr.GetNamespace(),
			cr.GetName(),
			cr.GetUID(),
			cr.GroupVersionKind(),
			nil,
			restMapper,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)

		expected := map[string]interface{}{
			"": map[string]interface{}{
				"password": "password",
				"user":     "user",
			},
		}
		require.Equal(t, expected, got[0].EnvVars)

	})
}

func TestFindOwnedResourcesCtxs_Secret(t *testing.T) {
	f := mocks.NewFake(t, "test")
	cr := mocks.DatabaseCRMock("test", "test")
	reference := metav1.OwnerReference{
		APIVersion:         cr.APIVersion,
		Kind:               cr.Kind,
		Name:               cr.Name,
		UID:                cr.UID,
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	secret := mocks.SecretMock("test", "test_database")
	us := &unstructured.Unstructured{}
	uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
	require.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]metav1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	require.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]metav1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	restMapper := testutils.BuildTestRESTMapper()

	t.Run("existing selectors", func(t *testing.T) {
		ownedResourcesCtxs, err := findOwnedResourcesCtxs(
			f.FakeDynClient(),
			cr.GetNamespace(),
			cr.GetName(),
			cr.GetUID(),
			cr.GroupVersionKind(),
			nil,
			restMapper,
		)
		require.NoError(t, err)
		require.NotEmpty(t, ownedResourcesCtxs)
	})
}
