package exports

// result returns the value and error in a format acceptable for JS
func result(value interface{}, err error) map[string]interface{} {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	return map[string]interface{}{
		"value": value,
		"err":   errStr,
	}
}
