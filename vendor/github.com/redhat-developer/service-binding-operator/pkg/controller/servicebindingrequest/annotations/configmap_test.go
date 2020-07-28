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

func TestConfigMapHandler(t *testing.T) {
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
			handler, err := NewConfigMapHandler(
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

	t.Run("configmap/scalar", assertHandler(args{
		name:  "servicebindingoperator.redhat.io/status.dbCredentials-password",
		value: "binding:env:object:configmap",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string]string{
					"password": "hunter2",
				},
			},
		},
		expected: map[string]interface{}{
			"configmap": map[string]interface{}{
				"password": "hunter2",
			},
		},
	}))

	t.Run("configmap/map", assertHandler(args{
		name:  "servicebindingoperator.redhat.io/status.dbCredentials",
		value: "binding:env:object:configmap",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},
				Data: map[string]string{
					"password": "hunter2",
					"username": "AzureDiamond",
				},
			},
		},
		expected: map[string]interface{}{
			"configmap": map[string]interface{}{
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
