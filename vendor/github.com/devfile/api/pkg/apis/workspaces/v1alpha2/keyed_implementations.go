package v1alpha2

import (
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
