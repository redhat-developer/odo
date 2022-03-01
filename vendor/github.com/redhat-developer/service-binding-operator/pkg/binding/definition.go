package binding

import (
	"encoding/base64"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type objectType string

type elementType string

const (
	// configMapObjectType indicates the path contains a name for a ConfigMap containing the binding
	// data.
	configMapObjectType objectType = "ConfigMap"
	// secretObjectType indicates the path contains a name for a Secret containing the binding data.
	secretObjectType objectType = "Secret"
	// stringObjectType indicates the path contains a value string.
	stringObjectType objectType = "string"
	// emptyObjectType is used as default value when the objectType key is present in the string
	// provided by the user but no value has been provided; can be used by the user to force the
	// system to use the default objectType.
	emptyObjectType objectType = ""

	// mapElementType indicates the value found at path is a map[string]interface{}.
	mapElementType elementType = "map"
	// sliceOfMapsElementType indicates the value found at path is a slice of maps.
	sliceOfMapsElementType elementType = "sliceOfMaps"
	// sliceOfStringsElementType indicates the value found at path is a slice of strings.
	sliceOfStringsElementType elementType = "sliceOfStrings"
	// stringElementType indicates the value found at path is a string.
	stringElementType elementType = "string"
)

//go:generate mockgen -destination=mocks/mocks.go -package=mocks . Definition,Value

type Definition interface {
	Apply(u *unstructured.Unstructured) (Value, error)
	GetPath() string
}

type DefinitionBuilder interface {
	Build() (Definition, error)
}

type definition struct {
	path string
}

func (d *definition) GetPath() string {
	return d.path
}

type stringDefinition struct {
	outputName string
	value      string
	definition
}

var _ Definition = (*stringDefinition)(nil)

func (d *stringDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	if d.outputName == "" {
		return nil, fmt.Errorf("cannot use generic service.binding annotation for string elements, need to specify binding key like service.binding/foo")
	}
	if d.value != "" {
		return &value{
			v: map[string]interface{}{
				d.outputName: d.value,
			},
		}, nil
	}
	val, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}

	if len(val) != 1 {
		return nil, fmt.Errorf("only one value should be returned for %v but we got %v", d.path, val)
	}

	m := map[string]interface{}{
		d.outputName: val[0].Interface(),
	}

	return &value{v: m}, nil
}

type stringFromDataFieldDefinition struct {
	secretConfigMapReader *secretConfigMapReader
	objectType            objectType
	outputName            string
	definition
	sourceKey string
}

var _ Definition = (*stringFromDataFieldDefinition)(nil)

func (d *stringFromDataFieldDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	if d.secretConfigMapReader == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	res, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, fmt.Errorf("only one value should be returned for %v but we got %v", d.path, res)
	}
	resourceName := res[0].String()

	var otherObj *unstructured.Unstructured
	if d.objectType == secretObjectType {
		otherObj, err = d.secretConfigMapReader.secretReader(u.GetNamespace(), resourceName)
	} else if d.objectType == configMapObjectType {
		otherObj, err = d.secretConfigMapReader.configMapReader(u.GetNamespace(), resourceName)
	}

	if err != nil {
		return nil, err
	}

	val, ok, err := unstructured.NestedString(otherObj.Object, "data", d.sourceKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}
	if d.objectType == secretObjectType {
		n, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return nil, err
		}
		val = string(n)
	}
	v := map[string]interface{}{
		"": val,
	}
	return &value{v: v}, nil
}

type mapFromDataFieldDefinition struct {
	secretConfigMapReader *secretConfigMapReader
	objectType            objectType
	outputName            string
	sourceValue           string
	definition
}

var _ Definition = (*mapFromDataFieldDefinition)(nil)

