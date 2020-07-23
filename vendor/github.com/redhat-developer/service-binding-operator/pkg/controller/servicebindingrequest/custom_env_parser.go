package servicebindingrequest

import (
	"bytes"
	"encoding/json"
	"text/template"

	corev1 "k8s.io/api/core/v1"
)

// CustomEnvParser is responsible to interpolate a given EnvVar containing templates.
type CustomEnvParser struct {
	EnvMap []corev1.EnvVar
	Cache  map[string]interface{}
}

// NewCustomEnvParser returns a new CustomEnvParser.
func NewCustomEnvParser(envMap []corev1.EnvVar, cache map[string]interface{}) *CustomEnvParser {
	return &CustomEnvParser{
		EnvMap: envMap,
		Cache:  cache,
	}
}

// Parse interpolates and caches the templates in EnvMap.
func (c *CustomEnvParser) Parse() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	for _, v := range c.EnvMap {
		tmpl, err := template.New("set").Funcs(template.FuncMap{"json": marshalToJSON}).Parse(v.Value)
		if err != nil {
			return data, err
		}

		// evaluating template and storing value in a buffer
		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, c.Cache)
		if err != nil {
			return data, err
		}

		// saving buffer in cache
		data[v.Name] = buf.String()
	}
	return data, nil
}
func marshalToJSON(m interface{}) (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
