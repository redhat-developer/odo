package data

import (
	"fmt"
	"reflect"
	"strings"

	"k8s.io/klog"
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
		var supportedVersions []string
		for version := range devfileApiVersionToJSONSchema {
			supportedVersions = append(supportedVersions, string(version))
		}
		return "", fmt.Errorf("unable to find schema for version %q. The parser supports devfile schema for version %s", version, strings.Join(supportedVersions, ", "))
	}
	klog.V(4).Infof("devfile apiVersion '%s' is supported", version)

	// Successful
	return schema, nil
}

// IsApiVersionSupported returns true if the API version is supported
func IsApiVersionSupported(version string) bool {
	return apiVersionToDevfileStruct[supportedApiVersion(version)] != nil
}
