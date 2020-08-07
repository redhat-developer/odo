package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestFindService(t *testing.T) {
	ns := "find-cr-tests"
	resourceRef := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	t.Run("missing service namespace", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(), "", db.GetObjectKind().GroupVersionKind(), resourceRef)
		require.Error(t, err)
		require.Equal(t, err, ErrUnspecifiedBackingServiceNamespace)
		require.Nil(t, cr)
	})

	t.Run("golden path", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(), ns, db.GetObjectKind().GroupVersionKind(), resourceRef)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestPlannerWithExplicitBackingServiceNamespace(t *testing.T) {
	ns := "planner"
	backingServiceNamespace := "backing-service-namespace"
	resourceRef := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(resourceRef, backingServiceNamespace)
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNamespace)

	t.Run("findService", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(),
			backingServiceNamespace,
			db.GetObjectKind().GroupVersionKind(),
			resourceRef,
		)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestFindServiceCRD(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)
	expected := f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	t.Run("golden path", func(t *testing.T) {
		crd, err := findServiceCRD(f.FakeDynClient(), cr.GetObjectKind().GroupVersionKind())
		require.NoError(t, err)
		require.NotNil(t, crd)
		require.Equal(t, expected, crd)
	})
}

func TestLoadDescriptor(t *testing.T) {
	type testCase struct {
		name       string
		path       string
		descriptor string
		root       string
		expected   map[string]string
	}

	testCases := []testCase{
		{
			name:       "should build proper annotation",
			descriptor: "binding:volumemount:secret:user",
			root:       "status",
			path:       "user",
			expected: map[string]string{
				"servicebindingoperator.redhat.io/status.user": "binding:volumemount:secret",
			},
		},
	}

	for _, args := range testCases {
		t.Run(args.name, func(t *testing.T) {
			anns := map[string]string{}
			loadDescriptor(anns, args.path, args.descriptor, args.root)
			require.Equal(t, args.expected, anns)
		})
	}
}

func TestBuildOwnerResourceContext(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)

	obj := f.AddMockedUnstructuredSecret("secret")

	type testCase struct {
		inputPath         string
		outputPath        string
		ownerEnvVarPrefix *string
	}

	testCases := []testCase{
		{
			inputPath:         "data.user",
			outputPath:        "user",
			ownerEnvVarPrefix: nil,
		},
	}

	for _, tc := range testCases {
		got, err := buildOwnedResourceContext(
			f.FakeDynClient(),
			obj,
			tc.ownerEnvVarPrefix,
			testutils.BuildTestRESTMapper(),
			tc.inputPath,
			tc.outputPath,
		)
		require.NoError(t, err)
		require.NotNil(t, got)
	}

}
