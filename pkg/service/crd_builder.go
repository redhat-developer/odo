package service

import (
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// BuildCRDFromParams iterates over the parameter maps provided by the user and builds the CRD
func BuildCRDFromParams(paramMap map[string]string, crd *spec.Schema, group, version, kind string) (map[string]interface{}, error) {
	spec := map[string]interface{}{}
	for k, v := range paramMap {
		err := addParam(spec, crd, k, v)
		if err != nil {
			return nil, err
		}
	}

	result := map[string]interface{}{}
	result["apiVersion"] = group + "/" + version
	result["kind"] = kind
	result["metadata"] = make(map[string]interface{})
	result["spec"] = spec
	return result, nil
}

func addParam(m map[string]interface{}, crd *spec.Schema, key string, value string) error {
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		property := parts[0]
		_, found := m[property]
		if !found {
			m[property] = map[string]interface{}{}
		}
		submap, ok := m[property].(map[string]interface{})
		if !ok {
			return errors.New("already defined")
		}
		var subCRD *spec.Schema
		if crd != nil {
			s := crd.Properties[property]
			subCRD = &s
		}
		err := addParam(submap, subCRD, parts[1], value)
		if err != nil {
			return err
		}
	} else {
		if _, found := m[key]; found {
			return errors.New("already defined")
		}

		var subCRD *spec.Schema
		if crd != nil {
			s := crd.Properties[key]
			subCRD = &s
		}
		m[key] = convertType(subCRD, value)
	}
	return nil
}

func convertType(crd *spec.Schema, value string) interface{} {
	if crd != nil {
		// do not use 'else' as the Schema can accept several types
		// the first matching type will be used
		if crd.Type.Contains("string") {
			return value
		}
		if crd.Type.Contains("integer") {
			intv, err := strconv.ParseInt(value, 10, 64)
			if err == nil {
				return int64(intv)
			}
		}
		if crd.Type.Contains("number") {
			floatv, err := strconv.ParseFloat(value, 64)
			if err == nil {
				return floatv
			}
		}
		if crd.Type.Contains("boolean") {
			boolv, err := strconv.ParseBool(value)
			if err == nil {
				return boolv
			}
		}
	} else {
		// no crd information available, guess the type depending on the value
		intv, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return int64(intv)
		}

		floatv, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return floatv
		}

		boolv, err := strconv.ParseBool(value)
		if err == nil {
			return boolv
		}
	}

	// as a last resort return the string value
	return value
}
