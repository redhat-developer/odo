package nested

import (
	"errors"
	"fmt"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested/accumulator"
)

// getValueFromMap attempts to retrieve from `obj` a value from the given path
// `path`.
func getValueFromMap(obj map[string]interface{}, path Path) (interface{}, bool, error) {
	head, exists := path.Head()
	if !exists {
		return obj, true, nil
	}
	val, exists := obj[head.Name]
	if !exists {
		return nil, false, nil
	}

	return getValue(val, path.Tail())
}

// collectValues accumulates the values found in all the given `path` in all
// elements present in `obj`.
func collectValues(obj []interface{}, path Path) (interface{}, error) {
	r := accumulator.NewAccumulator()
	for _, e := range obj {
		val, found, err := getValue(e, path.AdjustedPath())
		if err != nil || !found {
			return nil, err
		}
		err = r.Accumulate(val)
		if err != nil {
			return nil, err
		}
	}
	return r.Value(), nil
}

// UnsupportedTypeErr is returned when an unsupported type is encountered.
var UnsupportedTypeErr = errors.New("unsupported type")

// convertToSlice attempts to convert the given `src` into a `[]interface{}`. An
// error is returned when `src` is not one of the following:
//
// - []map[string]interface{}
// - []string
// - []int
//
func convertToSlice(src interface{}) ([]interface{}, error) {
	var obj []interface{}
	switch t := src.(type) {
	case []map[string]interface{}:
		obj = make([]interface{}, len(t))
		for i, e := range t {
			obj[i] = e
		}
	case []string:
		obj = make([]interface{}, len(t))
		for i, e := range t {
			obj[i] = e
		}
	case []int:
		obj = make([]interface{}, len(t))
		for i, e := range t {
			obj[i] = e
		}
	default:
		return nil, UnsupportedTypeErr
	}
	return obj, nil
}

// InvalidIndexErr is returned when a given index is out of bounds.
var InvalidIndexErr = errors.New("invalid index")

// getValueFromSlice attempts to return the value present at the given path.
func getValueFromSlice(s interface{}, path Path) (interface{}, bool, error) {
	// assert and convert s to []interface{}
	obj, err := convertToSlice(s)
	if err != nil {
		return nil, false, err
	}

	// it is required for path to have a head in this case, since it is expected to contain either an
	// Index or a sub-key in order to extract or aggregate the underlying value.
	head, ok := path.Head()
	if !ok {
		return nil, false, nil
	}

	if head.Index != nil {
		if *head.Index > len(obj) {
			return nil, false, InvalidIndexErr
		}
		m := obj[*head.Index]
		return getValue(m, path.Tail())
	}

	r, err := collectValues(obj, path)
	if err != nil {
		return nil, false, err
	}
	return r, true, nil
}

// getValue attempts to return the value present at the given path.
func getValue(obj interface{}, path Path) (interface{}, bool, error) {
	// return obj if path is empty.
	if _, ok := path.Head(); !ok {
		return obj, true, nil
	}

	switch val := obj.(type) {
	case string, int: // scalar
		if path.HasTail() {
			return nil, false, fmt.Errorf("type doesn't accept an index or key")
		}
		return val, true, nil
	case map[string]interface{}: // map
		return getValueFromMap(val, path)
	case []map[string]interface{}, []int, []string: // slice
		return getValueFromSlice(val, path)
	default:
		panic(fmt.Sprintf("missing type for %+v", val))
	}
}

// ComposeValue returns a map containing the structure of `path`, with `val` as value.
//
// The value is always transformed into a slice unless it is already one. For example, the call
//
//     ComposeValue(42, "foo.bar")
//
// yields the following result:
//
//     map[string]interface{}{
//         "foo": map[string]interface{}{
//             "bar": []int{42},
//         },
//     }
//
func ComposeValue(val interface{}, path Path) map[string]interface{} {
	// root is the resulting data-structure to be returned to caller.
	root := make(map[string]interface{})

	// n is a pointer to the current result node being processed.
	n := root

	// clean and split the path in `base` and `field`; for example, the path `a.b.*.c` is transformed
	// into `a.b.c`, resulting in `a.b` as base and `c` as field.
	base, field := path.Clean().Decompose()

	// populate the root structure with the wanted hierarchy; being each node a
	// map[string]interface{}.
	for _, f := range base {
		newVal := make(map[string]interface{})
		n[f.Name] = newVal
		// move the pointer to the last created value.
		n = newVal
	}

	n[field.Name] = val

	return root
}

// GetValue attempts to retrieve the value in the given string encoded path.
func GetValue(obj interface{}, p string, o string) (map[string]interface{}, bool, error) {
	path := NewPath(p)
	outputPath := NewPath(o)

	val, found, err := getValue(obj, path)
	if err != nil || !found {
		return nil, found, err
	}

	return ComposeValue(val, outputPath), found, nil
}
