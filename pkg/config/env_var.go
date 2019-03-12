package config

import (
	"errors"
	"strings"

	"github.com/redhat-developer/odo/pkg/util"
)

// EnvVar represents an enviroment variable
type EnvVar struct {
	Name  string `yaml:"Name"`
	Value string `yaml:"Value"`
}

// EnvVarList represents a list of environment variables
type EnvVarList []*EnvVar

// NewEnvVarFromString takes a string of format "name=value" and returns an Env
// variable struct
func NewEnvVarFromString(envStr string) (*EnvVar, error) {
	envList := strings.Split(envStr, "=")
	// if there is not = in the string
	if len(envList) < 2 {
		return nil, errors.New("invalid environment variable format")
	}

	return &EnvVar{
		Name:  strings.TrimSpace(envList[0]),
		Value: strings.TrimSpace(envList[1]),
	}, nil
}

// NewEnvVarListFromSlice takes multiple env variables with format
// "name=value" and returns an EnvVarList
func NewEnvVarListFromSlice(envList []string) (EnvVarList, error) {
	var envVarList EnvVarList
	for _, envStr := range envList {
		envVar, err := NewEnvVarFromString(envStr)
		if err != nil {
			return nil, err
		}
		envVarList = append(envVarList, envVar)
	}

	return envVarList, nil

}

// MergeEnvVarList merges the two provided envVarList
func MergeEnvVarList(evl EnvVarList, otherEvl EnvVarList) EnvVarList {
	for _, envVar := range otherEvl {
		evl = append(evl, envVar)
	}
	return evl
}

// RemoveEnvVarsFromList removes the env variables based on the keys provided
// and returns a new EnvVarList
func RemoveEnvVarsFromList(envVarList EnvVarList, keys []string) EnvVarList {
	newEnvVarList := EnvVarList{}
	for _, envVar := range envVarList {
		// if the env is in the keys we skip it
		if util.In(keys, envVar.Name) {
			continue
		}

		newEnvVarList = append(newEnvVarList, envVar)
	}
	return newEnvVarList
}
