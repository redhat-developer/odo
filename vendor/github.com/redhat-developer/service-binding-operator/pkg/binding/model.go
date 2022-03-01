package binding

import (
	"errors"
	"fmt"
	"strings"
)

type model struct {
	path        string
	elementType elementType
	objectType  objectType
	sourceKey   string
	sourceValue string
	value       string
}

func (m *model) isStringElementType() bool {
	return m.elementType == stringElementType
}

func (m *model) isStringObjectType() bool {
	return m.objectType == stringObjectType
}

func (m *model) isMapElementType() bool {
	return m.elementType == mapElementType
}

func (m *model) isSliceOfMapsElementType() bool {
	return m.elementType == sliceOfMapsElementType
}

func (m *model) isSliceOfStringsElementType() bool {
	return m.elementType == sliceOfStringsElementType
}

func (m *model) hasDataField() bool {
	return m.objectType == secretObjectType || m.objectType == configMapObjectType
}

var keys = []modelKey{pathModelKey, objectTypeModelKey, elementTypeModelKey, sourceKeyModelKey, sourceValueModelKey}

func newModel(annotationValue string) (*model, error) {

	raw := make(map[modelKey]string)

	for _, kv := range strings.Split(annotationValue, ",") {
		for i := range keys {
			k := keys[i]
			prefix := fmt.Sprintf("%v=", k)
			if strings.HasPrefix(kv, prefix) {
				raw[k] = kv[len(prefix):]
			}
		}
	}

	// assert PathModelKey is present
	path, found := raw[pathModelKey]
	if !found {
		if len(raw) == 0 {
			return &model{
				value: annotationValue,
			}, nil
		}
		return nil, fmt.Errorf("path not found: %q", annotationValue)
	}
	if n := strings.Count(path, "{"); n == 0 || n != strings.Count(path, "}") {
		return nil, fmt.Errorf("path has invalid syntax: %q", path)
	}

	// ensure ObjectTypeModelKey has a default value
	var objType objectType
	if rawObjectType, found := raw[objectTypeModelKey]; !found {
		objType = stringObjectType
	} else {
		// in the case the key is present but the value isn't (for example, "objectType=,") the
		// default string object type should be set
		if objType = objectType(rawObjectType); objType == emptyObjectType {
			objType = stringObjectType
		}
	}

	// ensure sourceKey has a default value
	sourceKey, found := raw[sourceKeyModelKey]
	if !found {
		sourceKey = ""
	}

	sourceValue, found := raw[sourceValueModelKey]
	if !found {
		sourceValue, found = raw[sourceKeyModelKey]
		if !found {
			sourceValue = ""
		}
	}

	// hasData indicates the configured or inferred objectType is either a Secret or ConfigMap
	hasData := objType == secretObjectType || objType == configMapObjectType
	// hasSourceKey indicates a value for sourceKey has been informed

	var eltType elementType
	if rawEltType, found := raw[elementTypeModelKey]; found {
		// the input string contains an elementType configuration, use it
		eltType = elementType(rawEltType)
	} else if hasData {
		// the input doesn't contain an elementType configuration, does contain a sourceKey
		// configuration, and is either a Secret or ConfigMap
		eltType = mapElementType
	} else {
		// elementType configuration hasn't been informed and there's no extra hints, assume it is a
		// string element
		eltType = stringElementType
	}

	// ensure an error is returned if not all required information is available for sliceOfMaps
	// element type
	if eltType == sliceOfMapsElementType && (len(sourceValue) == 0 || len(sourceKey) == 0) {
		return nil, errors.New("sliceOfMaps elementType requires sourceKey and sourceValue to be present")
	}

	return &model{
		path:        path,
		elementType: eltType,
		objectType:  objType,
		sourceValue: sourceValue,
		sourceKey:   sourceKey,
	}, nil
}
