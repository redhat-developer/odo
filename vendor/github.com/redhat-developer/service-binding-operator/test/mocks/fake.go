package mocks

import (
	"testing"

	pgapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"

	ocav1 "github.com/openshift/api/apps/v1"
	knativev1 "knative.dev/serving/pkg/apis/serving/v1"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Fake defines all the elements to fake a kubernetes api client.
type Fake struct {
	t    *testing.T       // testing instance
	ns   string           // namespace
	S    *runtime.Scheme  // runtime client scheme
	objs []runtime.Object // all fake objects
}

// AddMockedServiceBindingRequest add mocked object from ServiceBindingRequestMock.
func (f *Fake) AddMockedServiceBindingRequest(
	name string,
	backingServiceNamespace *string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *v1alpha1.ServiceBindingRequest {
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	sbr := ServiceBindingRequestMock(f.ns, name, backingServiceNamespace, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedServiceBindingRequestWithUnannotated add mocked object from ServiceBindingRequestMock with DetectBindingResources.
func (f *Fake) AddMockedServiceBindingRequestWithUnannotated(
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *v1alpha1.ServiceBindingRequest {
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	sbr := ServiceBindingRequestMock(f.ns, name, nil, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredServiceBindingRequestWithoutApplication creates a mock ServiceBindingRequest object
func (f *Fake) AddMockedUnstructuredServiceBindingRequestWithoutApplication(
	name string,
	backingServiceResourceRef string,
) *unstructured.Unstructured {
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	var emptyGVR = schema.GroupVersionResource{}
	sbr, err := UnstructuredServiceBindingRequestMock(f.ns, name, backingServiceResourceRef, "", emptyGVR, nil)
	require.NoError(f.t, err)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredServiceBindingRequest creates a mock ServiceBindingRequest object
func (f *Fake) AddMockedUnstructuredServiceBindingRequest(
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *unstructured.Unstructured {
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	sbr, err := UnstructuredServiceBindingRequestMock(f.ns, name, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	require.NoError(f.t, err)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredCSV add mocked unstructured CSV.
func (f *Fake) AddMockedUnstructuredCSV(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionMock(f.ns, name)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedCSVList add mocked object from ClusterServiceVersionListMock.
func (f *Fake) AddMockedCSVList(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListMock(f.ns, name))
}

// AddMockedCSVWithVolumeMountList add mocked object from ClusterServiceVersionListVolumeMountMock.
func (f *Fake) AddMockedCSVWithVolumeMountList(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListVolumeMountMock(f.ns, name))
}

// AddMockedUnstructuredCSVWithVolumeMount same than AddMockedCSVWithVolumeMountList but using
// unstructured object.
func (f *Fake) AddMockedUnstructuredCSVWithVolumeMount(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionVolumeMountMock(f.ns, name)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedDatabaseCR add mocked object from DatabaseCRMock.
func (f *Fake) AddMockedDatabaseCR(ref string, namespace string) runtime.Object {
	require.NoError(f.t, pgapis.AddToScheme(f.S))
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	mock := DatabaseCRMock(namespace, ref)
	f.objs = append(f.objs, mock)
	return mock
}

func (f *Fake) AddMockedUnstructuredDatabaseCR(ref string) {
	require.NoError(f.t, pgapis.AddToScheme(f.S))
	d, err := UnstructuredDatabaseCRMock(f.ns, ref)
	require.NoError(f.t, err)
	f.objs = append(f.objs, d)
}

// AddMockedUnstructuredDeploymentConfig adds mocked object from UnstructuredDeploymentConfigMock.
func (f *Fake) AddMockedUnstructuredDeploymentConfig(name string, matchLabels map[string]string) {
	require.Nil(f.t, ocav1.AddToScheme(f.S))
	d, err := UnstructuredDeploymentConfigMock(f.ns, name, matchLabels)
	require.Nil(f.t, err)
	f.S.AddKnownTypes(ocav1.SchemeGroupVersion, &ocav1.DeploymentConfig{})
	f.objs = append(f.objs, d)
}

// AddMockedUnstructuredDeployment add mocked object from UnstructuredDeploymentMock.
func (f *Fake) AddMockedUnstructuredDeployment(name string, matchLabels map[string]string) *unstructured.Unstructured {
	require.NoError(f.t, appsv1.AddToScheme(f.S))
	d, err := UnstructuredDeploymentMock(f.ns, name, matchLabels)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.Deployment{})
	f.objs = append(f.objs, d)
	return d
}

// AddMockedUnstructuredKnativeService add mocked object from UnstructuredKnativeService.
func (f *Fake) AddMockedUnstructuredKnativeService(name string, matchLabels map[string]string) {
	require.NoError(f.t, knativev1.AddToScheme(f.S))
	d, err := UnstructuredKnativeServiceMock(f.ns, name, matchLabels)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(knativev1.SchemeGroupVersion, &knativev1.Service{})
	f.objs = append(f.objs, d)
}

func (f *Fake) AddMockedUnstructuredDatabaseCRD() *unstructured.Unstructured {
	require.NoError(f.t, apiextensionv1beta1.AddToScheme(f.S))
	c, err := UnstructuredDatabaseCRDMock(f.ns)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(apiextensionv1beta1.SchemeGroupVersion, &apiextensionv1beta1.CustomResourceDefinition{})
	f.objs = append(f.objs, c)
	return c
}

func (f *Fake) AddMockedUnstructuredPostgresDatabaseCR(ref string) *unstructured.Unstructured {
	d, err := UnstructuredPostgresDatabaseCRMock(f.ns, ref)
	require.NoError(f.t, err)
	f.objs = append(f.objs, d)
	return d
}

// AddMockedUnstructuredSecret add mocked object from SecretMock.
func (f *Fake) AddMockedUnstructuredSecret(name string) *unstructured.Unstructured {
	s, err := UnstructuredSecretMock(f.ns, name)
	require.NoError(f.t, err)
	f.objs = append(f.objs, s)
	return s
}

// AddNamespacedMockedSecret add mocked object from SecretMock in a namespace
// which isn't necessarily same as that of the ServiceBindingRequest namespace.
func (f *Fake) AddNamespacedMockedSecret(name string, namespace string) {
	f.objs = append(f.objs, SecretMock(namespace, name))
}

// AddMockedUnstructuredConfigMap add mocked object from ConfigMapMock.
func (f *Fake) AddMockedUnstructuredConfigMap(name string) {
	mock := ConfigMapMock(f.ns, name)
	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mock)
	require.NoError(f.t, err)
	f.objs = append(f.objs, &unstructured.Unstructured{Object: uObj})
}

func (f *Fake) AddMockResource(resource runtime.Object) {
	f.objs = append(f.objs, resource)
}

// FakeDynClient returns fake dynamic api client.
func (f *Fake) FakeDynClient() *fakedynamic.FakeDynamicClient {
	return fakedynamic.NewSimpleDynamicClient(f.S, f.objs...)
}

// NewFake instantiate Fake type.
func NewFake(t *testing.T, ns string) *Fake {
	return &Fake{t: t, ns: ns, S: scheme.Scheme}
}
