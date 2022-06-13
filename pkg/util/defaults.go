package util

import (
	"time"
)

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

func GetTimeDefault(ptr *time.Duration, defaultValue time.Duration) time.Duration {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
