package data

import (
	"reflect"

	v100 "github.com/openshift/odo/pkg/devfile/parser/data/1.0.0"
	v200 "github.com/openshift/odo/pkg/devfile/parser/data/2.0.0"
)

// SupportedApiVersions stores the supported devfile API versions
type supportedApiVersion string

// Supported devfile API versions in odo
const (
	APIVersion100 supportedApiVersion = "1.0.0"
	APIVersion200 supportedApiVersion = "2.0.0"
)

// ------------- Init functions ------------- //

// apiVersionToDevfileStruct maps supported devfile API versions to their corresponding devfile structs
var apiVersionToDevfileStruct map[supportedApiVersion]reflect.Type

// Initializes a map of supported devfile api versions and devfile structs
func init() {
	apiVersionToDevfileStruct = make(map[supportedApiVersion]reflect.Type)
	apiVersionToDevfileStruct[APIVersion100] = reflect.TypeOf(v100.Devfile100{})
	apiVersionToDevfileStruct[APIVersion200] = reflect.TypeOf(v200.Devfile200{})

}

// Map to store mappings between supported devfile API versions and respective devfile JSON schemas
var devfileApiVersionToJSONSchema map[supportedApiVersion]string

// init initializes a map of supported devfile apiVersions with it's respective devfile JSON schema
func init() {
	devfileApiVersionToJSONSchema = make(map[supportedApiVersion]string)
	devfileApiVersionToJSONSchema[APIVersion100] = v100.JsonSchema100
	devfileApiVersionToJSONSchema[APIVersion200] = v200.JsonSchema200

}
