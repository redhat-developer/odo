package binding

import (
	"k8s.io/client-go/util/jsonpath"
	"reflect"
	"strings"
)

// getValuesByJSONPath returns values from the given map matching the provided JSONPath
// 'path' argument takes JSONPath expressions enclosed by curly braces {}
// see https://kubernetes.io/docs/reference/kubectl/jsonpath/ for more details
// It returns zero or more filtered values back,
// or error if the jsonpath is invalid or it cannot be applied on the given map
func getValuesByJSONPath(obj map[string]interface{}, path string) ([]reflect.Value, error) {
	j := jsonpath.New("")
	err := j.Parse(path)
	if err != nil {
		return nil, err
	}
	result, err := j.FindResults(obj)
	if err != nil {
		return nil, err
	}
	if len(result) > 1 {
		w := strings.Builder{}
		for i := range result {
			if err := j.PrintResults(&w, result[i]); err != nil {
				return nil, err
			}
		}
		return []reflect.Value{reflect.ValueOf(w.String())}, nil
	}
	return result[0], nil
}
