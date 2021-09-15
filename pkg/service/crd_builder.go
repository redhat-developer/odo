package service

import (
	"strings"
)

// BuildCRDFromParams iterates over the parameter maps provided by the user and builds the CRD
func BuildCRDFromParams(paramMap map[string]string, group, version, kind string) (map[string]interface{}, error) {
	spec := map[string]interface{}{}
	for k, v := range paramMap {
		addParam(spec, k, v)
	}

	result := map[string]interface{}{}
	result["apiVersion"] = group + "/" + version
	result["kind"] = kind
	result["metadata"] = make(map[string]interface{})
	result["spec"] = spec
	return result, nil
}

// TODO check errors
func addParam(m map[string]interface{}, key string, value string) {
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		_, ok := m[parts[0]]
		if !ok {
			m[parts[0]] = map[string]interface{}{}
		}
		addParam(m[parts[0]].(map[string]interface{}), parts[1], value)
	} else {
		m[key] = value
	}
}
