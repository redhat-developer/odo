package servicebindingrequest

import (
	"reflect"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	reconcilerNs   = "testing"
	reconcilerName = "binding-request"
)

var (
	secretsGVR           = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	deploymentsGVR       = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deploymentConfigsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deploymentconfigs"}
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

// reconcileRequest creates a reconcile.Request object using global variables.
func reconcileRequest() reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: reconcilerNs,
			Name:      reconcilerName,
		},
	}
}

// TestApplicationSelectorByName tests discovery of application by name
func TestApplicationSelectorByName(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef := "applicationRef"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, applicationResourceRef, deploymentsGVR, nil)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(applicationResourceRef, nil)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}

	t.Run("test-application-selector-by-name", func(t *testing.T) {

		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.Equal(t, BindingReady, sbrOutput.Status.Conditions[0].Type)
		require.Equal(t, corev1.ConditionTrue, sbrOutput.Status.Conditions[0].Status)
		require.Equal(t, 1, len(sbrOutput.Status.Applications))
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef,
			},
		}
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

// TestReconcilerReconcileUsingSecret test the reconciliation process using a secret, expected to be
// the regular approach.
func TestReconcilerReconcileUsingSecret(t *testing.T) {
	backingServiceResourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, reconcilerName, deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}

	t.Run("reconcile-using-secret", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}

		u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
		require.NoError(t, err)

		d := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
		require.NoError(t, err)

		containers := d.Spec.Template.Spec.Containers
		require.Equal(t, 1, len(containers))
		require.Equal(t, 1, len(containers[0].EnvFrom))
		require.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		require.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		namespacedName = types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.Equal(t, BindingReady, sbrOutput.Status.Conditions[0].Type)
		require.Equal(t, corev1.ConditionTrue, sbrOutput.Status.Conditions[0].Status)
		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)

		require.Equal(t, 1, len(sbrOutput.Status.Applications))
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: namespacedName.Name,
			},
		}
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

func TestReconcilerReconcileUsingVolumes(t *testing.T) {
	backingServiceResourceRef := "test-using-volumes"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, reconcilerName, deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSVWithVolumeMount("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}

	t.Run("reconcile-using-volume", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
		require.NoError(t, err)

		d := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
		require.NoError(t, err)

		containers := d.Spec.Template.Spec.Containers
		require.Equal(t, 1, len(containers[0].VolumeMounts))
		require.Equal(t, "/var/redhat", containers[0].VolumeMounts[0].MountPath)
		require.Equal(t, reconcilerName, containers[0].VolumeMounts[0].Name)

		volumes := d.Spec.Template.Spec.Volumes
		require.Equal(t, 1, len(volumes))
		require.Equal(t, reconcilerName, volumes[0].Name)
		require.Equal(t, reconcilerName, volumes[0].VolumeSource.Secret.SecretName)
	})
}

func TestReconcilerGenericBinding(t *testing.T) {
	backingServiceResourceRef := "backingService1"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}

	// Reconcile without deployment
	res, err := reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput.Status.Conditions,
			BindingReady,
			corev1.ConditionTrue,
		),
		"Ready condition should exist and true; existing conditions: %+v",
		sbrOutput.Status.Conditions,
	)
	require.Len(t, sbrOutput.Status.Applications, 0)

	// Reconcile with deployment
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	fakeDynClient = f.FakeDynClient()
	reconciler = &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}
	res, err = reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
	require.NoError(t, err)

	d := appsv1.Deployment{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
	require.NoError(t, err)

	sbrOutput2, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.Equal(t, BindingReady, sbrOutput2.Status.Conditions[0].Type)
	require.Equal(t, corev1.ConditionTrue, sbrOutput2.Status.Conditions[0].Status)
	require.Equal(t, reconcilerName, sbrOutput2.Status.Secret)
	require.Equal(t, 1, len(sbrOutput2.Status.Applications))

	u, err = fakeDynClient.Resource(secretsGVR).Get("db-credentials", metav1.GetOptions{})
	require.NoError(t, err)
	s := corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &s)
	require.NoError(t, err)

	// Update Credentials
	s.Data["password"] = []byte("abc123")
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&s)
	require.NoError(t, err)
	updated := unstructured.Unstructured{Object: obj}
	_, err = fakeDynClient.Resource(secretsGVR).Namespace(updated.GetNamespace()).Update(&updated, metav1.UpdateOptions{})
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	reconciler = &Reconciler{dynClient: fakeDynClient, RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}
	res, err = reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	sbrOutput3, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)
	require.Equal(t, BindingReady, sbrOutput3.Status.Conditions[0].Type)
	require.Equal(t, corev1.ConditionTrue, sbrOutput3.Status.Conditions[0].Status)
	require.Equal(t, reconcilerName, sbrOutput3.Status.Secret)
	require.Equal(t, s.Data["password"], []byte("abc123"))
	require.Equal(t, 1, len(sbrOutput3.Status.Applications))
}

//TestReconcilerReconcileWithConflictingAppSelc tests when sbr has conflicting ApplicationSel such as MatchLabels=App1 and ResourceRef=App2 it should prioritise the ResourceRef
func TestReconcilerReconcileWithConflictingAppSelc(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef1 := "applicationResourceRef1"
	matchLabels1 := map[string]string{
		"connects-to": "database",
		"environment": "testing",
	}
	applicationResourceRef2 := "applicationResourceRef2"

	f := mocks.NewFake(t, reconcilerNs)

	f.AddMockedUnstructuredDeployment(applicationResourceRef1, matchLabels1)
	f.AddMockedUnstructuredDeployment(applicationResourceRef2, nil)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, applicationResourceRef2, deploymentsGVR, matchLabels1)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()

	restMapper := testutils.BuildTestRESTMapper()

	reconciler := &Reconciler{
		dynClient:  fakeDynClient,
		scheme:     f.S,
		RestMapper: restMapper,
	}

	t.Run("test-reconciler-reconcile-with-conflicting-application-selector", func(t *testing.T) {

		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef2,
			},
		}

		require.Equal(t, BindingReady, sbrOutput.Status.Conditions[0].Type)
		require.Equal(t, corev1.ConditionTrue, sbrOutput.Status.Conditions[0].Status)
		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)
		require.Len(t, sbrOutput.Status.Applications, 1)
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

// TestEmptyApplicationSelector tests that Status is successfully updated when ApplicationSelector is missing
func TestEmptyApplicationSelector(t *testing.T) {
	backingServiceResourceRef := "backingService1"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequestWithoutApplication(reconcilerName, backingServiceResourceRef)
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)

	reconciler := &Reconciler{dynClient: f.FakeDynClient(), RestMapper: testutils.BuildTestRESTMapper(), scheme: f.S}

	res, err := reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.Equal(t, BindingReady, sbrOutput.Status.Conditions[0].Type)
	// Currently the Conditions[0].Status would be true as application's absence won't cause error
	// TODO New steps to conditions to be introduced - InjectionReady, CollectionReady
	require.Equal(t, corev1.ConditionTrue, sbrOutput.Status.Conditions[0].Status)
	require.Equal(t, 0, len(sbrOutput.Status.Applications))
}