func (d *mapFromDataFieldDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	if d.secretConfigMapReader == nil {
		return nil, errors.New("kubeClient required for this functionality")
	}

	res, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("only one value should be returned for %v but we got %v", d.path, res)
	}
	resourceName := fmt.Sprintf("%v", res[0].Interface())

	var otherObj *unstructured.Unstructured
	if d.objectType == secretObjectType {
		otherObj, err = d.secretConfigMapReader.secretReader(u.GetNamespace(), resourceName)
	} else if d.objectType == configMapObjectType {
		otherObj, err = d.secretConfigMapReader.configMapReader(u.GetNamespace(), resourceName)
	}

	if err != nil {
		return nil, err
	}

	val, ok, err := unstructured.NestedStringMap(otherObj.Object, "data")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not found")
	}

	outputVal := make(map[string]string)

	for k, v := range val {
		if len(d.sourceValue) > 0 && k != d.sourceValue {
			continue
		}
		var n string
		if d.objectType == secretObjectType {
			b, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return nil, err
			}
			n = string(b)
		} else {
			n = v
		}
		if len(d.sourceValue) > 0 && len(d.outputName) > 0 {
			outputVal[d.outputName] = string(n)
		} else {
			outputVal[k] = string(n)
		}
	}

	return &value{v: outputVal}, nil
}

type stringOfMapDefinition struct {
	outputName string
	definition
}

var _ Definition = (*stringOfMapDefinition)(nil)

func (d *stringOfMapDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}
	if len(val) != 1 {
		return nil, fmt.Errorf("only one value should be returned for %v but we got %v", d.path, val)
	}

	valMap, ok := val[0].Interface().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("returned value for %v should be map, but we got %v", d.path, val[0].Interface())
	}

	outputName := d.outputName

	if outputName != "" {
		return &value{v: map[string]interface{}{
			outputName: valMap,
		}}, nil
	}
	return &value{v: valMap}, nil
}

type sliceOfMapsFromPathDefinition struct {
	outputName string
	definition
	sourceKey   string
	sourceValue string
}

var _ Definition = (*sliceOfMapsFromPathDefinition)(nil)

func (d *sliceOfMapsFromPathDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, vv := range val {
		for k, v := range collectSourceValuesWithKey(vv.Interface(), d.sourceValue, d.sourceKey) {
			res[k] = v
		}
	}

	if d.outputName == "" {
		return &value{v: res}, nil
	}
	return &value{v: map[string]interface{}{d.outputName: res}}, nil
}

func collectSourceValuesWithKey(i interface{}, sourceValue string, sourceKey string) map[string]interface{} {
	res := make(map[string]interface{})
	switch v := i.(type) {
	case map[string]interface{}:
		key := v[sourceKey]
		res[fmt.Sprintf("%v", key)] = v[sourceValue]
	case []interface{}:
		for _, item := range v {
			for k, v := range collectSourceValuesWithKey(item, sourceValue, sourceKey) {
				res[k] = v
			}
		}
	}
	return res
}

type sliceOfStringsFromPathDefinition struct {
	outputName string
	definition
	sourceValue string
}

var _ Definition = (*sliceOfStringsFromPathDefinition)(nil)

func (d *sliceOfStringsFromPathDefinition) Apply(u *unstructured.Unstructured) (Value, error) {
	val, err := getValuesByJSONPath(u.Object, d.path)
	if err != nil {
		return nil, err
	}
	var res []interface{}
	for _, e := range val {
		res = append(res, collectSourceValues(e.Interface(), d.sourceValue)...)
	}

	return &value{v: map[string]interface{}{d.outputName: res}}, nil
}

func collectSourceValues(i interface{}, sourceValue string) []interface{} {
	var res []interface{}
	switch v := i.(type) {
	case map[string]interface{}:
		if sourceValue != "" {
			res = append(res, v[sourceValue])
		}
	case []interface{}:
		for _, item := range v {
			res = append(res, collectSourceValues(item, sourceValue)...)
		}
	case string:
		if sourceValue == "" {
			res = append(res, v)
		}
	}
	return res
}
