package common

import (
	"reflect"

	apiAttributes "github.com/devfile/api/v2/pkg/attributes"
)

// DevfileOptions provides options for Devfile operations
type DevfileOptions struct {
	// Filter is a map that lets you filter devfile object against their attributes. Interface can be string, float, boolean or a map
	Filter map[string]interface{}
}

// FilterDevfileObject filters devfile attributes with the given options
func FilterDevfileObject(attributes apiAttributes.Attributes, options DevfileOptions) (bool, error) {
	filterIn := true
	for key, value := range options.Filter {
		var err error
		currentFilterIn := false
		attrValue := attributes.Get(key, &err)
		var keyNotFoundErr = &apiAttributes.KeyNotFoundError{Key: key}
		if err != nil && err.Error() != keyNotFoundErr.Error() {
			return false, err
		} else if reflect.DeepEqual(attrValue, value) {
			currentFilterIn = true
		}

		filterIn = filterIn && currentFilterIn
	}

	return filterIn, nil
}
