package service

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// BuildCRDFromParams iterates over the parameter maps provided by the user and builds the CRD
func BuildCRDFromParams(paramMap map[string]string, group, version, kind string) (map[string]interface{}, error) {
	spec := map[string]interface{}{}
	for k, v := range paramMap {
		err := addParam(spec, k, v)
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

func addParam(m map[string]interface{}, key string, value string) error {
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		_, found := m[parts[0]]
		if !found {
			m[parts[0]] = map[string]interface{}{}
		}
		submap, ok := m[parts[0]].(map[string]interface{})
		if !ok {
			return errors.New("already defined")
		}
		err := addParam(submap, parts[1], value)
		if err != nil {
			return err
		}
	} else {
		if _, found := m[key]; found {
			return errors.New("already defined")
		}
		// TODO(feloy) convert based on declared type in schema
		m[key] = convertType(value)
	}
	return nil
}

func convertType(value string) interface{} {
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
	// if there were errors for everything else we return the string value
	return value
}
