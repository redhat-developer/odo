package envvars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {

	type testCase struct {
		name     string
		expected map[string]string
		src      interface{}
		path     []string
	}

	testCases := []testCase{
		{
			name: "should create envvars without prefix",
			expected: map[string]string{
				"STATUS_LISTENERS_0_TYPE":             "secure",
				"STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
			},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars with service prefix",
			expected: map[string]string{
				"KAFKA_STATUS_LISTENERS_0_TYPE":             "secure",
				"KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
			},
			path: []string{"kafka"},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars with binding and service prefixes",
			expected: map[string]string{
				"BINDING_KAFKA_STATUS_LISTENERS_0_TYPE":             "secure",
				"BINDING_KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"BINDING_KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
			},
			path: []string{"binding", "kafka"},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars without prefix",
			expected: map[string]string{
				"STATUS_LISTENERS_0_TYPE":             "secure",
				"STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
			},
			path: []string{""},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvar for int64 type",
			expected: map[string]string{
				"STATUS_VALUE": "-9223372036",
			},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"value": int64(-9223372036),
				},
			},
		},
		{
			name: "should create envvar for float64 type",
			expected: map[string]string{
				"": "100.72",
			},
			src: float64(100.72),
		},
		{
			name: "should create envvar for empty string type",
			expected: map[string]string{
				"": "",
			},
			src: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Build(tc.src, tc.path...)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
