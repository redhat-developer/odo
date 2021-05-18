package overriding

import (
	"fmt"
	"reflect"

	dw "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	attributesPkg "github.com/devfile/api/v2/pkg/attributes"
	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/util/sets"
)

type checkFn func(elementType string, keysSets []sets.String) []error

// checkKeys provides a generic way to apply some validation on the content of each type of top-level list
// contained in the `toplevelListContainers` passed in argument.
//
// For each type of top-level list, the `keysSets` argument that will be passed to the `doCheck` function
// contains the the key sets that correspond to the `toplevelListContainers` passed to this method,
// in the same order.
func checkKeys(doCheck checkFn, toplevelListContainers ...dw.TopLevelListContainer) error {
	var errors *multierror.Error

	// intermediate storage for the conversion []map[string]KeyedList -> map[string][]sets.String
	listTypeToKeys := map[string][]sets.String{}

	// Flatten []map[string]KeyedList -> map[string][]KeyedList based on map keys and convert each KeyedList
	// into a sets.String
	for _, topLevelListContainer := range toplevelListContainers {
		topLevelList := topLevelListContainer.GetToplevelLists()
		for listType, listElem := range topLevelList {
			listTypeToKeys[listType] = append(listTypeToKeys[listType], sets.NewString(listElem.GetKeys()...))
		}

		value := reflect.ValueOf(topLevelListContainer)

		var variableValue reflect.Value
		var attributeValue reflect.Value

		// toplevelListContainers can contain either a pointer or a struct and needs to be safeguarded when using reflect
		if value.Kind() == reflect.Ptr {
			variableValue = value.Elem().FieldByName("Variables")
			attributeValue = value.Elem().FieldByName("Attributes")
		} else {
			variableValue = value.FieldByName("Variables")
			attributeValue = value.FieldByName("Attributes")
		}

		if variableValue.IsValid() && variableValue.Kind() == reflect.Map {
			mapIter := variableValue.MapRange()

			var variableKeys []string
			for mapIter.Next() {
				k := mapIter.Key()
				v := mapIter.Value()
				if k.Kind() != reflect.String || v.Kind() != reflect.String {
					return fmt.Errorf("unable to fetch top-level Variables, top-level Variables should be map of strings")
				}
				variableKeys = append(variableKeys, k.String())
			}
			listTypeToKeys["Variables"] = append(listTypeToKeys["Variables"], sets.NewString(variableKeys...))
		}

		if attributeValue.IsValid() && attributeValue.CanInterface() {
			attributes, ok := attributeValue.Interface().(attributesPkg.Attributes)
			if !ok {
				return fmt.Errorf("unable to fetch top-level Attributes from the devfile data")
			}
			var attributeKeys []string
			for k := range attributes {
				attributeKeys = append(attributeKeys, k)
			}
			listTypeToKeys["Attributes"] = append(listTypeToKeys["Attributes"], sets.NewString(attributeKeys...))
		}
	}

	for listType, keySets := range listTypeToKeys {
		errors = multierror.Append(errors, doCheck(listType, keySets)...)
	}
	return errors.ErrorOrNil()
}
