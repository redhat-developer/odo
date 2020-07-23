package nested

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetValue(t *testing.T) {
	type args struct {
		src      map[string]interface{}
		path     string
		expected interface{}
	}

	assertGetValue := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			actual, found, err := GetValue(args.src, args.path, args.path)
			require.NoError(t, err)
			require.True(t, found)
			require.Equal(t, args.expected, actual)
		}
	}

	t.Run("key", assertGetValue(args{
		src: map[string]interface{}{
			"key": "value",
		},
		path: "key",
		expected: map[string]interface{}{
			"key": "value",
		},
	}))

	t.Run("key.subKey0", assertGetValue(args{
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"subKey0": "value0",
				"subKey1": "value1",
			},
		},
		path: "key.subKey0",
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"subKey0": "value0",
			},
		},
	}))

	t.Run("key.slice", assertGetValue(args{
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"subKey0": "value0-0",
				"slice": []map[string]interface{}{
					{"subKey0": "value0-1"},
				},
			},
		},
		path: "key.slice",
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey0": "value0-1"},
				},
			},
		},
	}))

	t.Run("key.slice.0", assertGetValue(args{
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey0": "value0"},
					{"subKey1": "value1"},
				},
			},
		},
		path: "key.slice.0",
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subKey0": "value0",
				},
			},
		},
	}))

	t.Run("key.slice.1", assertGetValue(args{
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey0": "value0"},
					{"subKey1": "value1"},
				},
			},
		},
		path: "key.slice.1",
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subKey1": "value1",
				},
			},
		},
	}))

	t.Run("key.slice.*", assertGetValue(args{
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey0": "value0"},
					{"subKey1": "value1"},
				},
			},
		},
		path: "key.slice.*",
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey0": "value0"},
					{"subKey1": "value1"},
				},
			},
		},
	}))

	t.Run("key.slice.*.subKey", assertGetValue(args{
		path: "key.slice.*.subKey",
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey": "value0"},
					{"subKey": "value1"},
				},
			},
		},
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subKey": []string{"value0", "value1"},
				},
			},
		},
	}))

	t.Run("key.slice.subKey", assertGetValue(args{
		path: "key.slice.subKey",
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{"subKey": "value0"},
					{"subKey": "value1"},
				},
			},
		},
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subKey": []string{"value0", "value1"},
				},
			},
		},
	}))

	t.Run("key.slice.subSlice", assertGetValue(args{
		path: "key.slice.subSlice",
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "value0",
								"intSubKey":    8080,
							},
						},
					},
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "host2",
								"intSubKey":    8081,
							},
						},
					},
				},
			},
		},
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subSlice": []map[string]interface{}{
						{
							"stringSubKey": "value0",
							"intSubKey":    8080,
						},
						{
							"stringSubKey": "host2",
							"intSubKey":    8081,
						},
					},
				},
			},
		},
	}))

	t.Run("key.slice.subSlice.stringSubKey", assertGetValue(args{
		path: "key.slice.subSlice.stringSubKey",
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "value0",
								"intSubKey":    8080,
							},
						},
					},
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "value1",
								"intSubKey":    8081,
							},
						},
					},
				},
			},
		},
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subSlice": map[string]interface{}{
						"stringSubKey": []string{"value0", "value1"},
					},
				},
			},
		},
	}))

	t.Run("key.slice.subSlice.intSubKey", assertGetValue(args{
		path: "key.slice.subSlice.intSubKey",
		src: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": []map[string]interface{}{
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "value0",
								"intSubKey":    8080,
							},
						},
					},
					{
						"subSlice": []map[string]interface{}{
							{
								"stringSubKey": "value1",
								"intSubKey":    8081,
							},
						},
					},
				},
			},
		},
		expected: map[string]interface{}{
			"key": map[string]interface{}{
				"slice": map[string]interface{}{
					"subSlice": map[string]interface{}{
						"intSubKey": []int{8080, 8081},
					},
				},
			},
		},
	}))
}
