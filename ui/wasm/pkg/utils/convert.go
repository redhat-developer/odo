package utils

func StringArrayToInterfaceArray(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, str := range strings {
		result[i] = str
	}
	return result
}
