package util

func GetStringOrEmpty(ptr *string) string {
	return GetStringOrDefault(ptr, "")
}

func GetStringOrDefault(ptr *string, defaultValue string) string {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func GetIntOrDefault(ptr *int, defaultValue int) int {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func GetBoolOrDefault(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
