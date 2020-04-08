package data

import (
	"fmt"
	"reflect"
)

// String converts supportedApiVersion type to string type
func (s supportedApiVersion) String() string {
	return string(s)
}

// NewDevfileData returns relevant devfile struct for the provided API version
func NewDevfileData(version string) (obj DevfileData, err error) {

	// Fetch devfile struct type from map
	devfileType, ok := apiVersionToDevfileStruct[supportedApiVersion(version)]
	if !ok {
		errMsg := fmt.Sprintf("devfile type not present for apiVersion '%s'", version)
		return obj, fmt.Errorf(errMsg)
	}

	return reflect.New(devfileType).Interface().(DevfileData), nil
}

// GetDevfileJSONSchema returns the devfile JSON schema of the supported apiVersion
func GetDevfileJSONSchema(version string) (string, error) {

	// Fetch json schema from the devfileApiVersionToJSONSchema map
	schema, ok := devfileApiVersionToJSONSchema[supportedApiVersion(version)]
	if !ok {
		return "", fmt.Errorf("unable to find schema for apiVersion '%s'", version)
	}

	// Successful
	return schema, nil
}

// IsApiVersionSupported returns true if the API version is supported in odo
func IsApiVersionSupported(version string) bool {
	for _, v := range supportedApiVersionsList {
		if v.String() == version {
			return true
		}
	}
	return false
}
