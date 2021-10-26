package config

import (
	"errors"
	"fmt"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/v2/pkg/util"
)

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `yaml:"Name"`
	Value string `yaml:"Value"`
}

// EnvVarList represents a list of environment variables
type EnvVarList []EnvVar

// ToStringSlice converts the EnvVarList into a slice of env var of kind
// "key=value"
func (evl EnvVarList) ToStringSlice() []string {
	var envSlice []string
	for _, envVar := range evl {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", envVar.Name, envVar.Value))
	}

	return envSlice
}

// ToDevfileEnv converts the EnvVarList to the list of Envs supported by devfile
func (evl EnvVarList) ToDevfileEnv() []devfilev1.EnvVar {
	var envList []devfilev1.EnvVar
	for _, ev := range evl {
		envList = append(envList, devfilev1.EnvVar{
			Name:  ev.Name,
			Value: ev.Value,
		})
	}
	return envList
}

// Merge merges the other EnvVarlist with keeping last value for duplicate EnvVars
// and returns a new EnvVarList
func (evl EnvVarList) Merge(other EnvVarList) EnvVarList {

	var dedupNewEvl EnvVarList
	newEvl := append(evl, other...)
	uniqueMap := make(map[string]string)
	// last value will be kept in case of duplicate env vars
	for _, envVar := range newEvl {
		uniqueMap[envVar.Name] = envVar.Value
	}

	for key, value := range uniqueMap {
		dedupNewEvl = append(dedupNewEvl, EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return dedupNewEvl

}

// NewEnvVarFromString takes a string of format "name=value" and returns an Env
// variable struct
func NewEnvVarFromString(envStr string) (EnvVar, error) {
	envList := strings.SplitN(envStr, "=", 2)
	// if there is not = in the string
	if len(envList) < 2 {
		return EnvVar{}, errors.New("invalid environment variable format")
	}

	return EnvVar{
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

// NewEnvVarListFromDevfileEnv creates an EnvVarList from the list of envs present in a devfile.
func NewEnvVarListFromDevfileEnv(envList []devfilev1.EnvVar) EnvVarList {
	var envVarList EnvVarList
	for _, env := range envList {
		envVarList = append(envVarList, EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return envVarList
}

// RemoveEnvVarsFromList removes the env variables based on the keys provided
// and returns a new EnvVarList
func RemoveEnvVarsFromList(envVarList EnvVarList, keys []string) (EnvVarList, error) {
	// convert the envVarList map to an array to easily search for env var(s)
	// to remove from the component
	envVarListArray := []string{}
	for _, env := range envVarList {
		envVarListArray = append(envVarListArray, env.Name)
	}

	// now check if the environment variable(s) requested for removal exists in
	// the env vars set for the component by odo
	for _, key := range keys {
		if !util.In(envVarListArray, key) {
			return nil, fmt.Errorf("unable to find environment variable %s in the component", key)
		}
	}

	// finally, let's remove the environment variables(s) requested by the user
	newEnvVarList := EnvVarList{}
	for _, envVar := range envVarList {
		// if the env is in the keys we skip it
		if util.In(keys, envVar.Name) {
			continue
		}
		newEnvVarList = append(newEnvVarList, envVar)
	}
	return newEnvVarList, nil
}
