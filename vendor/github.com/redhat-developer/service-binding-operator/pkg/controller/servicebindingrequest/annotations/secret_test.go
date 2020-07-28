package annotations

import (
	"testing"

	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestSecretHandler(t *testing.T) {
	type args struct {
		name      string
		value     string
		service   map[string]interface{}
		resources []runtime.Object
		expected  map[string]interface{}
	}

	assertHandler := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			f := mocks.NewFake(t, "test")

			for _, r := range args.resources {
				f.AddMockResource(r)
			}

			restMapper := testutils.BuildTestRESTMapper()

			bindingInfo, err := NewBindingInfo(args.name, args.value)
			require.NoError(t, err)
			handler, err := NewSecretHandler(
				f.FakeDynClient(),
				bindingInfo,
				unstructured.Unstructured{Object: args.service},
				restMapper,
			)
			require.NoError(t, err)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expected, got.Data)
		}
	}

	t.Run("secret/scalar", assertHandler(args{
		name:  "servicebindingoperator.redhat.io/status.dbCredentials-password",
		value: "binding:env:object:secret",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},
				Data: map[string][]byte{
					"password": []byte("hunter2"),
				},
			},
		},
		expected: map[string]interface{}{
			"secret": map[string]interface{}{
				"password": "hunter2",
			},
		},
	}))

	t.Run("secret/map", assertHandler(args{
		name:  "servicebindingoperator.redhat.io/status.dbCredentials",
		value: "binding:env:object:secret",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},
				Data: map[string][]byte{
					"password": []byte("hunter2"),
					"username": []byte("AzureDiamond"),
				},
			},
		},
		expected: map[string]interface{}{
			"secret": map[string]interface{}{
				"status": map[string]interface{}{
					"dbCredentials": map[string]interface{}{
						"username": "AzureDiamond",
						"password": "hunter2",
					},
				},
			},
		},
	}))
}
