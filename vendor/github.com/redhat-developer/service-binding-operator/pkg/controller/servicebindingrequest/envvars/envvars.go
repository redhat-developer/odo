package envvars

import (
	"errors"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
)

// ErrUnsupportedType is returned when an unsupported type is encountered.
var ErrUnsupportedType = errors.New("unsupported type")

// Build returns an environment variable dictionary with an entry for each
// leaf containing a scalar value.
//
// Example:
//
// 	src := map[string]interface{}{
// 		"status": map[string]interface{}{
//			"listeners": []map[string]interface{}{
//				{
//					"type": "secure",
//					"addresses": []map[string]interface{}{
//						{
//							"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
//							"port": "9093",
//						},
//					},
//				},
//			},
//		},
//	}
//	actual, _ := Build(src)
//
// actual should contain the following values:
//
// 	"STATUS_LISTENERS_0_TYPE":             "secure",
// 	"STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
// 	"STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
//
// Now, consider the following example:
//
//	actual, _ = Build(src, "kafka")
//
// actual should contain the following values instead:
//
// 	"KAFKA_STATUS_LISTENERS_0_TYPE":             "secure",
// 	"KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_HOST": "my-cluster-kafka-bootstrap.coffeeshop.svc",
// 	"KAFKA_STATUS_LISTENERS_0_ADDRESSES_0_PORT": "9093",
//
func Build(obj interface{}, path ...string) (map[string]string, error) {
	// perform the appropriate action depending on its type; maybe at some point
	// reflection might be required.
	switch val := obj.(type) {
	case map[string]interface{}:
		return buildMap(val, path)
	case []map[string]interface{}:
		return buildSliceOfMap(val, path)
	case string:
		return buildString(val, path), nil
	case int:
		return buildString(strconv.Itoa(val), path), nil
	case int64:
		return buildString(strconv.FormatInt(val, 10), path), nil
	case float64:
		return buildString(strconv.FormatFloat(val, 'f', -1, 64), path), nil
	default:
		return nil, ErrUnsupportedType
	}
}

// buildEnvVarName returns the environment variable name for the given `path`.
func buildEnvVarName(path []string) string {
	// remove empty values from path
	newPath := []string{}
	for _, p := range path {
		if len(p) > 0 {
			newPath = append(newPath, p)
		}
	}
	envVar := strings.Join(newPath, "_")
	envVar = strings.ToUpper(envVar)
	return envVar
}

// buildString returns a map containing the environment variable, named using
// the given `path` and the given `s` value.
func buildString(val string, path []string) map[string]string {
	return map[string]string{
		buildEnvVarName(path): val,
	}
}

// buildMap returns a map containing environment variables for all the leaves
// present in the given `obj` map.
func buildMap(obj map[string]interface{}, path []string) (map[string]string, error) {
	envVars := make(map[string]string)
	for k, v := range obj {
		if err := buildInner(path, k, v, envVars); err != nil {
			return nil, err
		}
	}
	return envVars, nil
}

// buildSliceOfMap returns a map containing environment variables for all the
// leaves present in the given `obj` slice.
func buildSliceOfMap(obj []map[string]interface{}, acc []string) (map[string]string, error) {
	envVars := make(map[string]string)
	for i, v := range obj {
		k := strconv.Itoa(i)
		if err := buildInner(acc, k, v, envVars); err != nil {
			return nil, err
		}
	}
	return envVars, nil
}

// buildInner builds recursively an environment variable map for the given value
// and merges it with the given `envVars` map.
func buildInner(
	path []string,
	key string,
	value interface{},
	envVars map[string]string,
) error {
	if envVar, err := Build(value, append(path, key)...); err != nil {
		return err
	} else {
		return mergo.Merge(&envVars, envVar)
	}
}
