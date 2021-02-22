package v1alpha2

import (
	"fmt"
	"reflect"
)

func extractKeys(keyedList interface{}) []Keyed {
	value := reflect.ValueOf(keyedList)
	keys := make([]Keyed, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.CanInterface() {
			i := elem.Interface()
			if keyed, ok := i.(Keyed); ok {
				keys = append(keys, keyed)
			}
		}
	}
	return keys
}

// CheckDuplicateKeys checks if duplicate keys are present in the devfile objects
func CheckDuplicateKeys(keyedList interface{}) error {
	seen := map[string]bool{}
	value := reflect.ValueOf(keyedList)
	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		if elem.CanInterface() {
			i := elem.Interface()
			if keyed, ok := i.(Keyed); ok {
				key := keyed.Key()
				if seen[key] {
					return fmt.Errorf("duplicate key: %s", key)
				}
				seen[key] = true
			}
		}
	}
	return nil
}
