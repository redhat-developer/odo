package annotations

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestAttributeHandler exercises the AttributeHandler's ability to extract values according to the
// given annotation name and value.
func TestAttributeHandler(t *testing.T) {
	type args struct {
		obj      *unstructured.Unstructured
		key      string
		value    string
		expected map[string]interface{}
	}

	assertHandler := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			bindingInfo, err := NewBindingInfo(args.key, args.value)
			require.NoError(t, err)
			require.NotNil(t, bindingInfo)
			handler := NewAttributeHandler(bindingInfo, *args.obj)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expected, got.Data)
		}
	}

	// "scalar" tests whether a single deep scalar value can be extracted from the given object.
	t.Run("should extract a single value from .status.dbConnectionIP into .dbConnectionIP",
		assertHandler(args{
			expected: map[string]interface{}{
				"dbConnectionIP": "127.0.0.1",
			},
			key:   "servicebindingoperator.redhat.io/status.dbConnectionIP",
			value: "binding:env:attribute",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIP": "127.0.0.1",
					},
				},
			},
		}),
	)

	// "scalar#alias" tests whether a single deep scalar value can be extracted from the given object
	// returning a different name than the original given path.
	t.Run("should extract a single value from .status.dbConnectionIP into .alias",
		assertHandler(args{
			expected: map[string]interface{}{
				"alias": "127.0.0.1",
			},
			key:   "servicebindingoperator.redhat.io/alias-status.dbConnectionIP",
			value: "binding:env:attribute",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIP": "127.0.0.1",
					},
				},
			},
		}),
	)

	// tests whether a deep slice value can be extracted from the given object.
	t.Run("should extract a slice from .status.dbConnectionIPs into .dbConnectionIPs",
		assertHandler(args{
			expected: map[string]interface{}{
				"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
			},
			key:   "servicebindingoperator.redhat.io/status.dbConnectionIPs",
			value: "binding:env:attribute",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"dbConnectionIPs": []string{"127.0.0.1", "1.1.1.1"},
					},
				},
			},
		}),
	)

	// tests whether a deep map value can be extracted from the given object.
	t.Run("should extract a map from .status.connection into .connection", assertHandler(args{
		expected: map[string]interface{}{
			"connection": map[string]interface{}{
				"host": "127.0.0.1",
				"port": "1234",
			},
		},
		key:   "servicebindingoperator.redhat.io/status.connection",
		value: "binding:env:attribute",
		obj: &unstructured.Unstructured{
			Object: map[string]interface{}{
				"status": map[string]interface{}{
					"connection": map[string]interface{}{
						"host": "127.0.0.1",
						"port": "1234",
					},
				},
			},
		},
	}))

	// "map.key" tests whether a deep map key can be extracted from the given object.
	t.Run("should extract a single map key from .status.connection into .connection",
		assertHandler(args{
			expected: map[string]interface{}{
				"connection": map[string]interface{}{
					"host": "127.0.0.1",
				},
			},
			key:   "servicebindingoperator.redhat.io/status.connection.host",
			value: "binding:env:attribute",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"connection": map[string]interface{}{
							"host": "127.0.0.1",
							"port": "1234",
						},
					},
				},
			},
		}),
	)
}
