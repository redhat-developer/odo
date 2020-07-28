package servicebindingrequest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	ns := "testing"
	backingServiceNs := "backing-servicec-ns"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNs)

	cr, err := mocks.UnstructuredDatabaseCRMock(backingServiceNs, crName)
	require.NoError(t, err)

	crInSameNamespace, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	serviceCtxs := ServiceContextList{
		{
			Service: cr,
		},
		{
			Service: crInSameNamespace,
		},
	}

	fakeDynClient := f.FakeDynClient()

	toTmpl := func(obj *unstructured.Unstructured) string {
		gvk := obj.GetObjectKind().GroupVersionKind()
		name := obj.GetName()
		return fmt.Sprintf(`{{ index . %q %q %q %q "metadata" "name" }}`, gvk.Version, gvk.Group, gvk.Kind, name)
	}

	actual, _, err := NewRetriever(fakeDynClient).ProcessServiceContexts(
		"SERVICE_BINDING",
		serviceCtxs,
		[]v1.EnvVar{
			{Name: "SAME_NAMESPACE", Value: toTmpl(crInSameNamespace)},
			{Name: "OTHER_NAMESPACE", Value: toTmpl(cr)},
			{Name: "DIRECT_ACCESS", Value: `{{ .v1alpha1.postgresql_baiju_dev.Database.db_testing.metadata.name }}`},
		},
	)
	require.NoError(t, err)
	require.Equal(t, map[string][]byte{
		"SERVICE_BINDING_SAME_NAMESPACE":  []byte(crInSameNamespace.GetName()),
		"SERVICE_BINDING_OTHER_NAMESPACE": []byte(cr.GetName()),
		"SERVICE_BINDING_DIRECT_ACCESS":   []byte(cr.GetName()),
	}, actual)
}

func TestBuildServiceEnvVars(t *testing.T) {

	type testCase struct {
		ctx                *ServiceContext
		globalEnvVarPrefix string
		expected           map[string]string
	}

	cr, err := mocks.UnstructuredDatabaseCRMock("namespace", "name")
	require.NoError(t, err)

	serviceEnvVarPrefix := "serviceprefix"
	emptyString := ""

	testCases := []testCase{
		{
			globalEnvVarPrefix: "",
			ctx: &ServiceContext{
				EnvVarPrefix: &emptyString,
				EnvVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "globalprefix",
			ctx: &ServiceContext{
				EnvVarPrefix: &emptyString,
				EnvVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"GLOBALPREFIX_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "globalprefix",
			ctx: &ServiceContext{
				EnvVarPrefix: &serviceEnvVarPrefix,
				EnvVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"GLOBALPREFIX_SERVICEPREFIX_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &ServiceContext{
				Service:      cr,
				EnvVarPrefix: nil,
				EnvVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"DATABASE_APIKEY": "my-secret-key",
			},
		},
		{
			globalEnvVarPrefix: "",
			ctx: &ServiceContext{
				EnvVarPrefix: &serviceEnvVarPrefix,
				EnvVars: map[string]interface{}{
					"apiKey": "my-secret-key",
				},
			},
			expected: map[string]string{
				"SERVICEPREFIX_APIKEY": "my-secret-key",
			},
		},
	}

	for _, tc := range testCases {
		actual, err := buildServiceEnvVars(tc.ctx, tc.globalEnvVarPrefix)
		require.NoError(t, err)
		require.Equal(t, tc.expected, actual)
	}
}
