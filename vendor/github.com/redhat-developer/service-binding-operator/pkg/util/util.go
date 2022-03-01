package util

func MergeMaps(dest map[string]string, src map[string]string) map[string]string {
	if dest == nil {
		return src
	}
	if src == nil {
		return dest
	}
	for k, v := range src {
		dest[k] = v
	}
	return dest
}
